package watersort

import (
	"testing"
	"time"
	"zalipuli/internal/storage"

	"github.com/stretchr/testify/require"
)

func TestNewWaterSortLevel(t *testing.T) {
	st := storage.New()
	level := NewWaterSortLevel(st)
	waitForGraphBuilt(level)
}

func TestWaterSortLevelHint(t *testing.T) {
	st := storage.New()
	level := NewWaterSortLevel(st)
	waitForGraphBuilt(level)

	state, err := level.StartLevelState()
	require.NoError(t, err)

	hint, err := level.Hint(*state)
	require.NoError(t, err)

	wsHint, err := hint.AsWaterSortHint()
	require.NoError(t, err)
	require.NotNil(t, wsHint)
}

func waitForGraphBuilt(level *Level) {
	FillConstants()

	for retry := 0; level.Status() != "ready" && retry < 10; retry++ {
		time.Sleep(500 * time.Millisecond)
	}

	if level.Status() != "ready" {
		panic("not ready")
	}
}
