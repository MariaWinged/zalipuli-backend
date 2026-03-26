package games

import (
	"errors"
	"zalipuli/pkg/api"
)

var (
	NotReadyErr      = errors.New("graph not ready")
	NotSuccessWayErr = errors.New("not success way")
)

type Level interface {
	Id() string
	Status() api.LevelResponseStatus
	GameName() api.GameName
	StartLevelState() (*api.LevelState, error)
	Hint(levelState api.LevelState) (*api.HintResponse_Hint, error)
	MinSteps() (*int, error)
}

type Graph interface {
	StartBuild(api.LevelState) error
	GameName() api.GameName
	GetMinSteps(api.LevelState) (int, error)
	GetRandomNextStep(api.LevelState) (*api.HintResponse_Hint, error)
	IsFinal(api.LevelState) bool
}
