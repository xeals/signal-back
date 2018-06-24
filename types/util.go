package types

import (
	"fmt"
	"io"
	"os"
)

// MultiWriter is a convenience wrapper around an io.Writer to allow multiple
// consecutive (safe) writes.
type MultiWriter struct {
	io.Writer
	err error
}

// NewMultiWriter returns a new instance of a multi writer.
func NewMultiWriter(w io.Writer) *MultiWriter {
	return &MultiWriter{w, nil}
}

// W writes a slice of bytes to the underlying writer, or silently fails if
// there was a previous error.
func (w *MultiWriter) W(p []byte) {
	if w.err != nil {
		return
	}
	_, w.err = w.Write(p)
}

// Error returns the final error message of the writer.
func (w *MultiWriter) Error() error {
	return w.err
}

func rescue(v ...interface{}) {
	if r := recover(); r != nil {
		fmt.Fprintln(os.Stderr, "Panicked:", r)
		if v != nil {
			fmt.Fprintln(os.Stderr, v)
			os.Exit(2)
		}
	}
}
