package watersort

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
)

// Position - сущность, представляющая собой определенный набор флаконов
// Позиция не хранит действительный порядок флаконов, внутри нее флаконы отсортированы по своим хэшам
// Таким образом, две позиции эквивалентны, если в отсортированном порядке флаконы совпадают по своим сегментам
// В построенном графе позиция будет хранить все позиции, в которые она может перейти за один ход, а также все предыдущие позиции
// Кроме того, isSuccessWay в построенном графе показывает, можно ли из этой позиции прийти к успешному завершению уровня
type Position struct {
	vials         []Vial
	nextPositions []*Position
	prevPositions []*Position
	isSuccessWay  bool
}

// NewPosition создает новую позицию из массива флаконов
func NewPosition(vials []Vial) *Position {
	p := &Position{
		vials:         make([]Vial, len(vials)),
		nextPositions: make([]*Position, 0),
		prevPositions: make([]*Position, 0),
	}

	copy(p.vials, vials)
	p.normalize()

	return p
}

// CanTransfuse проверяет, можно ли перелить воду из флакона from в флакон to
func (p *Position) CanTransfuse(from, to int) bool {
	if from == to {
		return false
	}
	if p.vials[from].Len() == 0 {
		return false
	}
	if p.vials[to].Len() == 0 {
		return true
	}

	return p.vials[from].LastSegment() == p.vials[to].LastSegment() && p.vials[to].Len() < VialHeight

}

// Transfuse создает новую позицию путем переливки воды из флакона from в флакон to.
// Переливается столько сегментов, сколько возможно перелить, от нуля до четырех
func (p *Position) Transfuse(from, to int) *Position {
	transfusePosition := NewPosition(p.vials)

	defer transfusePosition.normalize()

	for transfusePosition.CanTransfuse(from, to) {
		transfusePosition.transfuse(from, to)
	}

	return transfusePosition
}

// transfuse переливает воду из флакона from в флакон to
// у этой операции нет валидации
func (p *Position) transfuse(from, to int) {
	p.vials[to] <<= ColorSize
	p.vials[to] |= p.vials[from] & (1<<ColorSize - 1)
	p.vials[from] >>= ColorSize
}

// Len - количество всех флаконов в позиции, включая пустые
func (p *Position) Len() int {
	return len(p.vials)
}

// Less - вспомогательная функция для сортировки
func (p *Position) Less(i, j int) bool {
	return p.vials[i] < p.vials[j]
}

// Swap - вспомогательная функция для сортировки
func (p *Position) Swap(i, j int) {
	p.vials[i], p.vials[j] = p.vials[j], p.vials[i]
}

// normalize - сортирует флаконы в позиции
func (p *Position) normalize() {
	sort.Sort(p)
}

// AddNext - добавляет позицию, в которую можно перейти из текущей
// Нужно проверить, что позиция не переходит сама в себя, и отсутствие дубликатов
func (p *Position) AddNext(next *Position) {
	if p.Hash() == next.Hash() {
		return
	}
	for _, position := range p.nextPositions {
		if position.Hash() == next.Hash() {
			return
		}
	}

	p.nextPositions = append(p.nextPositions, next)
	next.prevPositions = append(next.prevPositions, p)
}

// GetNext - возвращает все последующие позиции
func (p *Position) GetNext() []*Position {
	return p.nextPositions
}

// GetPrev - возвращает все предыдущие позиции
func (p *Position) GetPrev() []*Position {
	return p.prevPositions
}

// Hash - формирует hash-строку позиции: по сути просто записывает все хэши флаконов через слэш
// По хэшам позиции можно сравнивать между собой
func (p *Position) Hash() string {
	strVials := make([]string, len(p.vials))
	for i, vial := range p.vials {
		strVials[i] = strconv.Itoa(int(vial))
	}

	hash := strings.Join(strVials, "/")

	return hash
}

// IsFinal - является ли позиция завершением уровня, то есть позицией, где все флаконы полностью заполнены одноцветными сегментами
// Для удобства хэш просто сравнивается с уже сформированным финальным хэшем для количества цветов в позиции
func (p *Position) IsFinal() bool {
	return p.Hash() == FinalPositionsHash[p.Len()-3]
}

// IsSuccessWay - в построенном графе говорит о том, что из этой позиции можно прийти к успешному завершению уровня
func (p *Position) IsSuccessWay() bool {
	return p.isSuccessWay
}

// SetIsSuccessWay - установить isSuccessWay
func (p *Position) SetIsSuccessWay(isSuccessWay bool) {
	p.isSuccessWay = isSuccessWay
}

// GetStepVials - находит флаконы, которые переливали из одного в другую
func (p *Position) GetStepVials(nextPosition *Position) (from Vial, to Vial) {
	changes := make([]int, 0, 2)

	// для начала сравним массивы флаконов у позиций и найдем флаконы, которых нет в следующей позиции
	for i, j := 0, 0; i < len(p.vials) && j < len(nextPosition.vials); {
		if p.vials[i] == nextPosition.vials[j] {
			i++
			j++
		} else {
			if p.vials[i] < nextPosition.vials[j] {
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
		changes = append(changes, p.Len()-1)
	}

	// дальше проверим, из какого флакона в какой совершалось переливание
	transfusePosition := p.Transfuse(changes[0], changes[1])
	if transfusePosition.Hash() == nextPosition.Hash() {
		return p.vials[changes[0]], p.vials[changes[1]]
	}

	transfusePosition = p.Transfuse(changes[1], changes[0])
	if transfusePosition.Hash() == nextPosition.Hash() {
		return p.vials[changes[1]], p.vials[changes[0]]
	}

	// вариант на случай, если nextPosition была передана не валидно
	return 0, 0
}

func (p *Position) ToJson() (json.RawMessage, error) {
	nextPositions := make([]string, len(p.nextPositions))
	for i, nextPos := range p.nextPositions {
		nextPositions[i] = nextPos.Hash()
	}

	return json.Marshal(position{
		Hash:          p.Hash(),
		IsSuccessWay:  p.IsSuccessWay(),
		NextPositions: nextPositions,
	})
}

func (p *Position) FromHash(hash string) error {
	strVials := strings.Split(hash, "/")
	vials := make([]Vial, len(strVials))
	for i, strVial := range strVials {
		vial, err := strconv.Atoi(strVial)
		if err != nil {
			return err
		}
		vials[i] = Vial(vial)
	}
	p.vials = vials
	p.nextPositions = make([]*Position, 0)

	return nil
}

func (p *Position) FromJson(jsonPos json.RawMessage) ([]string, error) {
	pos := position{}
	err := json.Unmarshal(jsonPos, &pos)
	if err != nil {
		return nil, err
	}

	err = p.FromHash(pos.Hash)
	if err != nil {
		return nil, err
	}
	p.isSuccessWay = pos.IsSuccessWay

	return pos.NextPositions, nil
}
