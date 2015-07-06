package main

// This command has significant overlap with the `mk-seq-profile` command.

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/BurntSushi/ty/fun"
	"github.com/TuftsBCB/apps/hhsuite"
	"github.com/ndaniels/esfragbag"
	"github.com/TuftsBCB/io/msa"
	"github.com/TuftsBCB/seq"
	"github.com/TuftsBCB/tools/util"
)

var cmdMkSeqHMM = &command{
	name: "mk-seq-hmm",
	positionalUsage: "struct-frag-lib out-frag-lib " +
		"pdb-chain-file [ pdb-chain-file ... ]",
	shortHelp: "create a new sequence fragment library with profile HMMs",
	help: `
The mk-seq-hmm command builds a sequence fragment library based on the
information from a structure fragment library and a set of PDB structures
to train on. The resulting library is a collection of fragments represented
as profile HMMs. The null model used for each HMM is the background null
model used by hhsuite's 'hhmake' program.

The algorithm for building a sequence fragment library is as follows:

  1. Initialize an empty multiple sequence alignment for each structure
     fragment in the library given.
  2. For every window in every PDB chain given, find the best matching
     structure fragment from the library provided.
  3. Add the corresponding region of sequence to that fragment's MSA.
  4. After all PDB chains are processed, build a profile HMM for each
     fragment's MSA using the 'hhmake' command with pseudocount correction.
  5. Each profile HMM corresponds to a fragment in the resulting sequence
     fragment library.

This process directly implies that the sequence fragment library produced will
have the same number of fragments and the same fragment size as the structure
fragment library given.
`,
	flags:    flag.NewFlagSet("mk-seq-hmm", flag.ExitOnError),
	run:      mkSeqHMM,
	addFlags: func(c *command) { c.setOverwriteFlag() },
}

func mkSeqHMM(c *command) {
	c.assertLeastNArg(3)

	structLib := util.StructureLibrary(c.flags.Arg(0))
	outPath := c.flags.Arg(1)
	entries := c.flags.Args()[2:]

	util.AssertOverwritable(outPath, flagOverwrite)
	saveto := util.CreateFile(outPath)

	// Stores intermediate files produced by hhmake.
	tempDir, err := ioutil.TempDir("", "mk-seqlib-hmm")
	util.Assert(err, "Could not create temporary directory.")
	defer os.RemoveAll(tempDir)

	// Initialize a MSA for each structural fragment.
	var msas []seq.MSA
	var msaChans []chan seq.Sequence
	for i := 0; i < structLib.Size(); i++ {
		msa := seq.NewMSA()
		msa.SetLen(structLib.FragmentSize())
		msas = append(msas, msa)
		msaChans = append(msaChans, make(chan seq.Sequence))
	}

	// Now spin up a goroutine for each fragment that is responsible for
	// adding a sequence slice to itself.
	for i := 0; i < structLib.Size(); i++ {
		addToMSA(msaChans[i], &msas[i])
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
					structureToSequence(structLib, chain, nil, msaChans)
				}
			}
			wgPDBChains.Done()
		}()
	}
	wgPDBChains.Wait()
	progress.Close()

	// We've finishing reading all the PDB inputs. Now close the channels
	// and let the sequence fragments finish.
	for i := 0; i < structLib.Size(); i++ {
		close(msaChans[i])
	}
	wgSeqFragments.Wait()

	util.Verbosef("Building profile HMMs from MSAs...")

	// Finally, add the sequence fragments to a new sequence fragment
	// library and save.
	hmms := make([]*seq.HMM, structLib.Size())
	hhmake := func(i int) struct{} {
		fname := path.Join(tempDir, fmt.Sprintf("%d.fasta", i))
		f := util.CreateFile(fname)
		util.Assert(msa.WriteFasta(f, msas[i]))

		hhm, err := hhsuite.HHMakePseudo.Run(fname)
		util.Assert(err)
		hmms[i] = hhm.HMM
		return struct{}{} // my unifier sucks, i guess
	}
	fun.ParMap(hhmake, fun.Range(0, structLib.Size()))

	lib, err := fragbag.NewSequenceHMM(structLib.Name(), hmms)
	util.Assert(err)
	util.Assert(fragbag.Save(saveto, lib))
}

func addToMSA(sequences chan seq.Sequence, msa *seq.MSA) {
	wgSeqFragments.Add(1)
	go func() {
		for s := range sequences {
			// We don't use Add or AddFasta since both are
			// O(#sequences * #frag-length), which gets to be quite slow
			// in the presence of a lot of sequences.
			//
			// The key here is that we know that every sequence has the same
			// length and is in the same format, so we can add entries in a
			// straight forward manner.
			msa.Entries = append(msa.Entries, s)
		}
		wgSeqFragments.Done()
	}()
}
