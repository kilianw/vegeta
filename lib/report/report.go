package report

import (
	"fmt"
	"github.com/kilianw/vegeta/lib"
	"io"
	"os"
	"os/signal"
	"time"
)

// Report generate metrics report
func Report(files []string, typ, output string, every time.Duration) error {
	if len(typ) < 4 {
		return fmt.Errorf("invalid report type: %s", typ)
	}

	dec, mc, err := vegeta.CreateDecoder(files)
	defer mc.Close()
	if err != nil {
		return err
	}
	create := true
	if output == "stdout" {
		create = false
	}

	out, err := vegeta.File(output, create)
	if err != nil {
		return err
	}
	if create {
		defer out.Close()
	}

	var (
		rep    vegeta.Reporter
		report vegeta.Report
	)

	switch typ[:4] {
	case "text":
		var m vegeta.Metrics
		rep, report = vegeta.NewTextReporter(&m), &m
	case "json":
		var m vegeta.Metrics
		rep, report = vegeta.NewJSONReporter(&m), &m
	case "hist":
		if len(typ) < 6 {
			return fmt.Errorf("bad buckets: '%s'", typ[4:])
		}
		var hist vegeta.Histogram
		if err := hist.Buckets.UnmarshalText([]byte(typ[4:])); err != nil {
			return err
		}
		rep, report = vegeta.NewHistogramReporter(&hist), &hist
	default:
		return fmt.Errorf("unknown report type: %q", typ)
	}

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)

	var ticks <-chan time.Time
	if every > 0 {
		ticker := time.NewTicker(every)
		defer ticker.Stop()
		ticks = ticker.C
	}

	rc, _ := report.(vegeta.Closer)
decode:
	for {
		select {
		case <-sigch:
			break decode
		case <-ticks:
			if err = writeReport(rep, rc, out); err != nil {
				return err
			}
		default:
			var r vegeta.Result
			if err = dec.Decode(&r); err != nil {
				if err == io.EOF {
					break decode
				}
				return err
			}

			report.Add(&r)
		}
	}

	return writeReport(rep, rc, out)
}

func writeReport(r vegeta.Reporter, rc vegeta.Closer, out io.Writer) error {
	if rc != nil {
		rc.Close()
	}
	return r.Report(out)
}

func clear(out io.Writer) error {
	if f, ok := out.(*os.File); ok && f == os.Stdout {
		return clearScreen()
	}
	return nil
}
