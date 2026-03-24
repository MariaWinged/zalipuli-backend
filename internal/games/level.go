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
