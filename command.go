package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/ndaniels/tools/util"
)

var (
	flagCpuProfile = ""
	flagCpu        = runtime.NumCPU()
	flagOverwrite  = false
)

func init() {
	log.SetFlags(0)
}

type command struct {
	name            string
	positionalUsage string
	shortHelp       string
	help            string
	flags           *flag.FlagSet
	addFlags        func(*command)
	run             func(*command)
}

func (c *command) showUsage() {
	log.Printf("Usage: flib %s [flags] %s\n", c.name, c.positionalUsage)
	c.showFlags()
	os.Exit(1)
}

func (c *command) showHelp() {
	log.Printf("Usage: flib %s [flags] %s\n\n", c.name, c.positionalUsage)
	log.Println(strings.TrimSpace(c.help))
	log.Printf("\nThe flags are:\n\n")
	c.showFlags()
	log.Println("")
	os.Exit(1)
}

func (c *command) showFlags() {
	c.flags.VisitAll(func(fl *flag.Flag) {
		var def string
		if len(fl.DefValue) > 0 {
			def = fmt.Sprintf(" (default: %s)", fl.DefValue)
		}
		usage := strings.Replace(fl.Usage, "\n", "\n    ", -1)
		log.Printf("-%s%s\n", fl.Name, def)
		log.Printf("    %s\n", usage)
	})
}

func (c *command) setCommonFlags() {
	c.flags.StringVar(&flagCpuProfile, "cpu-prof", flagCpuProfile,
		"When set, a CPU profile will be written to the file path provided.")
	c.flags.IntVar(&flagCpu, "cpu", flagCpu,
		"Sets the maximum number of CPUs that can be executing simultaneously.")
	c.flags.BoolVar(&util.FlagQuiet, "quiet", util.FlagQuiet,
		"When set, progress information and other status messages will\n"+
			"not be printed to stderr.")
}

func (c *command) setOverwriteFlag() {
	c.flags.BoolVar(&flagOverwrite, "overwrite", flagOverwrite,
		"When set, the output file will be overwritten if it already exists.")
}

func (c *command) assertNArg(n int) {
	if c.flags.NArg() != n {
		c.showUsage()
	}
}

func (c *command) assertLeastNArg(n int) {
	if c.flags.NArg() < n {
		c.showUsage()
	}
}
