package prefork

import (
	"os"

	"github.com/pkg/errors"
)

// Pipe is used for bi-directional communication between processes.
type Pipe struct {
	ReadFile  *os.File
	WriteFile *os.File
}

// NewPipes creates a pair of pipe objects. Data written to one of the pipes
// will be readable from the other.
func NewPipes() (*Pipe, *Pipe, error) {
	pr, cw, err := os.Pipe()
	if err != nil {
		return nil, nil, errors.Wrap(err, "error creating pipe")
	}
	cr, pw, err := os.Pipe()
	if err != nil {
		pr.Close()
		cw.Close()
		return nil, nil, errors.Wrap(err, "error creating pipe")
	}

	pp := &Pipe{ReadFile: pr, WriteFile: pw}
	cp := &Pipe{ReadFile: cr, WriteFile: cw}
	return pp, cp, nil
}

// Close the pipe.
func (p *Pipe) Close() error {
	rfErr := p.ReadFile.Close()
	wfErr := p.WriteFile.Close()
	if rfErr != nil {
		return errors.Wrap(rfErr, "error closing pipe reader")
	} else if wfErr != nil {
		return errors.Wrap(wfErr, "error closing pipe writer")
	}
	return nil
}

// Read data from the pipe.
func (p *Pipe) Read(b []byte) (n int, err error) {
	return p.ReadFile.Read(b)
}

// Write data to the pipe.
func (p *Pipe) Write(b []byte) (n int, err error) {
	return p.WriteFile.Write(b)
}
