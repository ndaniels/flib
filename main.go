package main

import (
	"log"
	"os"
	"runtime"
	"strings"
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
			if flagCpu < 1 {
				flagCpu = 1
			}
			runtime.GOMAXPROCS(flagCpu)
			if help {
				c.showHelp()
			} else {
				c.flags.Usage = c.showUsage
				c.flags.Parse(os.Args[2:])
				c.run(c)
				os.Exit(0)
			}
		}
	}
	log.Printf("Unknown command '%s'. Run 'flib help' for a list of "+
		"available commands.", cmd)
	os.Exit(1)
}
