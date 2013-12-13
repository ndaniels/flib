package main

import (
	"flag"
	"sync"

	"github.com/TuftsBCB/fragbag"
	"github.com/TuftsBCB/io/pdb"
	"github.com/TuftsBCB/seq"
	"github.com/TuftsBCB/structure"
	"github.com/TuftsBCB/tools/util"
)

var cmdMkSeqProfile = &command{
	name: "mk-seq-profile",
	positionalUsage: "struct-frag-lib out-frag-lib " +
		"pdb-chain-file [ pdb-chain-file ... ]",
	shortHelp: "create a new sequence fragment library with profiles",
	help: `
The mk-seq-profile command builds a sequence fragment library based on the
information from a structure fragment library and a set of PDB structures
to train on. The resulting library is a collection of fragments represented
as frequency profiles expressed as negative log odds scores. The null model
is built from the amino acid composition over all PDB chains given.

The algorithm for building a sequence fragment library is as follows:

  1. Initialize an empty frequency profile for each structure fragment in the
     library given.
  2. For every window in every PDB chain given, find the best matching
     structure fragment from the library provided.
  3. Add the corresponding region of sequence to that fragment's frequency
     profile. Also, update the null model for each amino acid seen.
  4. After all PDB chains are processed, build a profile in terms of negative
     log-odds using the null model constructed.
  5. Each profile corresponds to a fragment in the resulting sequence
     fragment library.

This process directly implies that the sequence fragment library produced will
have the same number of fragments and the same fragment size as the structure
fragment library given.
`,
	flags:    flag.NewFlagSet("mk-seq-profile", flag.ExitOnError),
	run:      mkSeqProfile,
	addFlags: func(c *command) { c.setOverwriteFlag() },
}

var (
	// There are two concurrent aspects going on here:
	// 1) processing entire PDB chains
	// 2) adding each part of each chain to a sequence fragment.
	// So we use two waitgroups: one for synchronizing on finishing
	// (1) and the other for synchronizing on finishing (2).
	wgPDBChains    = new(sync.WaitGroup)
	wgSeqFragments = new(sync.WaitGroup)
)

func mkSeqProfile(c *command) {
	c.assertLeastNArg(3)

	structLib := util.StructureLibrary(c.flags.Arg(0))
	outPath := c.flags.Arg(1)
	entries := c.flags.Args()[2:]

	util.AssertOverwritable(outPath, flagOverwrite)
	saveto := util.CreateFile(outPath)

	// Initialize a frequency and null profile for each structural fragment.
	var freqProfiles []*seq.FrequencyProfile
	var fpChans []chan seq.Sequence
	for i := 0; i < structLib.Size(); i++ {
		fp := seq.NewFrequencyProfile(structLib.FragmentSize())
		freqProfiles = append(freqProfiles, fp)
		fpChans = append(fpChans, make(chan seq.Sequence))
	}

	// Now spin up a goroutine for each fragment that is responsible for
	// adding a sequence slice to itself.
	nullChan, nullProfile := addToNull()
	for i := 0; i < structLib.Size(); i++ {
		addToProfile(fpChans[i], freqProfiles[i])
	}

	// Create a channel that sends the PDB entries given.
	entryChan := make(chan string)
	go func() {
		for _, fp := range entries {
			entryChan <- fp
		}
		close(entryChan)
	}()

	progress := util.NewProgress(len(entries))
	for i := 0; i < flagCpu; i++ {
		wgPDBChains.Add(1)
		go func() {
			for entryPath := range entryChan {
				_, chains, err := util.PDBOpen(entryPath)
				progress.JobDone(err)
				if err != nil {
					continue
				}

				for _, chain := range chains {
					structureToSequence(structLib, chain, nullChan, fpChans)
				}
			}
			wgPDBChains.Done()
		}()
	}
	wgPDBChains.Wait()
	progress.Close()

	// We've finishing reading all the PDB inputs. Now close the channels
	// and let the sequence fragments finish.
	close(nullChan)
	for i := 0; i < structLib.Size(); i++ {
		close(fpChans[i])
	}
	wgSeqFragments.Wait()

	// Finally, add the sequence fragments to a new sequence fragment
	// library and save.
	profs := make([]*seq.Profile, structLib.Size())
	for i := 0; i < structLib.Size(); i++ {
		profs[i] = freqProfiles[i].Profile(nullProfile)
	}
	lib, err := fragbag.NewSequenceProfile(structLib.Name(), profs)
	util.Assert(err)
	util.Assert(fragbag.Save(saveto, lib))
}

// structureToSequence uses structural fragments to categorize a segment
// of alpha-carbon atoms, and adds the corresponding residues to a
// corresponding sequence fragment.
func structureToSequence(
	lib fragbag.StructureLibrary,
	chain *pdb.Chain,
	nullChan chan seq.Sequence,
	seqChans []chan seq.Sequence,
) {
	sequence := chain.AsSequence()
	fragSize := lib.FragmentSize()

	// If the chain is shorter than the fragment size, we can do nothing
	// with it.
	if sequence.Len() < fragSize {
		util.Verbosef("Sequence '%s' is too short (length: %d)",
			sequence.Name, sequence.Len())
		return
	}

	// If we're accumulating a null model, add this sequence to it.
	if nullChan != nil {
		nullChan <- sequence
	}

	// This bit of trickery here is all about getting the call to
	// SequenceCaAtoms outside of the loop. In particular, it's a very
	// expensive call since it has to reconcile inconsistencies between
	// SEQRES and ATOM records in PDB files.
	limit := sequence.Len() - fragSize
	atoms := chain.SequenceCaAtoms()
	atomSlice := make([]structure.Coords, fragSize)
	noGaps := func(atoms []*structure.Coords) []structure.Coords {
		for i, atom := range atoms {
			if atom == nil {
				return nil
			}
			atomSlice[i] = *atom
		}
		return atomSlice
	}
	for start := 0; start <= limit; start++ {
		end := start + fragSize
		cas := noGaps(atoms[start:end])
		if cas == nil {
			// Nothing contiguous was found (a "disordered" residue perhaps).
			// So skip this part of the chain.
			continue
		}
		bestFrag := lib.BestStructureFragment(atomSlice)

		sliced := sequence.Slice(start, end)
		seqChans[bestFrag] <- sliced
	}
}

func addToProfile(sequences chan seq.Sequence, fp *seq.FrequencyProfile) {
	wgSeqFragments.Add(1)
	go func() {
		for s := range sequences {
			fp.Add(s)
		}
		wgSeqFragments.Done()
	}()
}

func addToNull() (chan seq.Sequence, *seq.FrequencyProfile) {
	nullChan := make(chan seq.Sequence, 100)
	nullProfile := seq.NewNullProfile()

	wgSeqFragments.Add(1)
	go func() {
		for s := range nullChan {
			for i := 0; i < s.Len(); i++ {
				nullProfile.Add(s.Slice(i, i+1))
			}
		}
		wgSeqFragments.Done()
	}()
	return nullChan, nullProfile
}
