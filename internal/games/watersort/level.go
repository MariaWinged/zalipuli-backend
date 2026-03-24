package watersort

import (
	"errors"
	"math/rand"
	"zalipuli/internal/storage"
	"zalipuli/pkg/api"

	"github.com/google/uuid"
)

type Level struct {
	id            string
	colorsCount   int
	graph         *Graph
	isCorrect     bool
	startPosition api.Vials
}

func NewWaterSortLevel(storage *storage.Storage) *Level {
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

	apiVials := make(api.Vials, colorsCount+2)
	for i, vial := range vials {
		apiVials[i] = vial.Segments()
	}

	// теперь формируем стартовую позицию, граф и уровень
	startPosition := NewPosition(vials)

	level := &Level{
		id:            uuid.NewString(),
		graph:         NewGraph(startPosition),
		isCorrect:     true,
		startPosition: apiVials,
		colorsCount:   colorsCount,
	}

	go func() {
		err := level.graph.Build()
		if err != nil {
			level.isCorrect = false
		}

		storage.Save(level)
	}()

	return level
}

func (l *Level) Id() string {
	return l.id
}

func (l *Level) Status() api.LevelResponseStatus {
	if !l.isCorrect {
		return api.Incorrect
	}

	if l.graph.IsBuilt() {
		return api.Ready
	}

	return api.Pending
}

func (l *Level) GameName() api.GameName {
	return api.Watersort
}

func (l *Level) ColorsCount() int {
	return l.colorsCount
}

func (l *Level) MinSteps() (*int, error) {
	minSteps, err := l.graph.MinSteps()
	if err != nil {
		return nil, err
	}
	return &minSteps, nil
}

func (l *Level) StartLevelState() (*api.LevelState, error) {
	var state api.LevelState
	err := state.FromWaterSortLevelState(api.WaterSortLevelState{
		ColorsCount: &l.colorsCount,
		GameName:    api.Watersort,
		Vials:       l.startPosition,
	})
	if err != nil {
		return nil, err
	}

	return &state, nil
}

func (l *Level) Hint(levelState api.LevelState) (*api.HintResponse_Hint, error) {
	if !l.isCorrect || !l.graph.IsBuilt() {
		return nil, errors.New("no hint available")
	}

	wsLevelState, err := levelState.AsWaterSortLevelState()
	if err != nil {
		return nil, err
	}

	apiVials := wsLevelState.Vials

	vials := make([]Vial, 0, len(apiVials))
	for _, vial := range apiVials {
		vials = append(vials, NewVial(vial))
	}

	position := NewPosition(vials)
	nextPosition, err := l.graph.GetSuccessStep(position)
	if err != nil {
		return nil, err
	}

	fromVial, toVial := position.GetStepVials(nextPosition)
	if fromVial == 0 && toVial == 0 {
		return nil, errors.New("no hint available")
	}

	var from, to int
	for i, vial := range vials {
		if vial == fromVial {
			from = i
		} else if vial == toVial {
			to = i
		}
	}

	var hint api.HintResponse_Hint
	err = hint.FromWaterSortHint(api.WaterSortHint{
		GameName:      api.Watersort,
		VialIndexFrom: from,
		VialIndexTo:   to,
	})
	if err != nil {
		return nil, err
	}

	return &hint, nil
}
