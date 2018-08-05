package chunk

import (
	"context"
	"io"
)

// Reader interface
type Reader interface {
	Read() (c Chunk, err error)
}

// Writer interface
type Writer interface {
	Write(c Chunk) (err error)
}

// Closer interface
type Closer interface {
	Close() error
}

// ReadWriter interface
type ReadWriter interface {
	Reader
	Writer
}

// ReadCloser interface
type ReadCloser interface {
	Reader
	Closer
}

// WriteCloser interface
type WriteCloser interface {
	Writer
	Closer
}

// Processor interface helps us build pipelines
type Processor interface {
	Process(ctx context.Context, r Reader) (rc ReadCloser)
}

// ReadAll reads all the chunks in the internal buffer stopping at EOF
func ReadAll(r Reader) (s []Chunk, err error) {
	for {
		c, err := r.Read()
		if err != nil {
			if err != io.EOF {
				return s, err
			}
			break
		}
		s = append(s, c)
	}
	return s, err
}
