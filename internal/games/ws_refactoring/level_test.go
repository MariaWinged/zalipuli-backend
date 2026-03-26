package ws

import (
	"testing"
	"time"
	"zalipuli/internal/storage/inmemory"

	"github.com/stretchr/testify/require"
)

func init() {
	storage := inmemory.New()
	WaterSortGraph = NewGraph(storage)
}

func TestNewWaterSortLevel(t *testing.T) {
	level, err := NewLevel()
	require.NoError(t, err)
	waitForGraphBuilt(level)
}

func waitForGraphBuilt(level *Level) {
	for retry := 0; level.Status() != "ready" && retry < 100; retry++ {
		time.Sleep(500 * time.Millisecond)
	}

	if level.Status() != "ready" {
		panic("not ready")
	}
}
