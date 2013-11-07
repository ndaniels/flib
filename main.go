package main

import (
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"

	"github.com/TuftsBCB/tools/util"
)

var commands = []*command{
	cmdPairdist,
}

func usage() {
	log.Println("Usage: flib {command} [flags] [arguments]\n")
	log.Println("Use 'flib help {command}' for more details on {command}.\n")
	log.Println("A list of all available commands:\n")
	for _, c := range commands {
		log.Printf("    flib %s [flags] %s\n", c.name, c.positionalUsage)
	}
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
