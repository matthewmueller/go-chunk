package chunk

import (
	"bytes"
	"io"
	"sync"
)

// Chunk interface
type Chunk interface{}

// NewBuffer from a slice of chunks
func NewBuffer(chunks []Chunk) *Buffer {
	bytes.NewBufferString("")
	return &Buffer{
		mu: sync.RWMutex{},
		s:  chunks,
	}
}

// Buffer struct
type Buffer struct {
	mu sync.RWMutex
	s  []Chunk
	i  int64 // current reading index
}

// Read one chunk
func (b *Buffer) Read() (c Chunk, err error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.i >= int64(len(b.s)) {
		return c, io.EOF
	}

	c = b.s[int(b.i)]
	b.i++

	return c, nil
}

func (b *Buffer) Write(c Chunk) (err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.s = append(b.s, c)

	return nil
}
