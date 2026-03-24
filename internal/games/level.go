package games

import "zalipuli/pkg/api"

type Level interface {
	Id() string
	Status() api.LevelResponseStatus
	GameName() api.GameName
	StartLevelState() (*api.LevelState, error)
	Hint(levelState api.LevelState) (*api.HintResponse_Hint, error)
	MinSteps() (*int, error)
}
