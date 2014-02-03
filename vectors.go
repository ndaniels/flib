package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/TuftsBCB/tools/util"
)

var cmdVectors = &command{
	name:            "vectors",
	positionalUsage: "frag-lib bower-file [ bower-file ... ]",
	shortHelp:       "compute bag-of-words",
	help: `
The vectors command returns the Fragbag bag-of-words (vector) for each
bower file given. The format returned is a tab-delimited CSV file where
the first column is the name of the entry and each subsequent column is
the corresponding frequency for each corresponding fragment in the library
given.

Every row is guaranteed to be the same length. Namely, each row will have
N+1 entries, where N is the number of fragments given in the library.

Note that if a weighted fragment library is given, then the frequencies
will be reported as floating point values.

Bower files may either be PDB files or FASTA files.
`,
	flags: flag.NewFlagSet("vectors", flag.ExitOnError),
	run:   vectors,
}

func vectors(c *command) {
	c.assertLeastNArg(2)
	flib := util.Library(c.flags.Arg(0))
	bowPaths := c.flags.Args()[1:]

	tostrs := func(freqs []float32) []string {
		strs := make([]string, len(freqs))
		for i := range freqs {
			strs[i] = strconv.FormatFloat(float64(freqs[i]), 'f', -1, 32)
		}
		return strs
	}

	results := util.ProcessBowers(bowPaths, flib, flagPairdistModels,
		flagCpu, true)
	for r := range results {
		fmt.Printf("%s\t%s\n", r.Id, strings.Join(tostrs(r.Bow.Freqs), "\t"))
	}
}
