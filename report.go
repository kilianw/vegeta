package main

import (
	"flag"
	"fmt"
	"github.com/kilianw/vegeta/lib/report"
	"os"
)

const reportUsage = `Usage: vegeta report [options] [<file>...]

Outputs a report of attack results.

Arguments:
  <file>  A file with vegeta attack results encoded with one of
          the supported encodings (gob | json | csv) [default: stdin]

Options:
  --type    Which report type to generate (text | json | hist[buckets]).
            [default: text]

  --every   Write the report to --output at every given interval (e.g 100ms)
            The default of 0 means the report will only be written after
            all results have been processed. [default: 0]

  --output  Output file [default: stdout]

Examples:
  echo "GET http://:80" | vegeta attack -rate=10/s > results.gob
  echo "GET http://:80" | vegeta attack -rate=100/s | vegeta encode > results.json
  vegeta report results.*
`

func reportCmd() command {
	fs := flag.NewFlagSet("vegeta report", flag.ExitOnError)
	typ := fs.String("type", "text", "Report type to generate [text, json, hist[buckets]]")
	every := fs.Duration("every", 0, "Report interval")
	output := fs.String("output", "stdout", "Output file")

	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, reportUsage)
	}

	return command{fs, func(args []string) error {
		fs.Parse(args)
		files := fs.Args()
		if len(files) == 0 {
			files = append(files, "stdin")
		}
		return report.Report(files, *typ, *output, *every)
	}}
}
