package chunk_test

import (
	"encoding/json"
	"testing"

	"github.com/matthewmueller/go-chunk"
	"github.com/stretchr/testify/assert"
)

func TestReader(t *testing.T) {
	r := chunk.NewBuffer([]chunk.Chunk{"hi"})
	v, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}

	s, ok := v.(string)
	if !ok {
		t.Fatal("not a string")
	}

	assert.Equal(t, s, "hi")
}

func TestReadOneByOne(t *testing.T) {
	r := chunk.NewBuffer([]chunk.Chunk{"hi", "hello"})

	// hi
	v, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}
	s, ok := v.(string)
	if !ok {
		t.Fatal("not a string")
	}
	assert.Equal(t, s, "hi")

	// hello
	v, err = r.Read()
	if err != nil {
		t.Fatal(err)
	}
	s, ok = v.(string)
	if !ok {
		t.Fatal("not a string")
	}
	assert.Equal(t, s, "hello")
}

func TestWrite(t *testing.T) {
	var b chunk.Buffer
	if err := b.Write("hi"); err != nil {
		t.Fatal(err)
	}

	// hi
	v, err := b.Read()
	if err != nil {
		t.Fatal(err)
	}
	s, ok := v.(string)
	if !ok {
		t.Fatal("not a string")
	}
	assert.Equal(t, s, "hi")
}

func TestConcurrentWrites(t *testing.T) {
	var b chunk.Buffer
	ch := make(chan struct{})

	go func() {
		if err := b.Write("hi"); err != nil {
			t.Fatal(err)
		}
		ch <- struct{}{}
	}()

	go func() {
		if err := b.Write("hello"); err != nil {
			t.Fatal(err)
		}
		ch <- struct{}{}
	}()

	<-ch
	<-ch

	// hi
	v, err := b.Read()
	if err != nil {
		t.Fatal(err)
	}
	s, ok := v.(string)
	if !ok {
		t.Fatal("not a string")
	}
	if s != "hi" && s != "hello" {
		t.Fatal("s is wrong")
	}

	// hello
	v, err = b.Read()
	if err != nil {
		t.Fatal(err)
	}
	s, ok = v.(string)
	if !ok {
		t.Fatal("not a string")
	}
	if s != "hi" && s != "hello" {
		t.Fatal("s is wrong")
	}
}

func TestConcurrentReadWrites(t *testing.T) {
	var b chunk.Buffer

	if err := b.Write("hi"); err != nil {
		t.Fatal(err)
	}

	ch := make(chan struct{})

	go func() {
		// hi
		v, err := b.Read()
		if err != nil {
			t.Fatal(err)
		}
		s, ok := v.(string)
		if !ok {
			t.Fatal("not a string")
		}
		if s != "hi" && s != "hello" {
			t.Fatal("s is wrong")
		}

		ch <- struct{}{}
	}()

	go func() {
		if err := b.Write("hello"); err != nil {
			t.Fatal(err)
		}
		ch <- struct{}{}
	}()

	<-ch
	<-ch

	// hello
	v, err := b.Read()
	if err != nil {
		t.Fatal(err)
	}
	s, ok := v.(string)
	if !ok {
		t.Fatal("not a string")
	}
	if s != "hi" && s != "hello" {
		t.Fatal("s is wrong")
	}
}

func TestMarshable(t *testing.T) {
	var a chunk.Chunk = struct {
		Name string `json:"name,omitempty"`
	}{"A"}

	var b chunk.Chunk = struct {
		Name string `json:"name,omitempty"`
	}{"B"}

	r := chunk.NewBuffer([]chunk.Chunk{a, b})

	ac, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}
	bc, err := r.Read()
	if err != nil {
		t.Fatal(err)
	}

	var chunk []chunk.Chunk
	chunk = append(chunk, ac, bc)

	buf, err := json.Marshal(chunk)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, `[{"name":"A"},{"name":"B"}]`, string(buf))
}
