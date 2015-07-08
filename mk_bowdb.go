package main

import (
	"flag"
	"github.com/ndaniels/esfragbag/bowdb"
	"github.com/ndaniels/tools/util"
)

var cmdMkBowDb = &command{
	name:            "mk-bowdb",
	positionalUsage: "bowdb-path frag-lib bower-file [ bower-file ... ]",
	shortHelp:       "create a database of proteins represented as BOWs",
	help: `
The mk-bowdb command creates a new database of proteins with each represented
as a bag-of-words in terms of the fragment library given.

The fragment library may be any kind; but the bower files provided must be able
to provide the type of BOW expected by the fragment library. For example, a
FASTA file can only be used with sequence fragment libraries, while a PDB file 
can be used with either structure or sequence fragment libraries.
`,
	flags: flag.NewFlagSet("mk-bowdb", flag.ExitOnError),
	run:   mkBowDb,
	addFlags: func(c *command) {
		c.setOverwriteFlag()
	},
}

func mkBowDb(c *command) {
	c.assertLeastNArg(3)

	dbPath := c.flags.Arg(0)
	flib := util.Library(c.flags.Arg(1))
	bowPaths := c.flags.Args()[2:]

	util.AssertOverwritable(dbPath, flagOverwrite)

	db, err := bowdb.Create(flib, dbPath)
	util.Assert(err)

	bows := util.ProcessBowers(bowPaths, flib, false, flagCpu, util.FlagQuiet)
	for b := range bows {
		db.Add(b)
	}
	util.Assert(db.Close())
}
