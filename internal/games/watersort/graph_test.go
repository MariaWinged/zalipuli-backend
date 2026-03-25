package watersort

import (
	"fmt"
	"log"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
)

func TestGraphPtr(t *testing.T) {
	g := &Graph{
		vialsCount: 10,
	}

	log.Printf("g: %s", fmt.Sprintf("%p", g))

	ptr := g.toPtr()

	log.Printf("ptr: %s", fmt.Sprintf("%d", ptr))

	newPtr := unsafe.Pointer(ptr)
	var newG = (*Graph)(newPtr)

	require.Equal(t, g, newG)
}
