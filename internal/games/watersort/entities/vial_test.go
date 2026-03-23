package entities

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVialSegments(t *testing.T) {
	for i := 0; i < (1 << 16); i++ {
		vial := Vial(i)
		segments := vial.Segments()

		newVial := NewVial(segments)
		require.Equal(t, vial, newVial)
	}
}
