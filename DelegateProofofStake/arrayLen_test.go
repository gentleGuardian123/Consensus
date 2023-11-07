package DPoS

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArrayLen(t *testing.T) {
	arr := [4]string{}

	expected := 0

	assert.Equal(t, expected, len(arr))
}
