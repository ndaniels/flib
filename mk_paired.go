package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/ndaniels/esfragbag"
	"github.com/TuftsBCB/seq"
	"github.com/TuftsBCB/structure"
	"github.com/ndaniels/tools/util"
)

var cmdMkPaired = &command{
	name:            "mk-paired",
	positionalUsage: "in-frag-lib out-frag-lib",
	shortHelp:       "concatenates all fragment pairs",
	help: `
The mk-paired command generates a new fragment library from the one given
by concatenating all pairs of fragments and using each concatenation as
a fragment in the new library.

This command will produce a fragment library with N * (N-1) fragments each of
M*2 size, where N is the number of fragments in the input library and M is
the fragment size of the input library.

The 'in-frag-lib' is the source library with which to generate fragment
pairs. The file given is not modified. It must NOT be a weighted fragment
library. (Weights may be added to the paired fragment library with the
mk-weighted command.)

The 'out-frag-lib' is the path to write the new library with fragment pairs.
`,
	flags: flag.NewFlagSet("mk-paired", flag.ExitOnError),
	run:   mkPaired,
	addFlags: func(c *command) {
		c.setOverwriteFlag()
	},
}

func mkPaired(c *command) {
	c.assertNArg(2)

	in := util.Library(c.flags.Arg(0))
	outPath := c.flags.Arg(1)
	util.AssertOverwritable(outPath, flagOverwrite)

	if _, ok := in.(fragbag.WeightedLibrary); ok {
		util.Fatalf("%s is a weighted library (not allowed)", in.Name())
	}

	name := fmt.Sprintf("paired-%s", in.Name())
	if fragbag.IsStructure(in) {
		var pairs [][]structure.Coords
		lib := in.(fragbag.StructureLibrary)
		nfrags := lib.Size()
		for i := 0; i < nfrags; i++ {
			for j := 0; j < nfrags; j++ {
				if i == j {
					continue
				}
				f1, f2 := lib.Atoms(i), lib.Atoms(j)
				pairs = append(pairs, append(f1, f2...))
			}
		}
		pairLib, err := fragbag.NewStructureAtoms(name, pairs)
		util.Assert(err)
		fragbag.Save(util.CreateFile(outPath), pairLib)
	} else if strings.Contains(in.Tag(), "hmm") {
		var pairs []*seq.HMM
		lib := in.(fragbag.SequenceLibrary)
		nfrags := lib.Size()
		for i := 0; i < nfrags; i++ {
			for j := 0; j < nfrags; j++ {
				if i == j {
					continue
				}
				f1, f2 := lib.Fragment(i).(*seq.HMM), lib.Fragment(j).(*seq.HMM)
				pairs = append(pairs, seq.HMMCat(f1, f2))
			}
		}
		pairLib, err := fragbag.NewSequenceHMM(name, pairs)
		util.Assert(err)
		fragbag.Save(util.CreateFile(outPath), pairLib)
	} else if strings.Contains(in.Tag(), "profile") {
		util.Fatalf("Sequence profiles not implemented.")
	} else {
		util.Fatalf("Unrecognized fragment library: %s", in.Tag())
	}
}
