package ws

import (
	"sort"
	"strconv"
	"strings"
	"zalipuli/pkg/api"
)

const infinity = 1<<16 - 1

type Position struct {
	Vials         Vials    `json:"-"`
	Hash          string   `json:"hash"`
	NextPositions []string `json:"next_positions"`
	MinSteps      int      `json:"min_steps"`
	PrevPositions []string `json:"-"`
}

func (p *Position) isFinal() bool {
	return p.Hash == FinalPositionsHash[len(p.Vials)-3]
}

func NewPosition(vials Vials) *Position {
	copyVials := make(Vials, len(vials))
	copy(copyVials, vials)

	pos := &Position{
		Vials: copyVials, NextPositions: make([]string, 0), PrevPositions: make([]string, 0), MinSteps: infinity,
	}
	pos.normalize()
	return pos
}

func (p *Position) normalize() {
	sort.Sort(p.Vials)
	p.hash()
}

func (p *Position) hash() {
	strVials := make([]string, len(p.Vials))
	for i, vial := range p.Vials {
		strVials[i] = strconv.Itoa(int(vial))
	}

	p.Hash = strings.Join(strVials, "/")
}

func (p *Position) restoreVials() error {
	strVials := strings.Split(p.Hash, "/")
	p.Vials = make(Vials, len(strVials))

	for i, strVial := range strVials {
		vial, err := strconv.Atoi(strVial)
		if err != nil {
			return err
		}

		p.Vials[i] = Vial(vial)
	}

	return nil
}

func (p *Position) getStepVials(nextPosition *Position) (from Vial, to Vial) {
	changes := make([]int, 0, 2)

	// для начала сравним массивы флаконов у позиций и найдем флаконы, которых нет в следующей позиции
	for i, j := 0, 0; i < len(p.Vials) && j < len(nextPosition.Vials); {
		if p.Vials[i] == nextPosition.Vials[j] {
			i++
			j++
		} else {
			if p.Vials[i] < nextPosition.Vials[j] {
				changes = append(changes, i)
				if len(changes) == 2 {
					break
				}
				i++
			} else {
				j++
			}
		}
	}

	if len(changes) < 2 {
		changes = append(changes, p.Vials.Len()-1)
	}

	// дальше проверим, из какого флакона в какой совершалось переливание
	transfusePosition := p.transfuse(changes[0], changes[1])
	if transfusePosition.Hash == nextPosition.Hash {
		return p.Vials[changes[0]], p.Vials[changes[1]]
	}

	transfusePosition = p.transfuse(changes[1], changes[0])
	if transfusePosition.Hash == nextPosition.Hash {
		return p.Vials[changes[1]], p.Vials[changes[0]]
	}

	// вариант на случай, если nextPosition была передана не валидно
	return 0, 0
}

func (p *Position) isSuccessWay() bool {
	return p.MinSteps < infinity && p.MinSteps >= 0
}

func (p *Position) setMinStepsCount(nextPosition *Position) {
	var minSteps int
	if nextPosition.isSuccessWay() {
		minSteps = nextPosition.MinSteps + 1
	} else {
		minSteps = infinity
	}

	p.MinSteps = min(p.MinSteps, minSteps)
}

func (p *Position) canTransfuse(from, to int) bool {
	if from == to {
		return false
	}
	if p.Vials[from].Len() == 0 {
		return false
	}
	if p.Vials[to].Len() == 0 {
		return true
	}

	return p.Vials[from].LastSegment() == p.Vials[to].LastSegment() && p.Vials[to].Len() < VialHeight

}

// Transfuse создает новую позицию путем переливки воды из флакона from в флакон to.
// Переливается столько сегментов, сколько возможно перелить, от нуля до четырех
func (p *Position) transfuse(from, to int) *Position {
	transfusePosition := NewPosition(p.Vials)

	defer transfusePosition.normalize()

	for transfusePosition.canTransfuse(from, to) {
		transfusePosition.Vials[to] <<= ColorSize
		transfusePosition.Vials[to] |= transfusePosition.Vials[from] & (1<<ColorSize - 1)
		transfusePosition.Vials[from] >>= ColorSize
	}

	return transfusePosition
}

func (p *Position) addNext(nextPos *Position) {
	if p.Hash == nextPos.Hash {
		return
	}

	for _, next := range p.NextPositions {
		if next == nextPos.Hash {
			return
		}
	}

	p.NextPositions = append(p.NextPositions, nextPos.Hash)
	nextPos.PrevPositions = append(nextPos.PrevPositions, p.Hash)
}

func newPositionFromLevelState(ls api.LevelState) (*Position, error) {
	levelState, err := ls.AsWaterSortLevelState()
	if err != nil {
		return nil, err
	}

	apiVials := levelState.Vials
	vials := make(Vials, len(apiVials))

	for i, vial := range apiVials {
		vials[i] = NewVial(vial)
	}

	return NewPosition(vials), nil
}

func (p *Position) minSteps() int {
	if !p.isSuccessWay() {
		return -1
	}

	return p.MinSteps
}
