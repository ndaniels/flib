package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/ndaniels/esfragbag"
	"github.com/TuftsBCB/tools/util"
)

var cmdViewLib = &command{
	name:            "view-lib",
	positionalUsage: "frag-lib",
	shortHelp:       "view information about a fragment library",
	help: `
View information (size, fragment size, type, etc.) about a fragment library.
The information shown is fairly straight-forward. Although, note that the tag
shown is the full tag, including all sub-libraries. Sub-tags are delineated
by the "/" character.

This command may also be used to check if a fragment library is valid. In
particular, if this command fails, then the file given is not in a format
understood by the fragbag package installed.
`,
	flags:    flag.NewFlagSet("view-lib", flag.ExitOnError),
	run:      viewLib,
	addFlags: nil,
}

func viewLib(c *command) {
	c.assertNArg(1)

	lib := util.Library(c.flags.Arg(0))

	fmt.Printf("Name: %s\n", lib.Name())
	fmt.Printf("Tag: %s\n", strings.Join(libraryTag(lib), "/"))
	fmt.Printf("Size: %d\n", lib.Size())
	fmt.Printf("Fragment Size: %d\n", lib.FragmentSize())
	fmt.Printf("IsStructure: %v\n", fragbag.IsStructure(lib))
	fmt.Printf("IsSequence: %v\n", fragbag.IsSequence(lib))
}

func libraryTag(lib fragbag.Library) []string {
	if sub := lib.SubLibrary(); sub == nil {
		return []string{lib.Tag()}
	} else {
		return append([]string{lib.Tag()}, libraryTag(sub)...)
	}
}
