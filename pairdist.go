package main

import (
	"flag"
	"fmt"
	"math"

	"github.com/TuftsBCB/fragbag/bow"
	"github.com/TuftsBCB/tools/util"
)

var flagPairdistModels = false

var cmdPairdist = &command{
	name:            "pairdist",
	positionalUsage: "frag-lib bower-file [ bower-file ... ]",
	shortHelp:       "compute pairwise BOW distances",
	help: `
The pairdist command returns the cosine distance between every pair of
Fragbag frequency vectors produced by the given bower files.

Bower files may either be PDB files or FASTA files.
`,
	flags: flag.NewFlagSet("pairdist", flag.ExitOnError),
	run:   pairdist,
	addFlags: func(c *command) {
		c.flags.BoolVar(&flagPairdistModels, "models", flagPairdistModels,
			"When set, the models for each bower file given (if a PDB file)\n"+
				"will be used. Otherwise, the first the model from each\n"+
				"chain specified will be used.")
	},
}

func pairdist(c *command) {
	c.assertLeastNArg(2)
	flib := util.Library(c.flags.Arg(0))
	bowPaths := c.flags.Args()[1:]

	bows := make([]bow.Bowed, 0, 1000)
	results := util.ProcessBowers(bowPaths, flib, flagPairdistModels,
		flagCpu, util.FlagQuiet)
	for r := range results {
		bows = append(bows, r)
	}
	for i := 0; i < len(bows); i++ {
		b1 := bows[i]
		for j := i + 1; j < len(bows); j++ {
			b2 := bows[j]
			dist := math.Abs(b1.Bow.Cosine(b2.Bow))
			fmt.Printf("%s\t%s\t%0.4f\n", b1.Id, b2.Id, dist)
		}
	}
}
