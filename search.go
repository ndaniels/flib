package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/TuftsBCB/fragbag/bow"
	"github.com/TuftsBCB/fragbag/bowdb"
	"github.com/TuftsBCB/tools/util"
)

var (
	flagSearchOpts   = bowdb.SearchDefault
	flagSearchOutFmt = "plain"
	flagSearchSort   = "cosine"
	flagSearchDesc   = false
)

var cmdSearch = &command{
	name:            "search",
	positionalUsage: "bowdb-path bower-file [ bower-file ... ]",
	shortHelp:       "search a BOW database",
	help: `
The search command searches the given BOW database for entries closest to the
bower files given. The fragment library used to compute BOWs for the queries
is the one contained inside the given BOW database.
`,
	flags: flag.NewFlagSet("search", flag.ExitOnError),
	run:   search,
	addFlags: func(c *command) {
		c.flags.StringVar(&flagSearchOutFmt, "outfmt", flagSearchOutFmt,
			"The output format of the search results. Valid values are\n"+
				"'plain' and 'csv'.")
		c.flags.IntVar(&flagSearchOpts.Limit, "limit", flagSearchOpts.Limit,
			"The maximum number of search results to return.\n"+
				"To specify no limit, set this to -1.")
		c.flags.Float64Var(&flagSearchOpts.Min, "min", flagSearchOpts.Min,
			"All search results will have at least this distance with the "+
				"query.")
		c.flags.Float64Var(&flagSearchOpts.Max, "max", flagSearchOpts.Max,
			"All search results will have at most this distance with the "+
				"query.")

		c.flags.StringVar(&flagSearchSort, "sort", flagSearchSort,
			"The field to sort search results by.\n"+
				"Valid values are 'cosine' and 'euclid'.")
		c.flags.BoolVar(&flagSearchDesc, "desc", flagSearchDesc,
			"When set, results will be shown in descending order.")
	},
}

func search(c *command) {
	c.assertLeastNArg(2)

	// Some search options don't translate directly to command line parameters
	// specified by the flag package.
	if flagSearchDesc {
		flagSearchOpts.Order = bowdb.OrderDesc
	}
	switch flagSearchSort {
	case "cosine":
		flagSearchOpts.SortBy = bowdb.Cosine
	case "euclid":
		flagSearchOpts.SortBy = bowdb.Euclid
	default:
		util.Fatalf("Unknown sort field '%s'.", flagSearchSort)
	}

	db := util.OpenBowDB(c.flags.Arg(0))
	bowPaths := c.flags.Args()[1:]

	_, err := db.ReadAll()
	util.Assert(err, "Could not read BOW database entries")

	// always hide the progress bar here.
	bows := util.ProcessBowers(bowPaths, db.Lib, flagCpu, true)
	out, outDone := outputter()

	// launch goroutines to search queries in parallel
	wgSearch := new(sync.WaitGroup)
	for i := 0; i < flagCpu; i++ {
		wgSearch.Add(1)
		go func() {
			defer wgSearch.Done()

			for b := range bows {
				sr := db.Search(flagSearchOpts, b)
				out <- searchResult{b, sr}
			}
		}()
	}

	wgSearch.Wait()
	close(out)
	<-outDone
	util.Assert(db.Close())
}

type searchResult struct {
	query   bow.Bowed
	results []bowdb.SearchResult
}

func outputter() (chan searchResult, chan struct{}) {
	out := make(chan searchResult)
	done := make(chan struct{})
	go func() {
		if flagSearchOutFmt == "csv" {
			fmt.Printf("QueryID\tHitID\tCosine\tEuclid\n")
		}

		first := true
		for sr := range out {
			switch flagSearchOutFmt {
			case "plain":
				outputPlain(sr, first)
			case "csv":
				outputCsv(sr, first)
			default:
				util.Fatalf("Invalid output format '%s'.", flagSearchOutFmt)
			}
			first = false
		}
		done <- struct{}{}
	}()
	return out, done
}

func outputPlain(sr searchResult, first bool) {
	w := tabwriter.NewWriter(os.Stdout, 5, 0, 4, ' ', 0)
	wf := func(format string, v ...interface{}) {
		fmt.Fprintf(w, format, v...)
	}

	if !first {
		fmt.Println(strings.Repeat("-", 80))
	}
	header := fmt.Sprintf("%s (%d hits)", sr.query.Id, len(sr.results))

	fmt.Println(header)
	fmt.Println(strings.Repeat("-", len(header)))
	wf("Hit\tCosine\tEuclid\n")
	for _, result := range sr.results {
		wf("%s\t%0.4f\t%0.4f\n", result.Bowed.Id, result.Cosine, result.Euclid)
	}
	w.Flush()
}

func outputCsv(sr searchResult, first bool) {
	for _, result := range sr.results {
		fmt.Printf("%s\t%s\t%0.4f\t%0.4f\n",
			sr.query.Id, result.Bowed.Id, result.Cosine, result.Euclid)
	}
}
