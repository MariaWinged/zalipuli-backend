package watersort

import (
	"testing"
	"time"
	"zalipuli/internal/storage/inmemory"

	"github.com/stretchr/testify/require"
)

func init() {
	FillConstants()
}

func TestNewWaterSortLevel(t *testing.T) {
	st := inmemory.New()
	level := NewWaterSortLevel(st)
	waitForGraphBuilt(level)
}

func TestWaterSortLevelHint(t *testing.T) {
	st := inmemory.New()
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

func TestWaterSortLevelJson(t *testing.T) {
	st := inmemory.New()
	level := NewWaterSortLevel(st)
	waitForGraphBuilt(level)

	json, err := level.ToJson()
	require.NoError(t, err)
	require.NotNil(t, json)

	newLevel := &Level{}
	err = newLevel.FromJson(json)
	require.NoError(t, err)
	require.Equal(t, level.startState, newLevel.startState)
	require.Equal(t, level.graph.minStepsCount, newLevel.graph.minStepsCount)

	for hash, pos := range level.graph.allPositions {
		newPos, ok := newLevel.graph.allPositions[hash]
		require.True(t, ok)

		require.Equal(t, pos.Hash(), newPos.Hash())
		nextPositions := make(map[string]bool)
		for _, next := range pos.nextPositions {
			nextPositions[next.Hash()] = true
		}
		for _, next := range newPos.nextPositions {
			require.True(t, nextPositions[next.Hash()])
		}

		require.Equal(t, len(pos.nextPositions), len(newPos.nextPositions))
		require.Equal(t, pos.isSuccessWay, newPos.isSuccessWay)
	}
}

func waitForGraphBuilt(level *Level) {
	for retry := 0; level.Status() != "ready" && retry < 100; retry++ {
		time.Sleep(500 * time.Millisecond)
	}

	if level.Status() != "ready" {
		panic("not ready")
	}
}
