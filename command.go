package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
)

var (
	flagCpu   = runtime.NumCPU()
	flagQuiet = false
)

func init() {
	log.SetFlags(0)
}

type command struct {
	name            string
	positionalUsage string
	help            string
	flags           *flag.FlagSet
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
	c.flags.IntVar(&flagCpu, "cpu", flagCpu,
		"Sets the maximum number of CPUs that can be executing simultaneously.")
	c.flags.BoolVar(&flagQuiet, "quiet", flagQuiet,
		"When set, progress information and other status messages will\n"+
			"not be printed to stderr.")
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
