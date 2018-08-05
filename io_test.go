package chunk_test

import (
	"testing"

	"github.com/matthewmueller/go-chunk"
	"github.com/stretchr/testify/assert"
)

func TestReadAll(t *testing.T) {
	buf := chunk.NewBuffer([]chunk.Chunk{"hi", "hello"})
	res, err := chunk.ReadAll(buf)
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, res, 2)
	assert.Equal(t, "hi", res[0].(string))
	assert.Equal(t, "hello", res[1].(string))
}
