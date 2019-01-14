package report

import (
	"io"
	"os"
	"os/signal"

	"github.com/kilianw/vegeta/lib"
	"github.com/kilianw/vegeta/lib/plot"
)

// PlotRun Generate time series plot
func PlotRun(files []string, threshold int, title, output string) error {
	dec, mc, err := vegeta.CreateDecoder(files)
	defer mc.Close()
	if err != nil {
		return err
	}

	out, err := vegeta.File(output, true)
	if err != nil {
		return err
	}
	defer out.Close()

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)

	p := plot.New(
		plot.Title(title),
		plot.Downsample(threshold),
		plot.Label(plot.ErrorLabeler),
	)

decode:
	for {
		select {
		case <-sigch:
			break decode
		default:
			var r vegeta.Result
			if err = dec.Decode(&r); err != nil {
				if err == io.EOF {
					break decode
				}
				return err
			}

			if err = p.Add(&r); err != nil {
				return err
			}
		}
	}

	p.Close()

	_, err = p.WriteTo(out)
	return err
}
