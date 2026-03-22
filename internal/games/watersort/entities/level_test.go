package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewWaterSortLevel(t *testing.T) {
	level := NewWaterSortLevel()
	waitForGraphBuilt(level)
}

func TestWaterSortLevelHint(t *testing.T) {
	level := NewWaterSortLevel()
	waitForGraphBuilt(level)

	from, to := level.Hint(level.StartPosition())
	require.NotEqual(t, from, -1)
	require.NotEqual(t, to, -1)
}

func waitForGraphBuilt(level *WaterSortLevel) {
	FillConstants()

	for retry := 0; level.Status() != "ready" && retry < 10; retry++ {
		time.Sleep(500 * time.Millisecond)
	}

	if level.Status() != "ready" {
		panic("not ready")
	}
}
