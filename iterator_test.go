package iterator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewIterator(t *testing.T) {
	i, err := NewIterator(
		func(i int, s string) (string, error) {
			return "", nil
		},
		func() int {
			return 0
		},
		func() string {
			return ""
		},
	)

	assert.NoError(t, err())
	assert.NotNil(t, i)
}

func TestLen(t *testing.T) {
	i, err := NewIterator(
		func(i int, s string) (string, error) {
			return "", nil
		},
		func() int {
			return 0
		},
		func() string {
			return ""
		},
	)

	assert.NoError(t, err())
	assert.NotNil(t, i)

	assert.Equal(t, 0, i.Len())
}
