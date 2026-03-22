package entities

import (
	"math/rand"

	"github.com/google/uuid"
)

type WaterSortLevel struct {
	id            string
	colorsCount   uint8
	graph         *Graph
	isCorrect     bool
	startPosition [][]uint8
}

func NewWaterSortLevel() *WaterSortLevel {
	// Случайно выбираем число цветов, заполняем колбы вперемешку, и сами колбы тоже мешаем
	colorsCount := rand.Intn(MaxColorsCount-MinColorsCount) + MinColorsCount
	allSegments := make([]uint8, colorsCount*VialHeight)
	for i := 0; i < len(allSegments); i++ {
		allSegments[i] = uint8(i/VialHeight) + 1
	}
	rand.Shuffle(len(allSegments), func(i, j int) {
		allSegments[i], allSegments[j] = allSegments[j], allSegments[i]
	})

	flasks := make([]Vial, colorsCount+2)
	for i := 0; i < colorsCount; i++ {
		flasks[i] = NewVial(allSegments[i*VialHeight : i*VialHeight+VialHeight])
	}
	flasks[colorsCount] = NewVial(nil)
	flasks[colorsCount+1] = NewVial(nil)
	rand.Shuffle(len(flasks), func(i, j int) {
		flasks[i], flasks[j] = flasks[j], flasks[i]
	})

	apiFlasks := make([][]uint8, colorsCount+2)
	for i, flask := range flasks {
		apiFlasks[i] = flask.Segments()
	}

	// теперь формируем стартовую позицию, граф и уровень
	startPosition := NewPosition(flasks)

	level := &WaterSortLevel{
		id:            uuid.NewString(),
		graph:         NewGraph(startPosition),
		isCorrect:     true,
		startPosition: apiFlasks,
		colorsCount:   uint8(colorsCount),
	}

	go func() {
		err := level.graph.Build()
		if err != nil {
			level.isCorrect = false
		}
	}()

	return level
}

func (l *WaterSortLevel) Id() string {
	return l.id
}

func (l *WaterSortLevel) Status() string {
	if !l.isCorrect {
		return "not correct"
	}

	if l.graph.IsBuilt() {
		return "ready"
	}

	return "pending"
}

func (l *WaterSortLevel) ColorsCount() uint8 {
	return l.colorsCount
}

func (l *WaterSortLevel) MinSteps() (uint, error) {
	return l.graph.MinSteps()
}

func (l *WaterSortLevel) StartPosition() [][]uint8 {
	return l.startPosition
}

func (l *WaterSortLevel) Hint(apiFlasks [][]uint8) (int8, int8) {
	if !l.isCorrect || !l.graph.IsBuilt() {
		return -1, -1
	}

	flasks := make([]Vial, 0, len(apiFlasks))
	for _, flask := range apiFlasks {
		flasks = append(flasks, NewVial(flask))
	}

	position := NewPosition(flasks)
	nextPosition, err := l.graph.GetSuccessStep(position)
	if err != nil {
		return -1, -1
	}

	fromFlask, toFlask := position.GetStepVials(nextPosition)
	if fromFlask == 0 && toFlask == 0 {
		return -1, -1
	}

	var from, to int8
	for i, flask := range flasks {
		if flask == fromFlask {
			from = int8(i)
		} else if flask == toFlask {
			to = int8(i)
		}
	}

	return from, to
}
