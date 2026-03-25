package games

import (
	"encoding/json"
	"zalipuli/pkg/api"
)

type Level interface {
	Id() string
	Status() api.LevelResponseStatus
	GameName() api.GameName
	StartLevelState() (*api.LevelState, error)
	Hint(levelState api.LevelState) (*api.HintResponse_Hint, error)
	MinSteps() (*int, error)
	ToJson() (json.RawMessage, error)
	FromJson(json.RawMessage) error
}

type Position interface {
	Hash() string
	IsFinal() bool
	NextPositions() []string
	MinSteps() int
}

type Graph interface {
	Build() error
}

type PositionStorage interface {
	BuildGraph(api.LevelState) error
	Get(api.LevelState) (Position, error)
}
