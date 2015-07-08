package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	path "path/filepath"
	"strings"

	"github.com/ndaniels/esfragbag"
	"github.com/TuftsBCB/io/pdb"
	"github.com/TuftsBCB/structure"
	"github.com/ndaniels/tools/util"
)

var cmdMkStructure = &command{
	name:            "mk-structure",
	positionalUsage: "kolodny-brk-file out-frag-lib",
	shortHelp:       "create a new structure fragment library",
	help: `
The mk-structure command generates a structure fragment library from a brk
file produced by Rachel Kolodny's fragment library software. (???)
`,
	flags:    flag.NewFlagSet("mk-structure", flag.ExitOnError),
	run:      mkStructure,
	addFlags: func(c *command) { c.setOverwriteFlag() },
}

func mkStructure(c *command) {
	c.assertNArg(2)

	brkFile := c.flags.Arg(0)
	saveto := c.flags.Arg(1)

	util.AssertOverwritable(saveto, flagOverwrite)

	brkContents, err := ioutil.ReadAll(util.OpenFile(c.flags.Arg(0)))
	util.Assert(err)

	pdbFragments := bytes.Split(brkContents, []byte("TER"))
	fragments := make([][]structure.Coords, 0)
	for i, pdbFrag := range pdbFragments {
		pdbFrag = bytes.TrimSpace(pdbFrag)
		if len(pdbFrag) == 0 {
			continue
		}
		fragments = append(fragments, coords(i, pdbFrag))
	}

	libName := stripExt(path.Base(brkFile))
	lib, err := fragbag.NewStructureAtoms(libName, fragments)
	util.Assert(err)
	fragbag.Save(util.CreateFile(saveto), lib)
}

func coords(num int, atomRecords []byte) []structure.Coords {
	r := bytes.NewReader(atomRecords)
	name := fmt.Sprintf("fragment %d", num)

	entry, err := pdb.Read(r, name)
	util.Assert(err, "Fragment contents could not be read in PDB format")

	atoms := entry.OneChain().CaAtoms()
	if len(atoms) == 0 {
		util.Fatalf("Fragment %d has no ATOM coordinates.", num)
	}
	return atoms
}

func stripExt(s string) string {
	return strings.TrimSuffix(s, path.Ext(s))
}
