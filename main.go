package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"text/tabwriter"

	"github.com/ndaniels/tools/util"
)

var commands = []*command{
	cmdMkBowDb,
	cmdMkPaired,
	cmdMkSeqHMM,
	cmdMkSeqProfile,
	cmdMkStructure,
	cmdMkWeighted,
	cmdPairdist,
	cmdSearch,
	cmdVectors,
	cmdViewLib,
}

func usage() {
	log.Println("flib is a tool for creating and using fragment libraries.\n")
	log.Println("Usage:\n\n    flib {command} [flags] [arguments]\n")
	log.Println("Use 'flib help {command}' for more details on {command}.\n")
	log.Println("A list of all available commands:\n")

	tabw := tabwriter.NewWriter(os.Stderr, 0, 0, 4, ' ', 0)
	for _, c := range commands {
		fmt.Fprintf(tabw, "    %s\t%s\n", c.name, c.shortHelp)
	}
	tabw.Flush()
	log.Println("")
	os.Exit(1)
}

func main() {
	var cmd string
	var help bool
	if len(os.Args) < 2 {
		usage()
	} else if strings.TrimLeft(os.Args[1], "-") == "help" {
		if len(os.Args) < 3 {
			usage()
		} else {
			cmd = os.Args[2]
			help = true
		}
	} else {
		cmd = os.Args[1]
	}

	for _, c := range commands {
		if c.name == cmd {
			c.setCommonFlags()
			if c.addFlags != nil {
				c.addFlags(c)
			}
			if help {
				c.showHelp()
			} else {
				c.flags.Usage = c.showUsage
				c.flags.Parse(os.Args[2:])

				if flagCpu < 1 {
					flagCpu = 1
				}
				runtime.GOMAXPROCS(flagCpu)

				if len(flagCpuProfile) > 0 {
					f := util.CreateFile(flagCpuProfile)
					pprof.StartCPUProfile(f)
					defer f.Close()
					defer pprof.StopCPUProfile()
				}

				c.run(c)
				return
			}
		}
	}
	log.Printf("Unknown command '%s'. Run 'flib help' for a list of "+
		"available commands.", cmd)
	os.Exit(1)
}
