package main

import (
	"flag"
	"fmt"
	"math"

	"github.com/TuftsBCB/tools/util"
)

var cmdPairdist = &command{
	name: "pairdist",
	positionalUsage: "frag-lib bower-file [ bower-file ... ]",
	help: `
The pairdist command returns the cosine distance between every pair of
Fragbag frequency vectors produced by the given bower files.

Bower files may either be PDB files or FASTA files.
`,
	flags: flag.NewFlagSet("pairdist", flag.ExitOnError),
	run: pairdist,
}

func pairdist(c *command) {
	c.assertLeastNArg(2)
	flib := util.Library(c.flags.Arg(0))
	bowPaths := c.flags.Args()[1:]

	bows := make([]util.Bowered, 0, 1000)
	results := util.ProcessBowers(bowPaths, flib, flagCpu, flagQuiet)
	for r := range results {
		bows = append(bows, r)
	}
	for i := 0; i < len(bows); i++ {
		b1 := bows[i]
		for j := i+1; j < len(bows); j++ {
			b2 := bows[j]
			dist := math.Abs(b1.Bow.Cosine(b2.Bow))
			fmt.Printf("%s\t%s\t%0.4f\n", b1.Id, b2.Id, dist)
		}
	}
}
