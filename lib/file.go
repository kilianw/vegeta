package vegeta

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// File create and open file
func File(name string, create bool) (*os.File, error) {
	switch name {
	case "stdin":
		return os.Stdin, nil
	case "stdout":
		return os.Stdout, nil
	default:
		if create {
			return os.Create(name)
		}
		return os.Open(name)
	}
}

// CreateDecoder create report decoder
func CreateDecoder(files []string) (Decoder, io.Closer, error) {
	closer := make(multiCloser, 0, len(files))
	decs := make([]Decoder, 0, len(files))
	for _, f := range files {
		rc, err := File(f, false)
		if err != nil {
			return nil, closer, err
		}

		dec := DecoderFor(rc)
		if dec == nil {
			return nil, closer, fmt.Errorf("encode: can't detect encoding of %q", f)
		}

		decs = append(decs, dec)
		closer = append(closer, rc)
	}
	return NewRoundRobinDecoder(decs...), closer, nil
}

type multiCloser []io.Closer

func (mc multiCloser) Close() error {
	var errs []string
	for _, c := range mc {
		if err := c.Close(); err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}
