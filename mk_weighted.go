package main

import (
	"flag"
	"math"

	"github.com/TuftsBCB/fragbag"
	"github.com/TuftsBCB/tools/util"
)

var flagWeightedScheme = "tfidf"

var cmdMkWeighted = &command{
	name: "mk-weighted",
	positionalUsage: "train-frag-lib in-frag-lib out-frag-lib " +
		"bower-file [ bower-file ... ]",
	shortHelp: "add weights to an existing fragment library",
	help: `
The mk-weighted command trains weights on a fragment library and outputs a
new fragment library with those weights embedded in its representation.

The 'train-frag-lib' is the library to use to compute BOWs for the given
bower files. The bower files correspond to the document corpus to train
on. Namely, it should be representative of the space you are searching.
Bower files may either be PDB files or FASTA files.

The 'in-frag-lib' is the unweighted library with which to add weights.
The file given is not modified.

The 'out-frag-lib' is the path to write the new library. Namely, the new
library is the same as 'in-frag-lib', but with weight information.

All three fragment libraries must have the same parameters. Namely, they
must all have the same number of fragments and all must have the same
fragment size.

For example, to add weights to a sequence fragment library based on the
fragment frequencies from its corresponding structure fragment library,
you could use this command (with the PDB Select 25 acting as the
representative of the document space):

    pdbs-chains pdb25-file | xargs flib weighted structure.json sequence.json sequence-weighted.json
`,
	flags: flag.NewFlagSet("mk-weighted", flag.ExitOnError),
	run:   mkWeighted,
	addFlags: func(c *command) {
		c.setOverwriteFlag()
		c.flags.StringVar(&flagWeightedScheme, "scheme", flagWeightedScheme,
			"The weight scheme to use. Currently, only 'tfidf' is supported.")
	},
}

func mkWeighted(c *command) {
	c.assertLeastNArg(4)

	train := util.Library(c.flags.Arg(0))
	in := util.Library(c.flags.Arg(1))
	outPath := c.flags.Arg(2)
	bowPaths := c.flags.Args()[3:]

	util.AssertOverwritable(outPath, flagOverwrite)

	// The inverse-document-frequencies of each fragment in the "in" fragment
	// library.
	numFrags := in.Size()
	idfs := make([]float32, numFrags)

	// Compute the BOWs for each bower against the training fragment lib.
	bows := util.ProcessBowers(bowPaths, train, flagCpu, util.FlagQuiet)

	// Now tally the number of bowers that each fragment occurred in.
	totalBows := float32(0)
	for bow := range bows {
		totalBows += 1
		for fragi := 0; fragi < numFrags; fragi++ {
			if bow.Bow.Freqs[fragi] > 0 {
				idfs[fragi]++
			}
		}
	}

	// Compute the IDF using the frequencies against all the BOWs.
	for i := range idfs {
		idfs[i] = float32(math.Log(float64(totalBows / idfs[i])))
	}

	// Finally, wrap the given library as a weighted library and save it.
	wlib, err := fragbag.NewWeightedTfIdf(in, idfs)
	util.Assert(err)
	fragbag.Save(util.CreateFile(outPath), wlib)
}
