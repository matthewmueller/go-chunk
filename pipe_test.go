package chunk_test

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/matthewmueller/go-chunk"
	"github.com/stretchr/testify/assert"
)

func TestReadPipe(t *testing.T) {
	r, w := chunk.Pipe()
	go func() {
		w.Write("hi")
	}()

	c, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, c, "hi")
}

func TestWritePipe(t *testing.T) {
	r, w := chunk.Pipe()
	ch := make(chan struct{})

	go func() {
		c, err := r.Read()
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, c, "hi")
		close(ch)
	}()

	w.Write("hi")
	<-ch
}

func TestConcurrentWritePipe(t *testing.T) {
	r, w := chunk.Pipe()
	ch := make(chan struct{}, 3)
	go func() {
		go func() {
			w.Write("hi")
			ch <- struct{}{}
		}()
		go func() {
			w.Write("hi")
			ch <- struct{}{}
		}()
		go func() {
			w.Write("hi")
			ch <- struct{}{}
		}()
		w.Write("hi")
		<-ch
		<-ch
		<-ch
		w.Close()
	}()

	c, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, c, "hi")
	c, err = r.Read()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, c, "hi")
	c, err = r.Read()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, c, "hi")
	c, err = r.Read()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, c, "hi")

	c, err = r.Read()
	assert.Nil(t, c)
	assert.Equal(t, err, io.EOF)
}

func TestConcurrentWriteWithErrorPipe(t *testing.T) {
	r, w := chunk.Pipe()
	ch := make(chan struct{})
	go func() {
		if err := w.Write("hi"); err != nil {
			t.Fatal(err)
		}

		go func() {
			w.CloseWithError(errors.New("err"))
			ch <- struct{}{}
		}()

		<-ch

		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	c, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, c, "hi")

	c, err = r.Read()
	assert.Nil(t, c)
	assert.EqualError(t, err, "err")

	c, err = r.Read()
	assert.Nil(t, c)
	assert.EqualError(t, err, "err")
}

type Op interface {
	chunk.Chunk
	fmt.Stringer
}

var _ Op = (*Request)(nil)

type Request struct {
	name string
}

func (r *Request) String() string {
	return r.name
}

type Transformer interface {
	Transform(r chunk.Reader) chunk.ReadCloser
}

type A struct {
	name string
}

func (a *A) Transform(r chunk.Reader) chunk.ReadCloser {
	rc, wc := chunk.Pipe()

	go func() {
		cs, err := chunk.ReadAll(r)
		if err != nil {
			wc.CloseWithError(err)
			return
		}
		for _, c := range cs {
			if s, ok := c.(*Request); ok {
				s.name = a.name + s.name
				wc.Write(s)
			}
		}

		wc.Close()
	}()

	return rc
}

type B struct {
	name string
}

func (b *B) Transform(r chunk.Reader) chunk.ReadCloser {
	rc, wc := chunk.Pipe()

	go func() {
		cs, err := chunk.ReadAll(r)
		if err != nil {
			wc.CloseWithError(err)
			return
		}

		for _, c := range cs {
			if s, ok := c.(*Request); ok {
				s.name = b.name + s.name
				wc.Write(s)
			}
		}

		wc.Close()
	}()

	return rc
}

func TestInterface(t *testing.T) {
	a := &A{name: "a"}
	b := &B{name: "b"}

	r := chunk.NewBuffer([]chunk.Chunk{&Request{"m"}})
	t1 := a.Transform(r)
	t2 := b.Transform(t1)

	result, err := t2.Read()
	if err != nil {
		t.Fatal(err)
	}

	if err := t2.Close(); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "bam", result.(*Request).String())
}

func TestMultipleCloseOk(t *testing.T) {
	rc, wc := chunk.Pipe()

	if err := wc.Close(); err != nil {
		t.Fatal(err)
	}
	if err := wc.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := chunk.ReadAll(rc); err != nil {
		t.Fatal(err)
	}
}

func TestMultipleCloseWithErrorsFirst(t *testing.T) {
	rc, wc := chunk.Pipe()

	if err := wc.CloseWithError(errors.New("oh noz")); err != nil {
		t.Fatal(err)
	}
	if err := wc.CloseWithError(errors.New("oh nozz")); err != nil {
		t.Fatal(err)
	}
	if err := wc.Close(); err != nil {
		t.Fatal(err)
	}

	_, err := chunk.ReadAll(rc)
	assert.EqualError(t, err, "oh noz")
}

func TestMultipleCloseOkFirst(t *testing.T) {
	rc, wc := chunk.Pipe()

	if err := wc.Close(); err != nil {
		t.Fatal(err)
	}
	if err := wc.CloseWithError(errors.New("oh noz")); err != nil {
		t.Fatal(err)
	}
	if err := wc.CloseWithError(errors.New("oh nozz")); err != nil {
		t.Fatal(err)
	}

	if _, err := chunk.ReadAll(rc); err != nil {
		t.Fatal(err)
	}
}
