package ws

import (
	"errors"
	"math/rand"
	"zalipuli/internal/games"
	"zalipuli/pkg/api"

	"github.com/google/uuid"
)

type Level struct {
	ID         string                  `json:"id"`
	Game       string                  `json:"game_name"`
	StartState api.WaterSortLevelState `json:"start_state"`
}

func NewLevel() (*Level, error) {
	// Случайно выбираем число цветов, заполняем колбы вперемешку, и сами колбы тоже мешаем
	colorsCount := rand.Intn(MaxColorsCount-MinColorsCount) + MinColorsCount
	allSegments := make([]int, colorsCount*VialHeight)
	for i := 0; i < len(allSegments); i++ {
		allSegments[i] = i/VialHeight + 1
	}
	rand.Shuffle(len(allSegments), func(i, j int) {
		allSegments[i], allSegments[j] = allSegments[j], allSegments[i]
	})

	vials := make([]Vial, colorsCount+2)
	for i := 0; i < colorsCount; i++ {
		vials[i] = NewVial(allSegments[i*VialHeight : i*VialHeight+VialHeight])
	}
	vials[colorsCount] = NewVial(nil)
	vials[colorsCount+1] = NewVial(nil)
	rand.Shuffle(len(vials), func(i, j int) {
		vials[i], vials[j] = vials[j], vials[i]
	})

	apiVials := convertToApiVials(vials)
	startState := api.WaterSortLevelState{Vials: apiVials, GameName: api.Watersort, ColorsCount: &colorsCount}

	var state api.LevelState
	err := state.FromWaterSortLevelState(startState)
	if err != nil {
		return nil, err
	}

	err = WaterSortGraph.StartBuild(state)
	if err != nil {
		return nil, err
	}

	level := &Level{
		ID:         uuid.New().String(),
		Game:       gameName,
		StartState: api.WaterSortLevelState{Vials: apiVials, GameName: api.Watersort, ColorsCount: &colorsCount},
	}

	return level, nil
}

func (l *Level) GameName() api.GameName {
	return api.Watersort
}

func (l *Level) Id() string {
	return l.ID
}

func (l *Level) Status() api.LevelResponseStatus {
	state := &api.LevelState{}
	err := state.FromWaterSortLevelState(l.StartState)
	if err != nil {
		return api.Incorrect
	}

	_, err = WaterSortGraph.GetMinSteps(*state)
	if err != nil {
		if errors.Is(err, games.NotReadyErr) {
			return api.Pending
		}
		return api.Incorrect
	}

	return api.Ready
}

func (l *Level) StartLevelState() (*api.LevelState, error) {
	state := &api.LevelState{}
	err := state.FromWaterSortLevelState(l.StartState)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (l *Level) Hint(levelState api.LevelState) (*api.HintResponse_Hint, error) {
	hint, err := WaterSortGraph.GetRandomNextStep(levelState)
	if err != nil {
		return nil, err
	}

	return hint, nil
}

func (l *Level) MinSteps() (*int, error) {
	state := &api.LevelState{}
	err := state.FromWaterSortLevelState(l.StartState)
	if err != nil {
		return nil, err
	}

	minSteps, err := WaterSortGraph.GetMinSteps(*state)
	if err != nil {
		return nil, err
	}

	return &minSteps, nil
}
