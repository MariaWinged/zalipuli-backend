package entities

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlaskSegments(t *testing.T) {
	for i := 0; i < (1 << 16); i++ {
		flask := Vial(i)
		segments := flask.Segments()

		newFlask := NewVial(segments)
		require.Equal(t, flask, newFlask)
	}
}
