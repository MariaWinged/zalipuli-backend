package watersort

import (
	"encoding/json"
	"errors"
	"math/rand"
	"sync"
)

// Graph - сущность, представляющая собой все возможные пути решения уровня
// Алгоритм построения работает достаточно долго, поэтому подразумевается асинхронное построение
// Основные функции графа это подсказки, а также информация о минимальном количестве ходов для прохождения уровня
type Graph struct {
	startPosition *Position
	allPositions  map[string]*Position
	vialsCount    int
	isBuilt       bool
	minStepsCount int
	buildMt       sync.RWMutex
	stateMt       sync.RWMutex
}

// NewGraph - создает граф из стартовой позиции
func NewGraph(startPosition *Position) *Graph {
	return &Graph{
		startPosition: startPosition,
		allPositions:  make(map[string]*Position),
		vialsCount:    startPosition.Len(),
	}
}

// Build - строит граф всех позиций по алгоритму поиска в ширину
// Если стартовая позиция корректна (содержит 2 пустых колбы и ровно 4 сегмента каждого цвета),
// то граф гарантированно построится. В противном случае успешное построение не гарантированно
func (g *Graph) Build() error {
	g.buildMt.Lock()
	defer g.buildMt.Unlock()

	if g.isBuilt {
		return nil
	}

	queue := make([]*Position, 0)
	queue = append(queue, g.startPosition)
	g.allPositions[g.startPosition.Hash()] = g.startPosition

	// Строим граф поиском в ширину: из стартовой позиции ищем все возможные позиции, в которые мы можем перейти.
	// Если новой позиции еще нет в очереди, добавляем ее туда.
	// Граф в любом случае конечный, тк каждым переливанием мы либо не изменяем,
	// либо сокращаем число одноцветных сегментов, объединяя их друг с другом.
	// При этом количество шагов, при которых число сегментов не изменяется, ограничено из-за общего количества сегментов одного цвета
	for p := 0; p < len(queue); p++ {
		for from := 0; from < g.vialsCount; from++ {
			for to := 0; to < g.vialsCount; to++ {
				if queue[p].CanTransfuse(from, to) {
					nextPosition := queue[p].Transfuse(from, to)
					if next, ok := g.allPositions[nextPosition.Hash()]; !ok {
						g.allPositions[nextPosition.Hash()] = nextPosition
						queue = append(queue, nextPosition)
					} else {
						// переприсвоим на уже существующую, чтобы сохранить связи позиции
						nextPosition = next
					}

					queue[p].AddNext(nextPosition)
				}
			}
		}
	}

	// ищем финальную позицию успешного завершения уровня
	finalPosition := g.allPositions[FinalPositionsHash[g.vialsCount-3]]
	if finalPosition == nil {
		return errors.New("final position not found")
	}

	// восстанавливаем пути к успеху,
	// в мапе successPositions храним количество ходов, необходимых для того, чтобы прийти из позиции к финальной
	successPositions := make(map[string]int)
	queue = nil
	queue = append(queue, finalPosition)
	finalPosition.SetIsSuccessWay(true)
	successPositions[finalPosition.Hash()] = 0

	for p := 0; p < len(queue); p++ {
		prevPositions := queue[p].GetPrev()
		for _, prevPosition := range prevPositions {
			prevPosition.SetIsSuccessWay(true)
			if _, ok := successPositions[prevPosition.Hash()]; !ok {
				successPositions[prevPosition.Hash()] = successPositions[queue[p].Hash()] + 1
				queue = append(queue, prevPosition)
			} else {
				// если мы уже встречали эту позицию, то стоит проверить, что в ней лежит минимальный путь
				successPositions[prevPosition.Hash()] = min(successPositions[prevPosition.Hash()], successPositions[queue[p].Hash()]+1)
			}
		}
	}

	ok := false
	// так как мы уже построили граф, то и обратно тоже должны найти путь
	// тем не менее, лучше все-таки обработать ошибку из-за потенциальных багов в алгоритме
	if g.minStepsCount, ok = successPositions[g.startPosition.Hash()]; !ok {
		return errors.New("could not find success way")
	}

	g.isBuilt = true
	g.buildMt.Unlock()

	// имеет смысл хранить только успешные позиции
	// кроме того, можно уже не хранить предыдущие позиции
	go func() {
		newPositionsMap := make(map[string]*Position)
		for hash := range successPositions {
			pos := g.allPositions[hash]
			nextPositions := make([]*Position, 0)
			for _, next := range pos.nextPositions {
				if next.isSuccessWay {
					nextPositions = append(nextPositions, next)
				}
			}
			pos.nextPositions = nextPositions
			pos.prevPositions = nil

			newPositionsMap[hash] = pos
		}

		g.stateMt.Lock()
		defer g.stateMt.Unlock()
		g.allPositions = newPositionsMap
	}()

	return nil
}

// IsBuilt - статус постройки графа
func (g *Graph) IsBuilt() bool {
	g.buildMt.RLock()
	defer g.buildMt.RUnlock()

	return g.isBuilt
}

// MinSteps - минимальное число шагов, за которое можно из стартовой позиции прийти к успеху
func (g *Graph) MinSteps() (int, error) {
	if !g.isBuilt {
		return 0, errors.New("graph is not built yet")
	}

	return g.minStepsCount, nil
}

// GetSuccessStep возвращает следующую позицию для успешного завершения уровня
// Следующая позиция выбирается случайно из доступных путей к успеху
// и не гарантирует, что этот путь будет кратчайшим.
// Если передаваемая позиция некорректна или ведет в тупик, вернется ошибка
func (g *Graph) GetSuccessStep(position *Position) (*Position, error) {
	if !g.isBuilt {
		return nil, errors.New("graph is not built yet")
	}

	g.stateMt.RLock()
	defer g.stateMt.RUnlock()

	graphPosition, ok := g.allPositions[position.Hash()]
	if !ok {
		return nil, errors.New("position not found in graph")
	}

	if !graphPosition.IsSuccessWay() {
		return nil, errors.New("no success way from this position")
	}

	successPositions := make([]*Position, 0)

	for _, nextPosition := range graphPosition.GetNext() {
		if nextPosition.IsSuccessWay() {
			successPositions = append(successPositions, nextPosition)
		}
	}

	return successPositions[rand.Intn(len(successPositions))], nil
}

func (g *Graph) ToJson() (json.RawMessage, error) {
	g.buildMt.Lock()
	defer g.buildMt.Unlock()

	g.stateMt.RLock()
	defer g.stateMt.RUnlock()

	allPositions := make(map[string]json.RawMessage)
	for hash, pos := range g.allPositions {
		jsonPos, err := pos.ToJson()
		if err != nil {
			return nil, err
		}
		allPositions[hash] = jsonPos
	}

	return json.Marshal(graph{
		StartPosition: g.startPosition.Hash(),
		AllPositions:  allPositions,
		VialsCount:    g.vialsCount,
		IsBuilt:       g.isBuilt,
		MinStepsCount: g.minStepsCount,
	})
}

func (g *Graph) FromJson(jsonGraph json.RawMessage) error {
	gr := graph{}
	err := json.Unmarshal(jsonGraph, &gr)
	if err != nil {
		return err
	}

	g.isBuilt = gr.IsBuilt
	g.minStepsCount = gr.MinStepsCount
	g.vialsCount = gr.VialsCount
	g.startPosition = &Position{}
	err = g.startPosition.FromHash(gr.StartPosition)
	if err != nil {
		return err
	}

	if !g.isBuilt {
		return nil
	}

	g.stateMt.Lock()
	defer g.stateMt.Unlock()

	g.allPositions = make(map[string]*Position)
	nextPositions := make(map[string][]string)
	for hash, pos := range gr.AllPositions {
		newPosition := &Position{}
		next, err := newPosition.FromJson(pos)
		if err != nil {
			return err
		}

		g.allPositions[hash] = newPosition
		nextPositions[hash] = next
	}

	for hash, next := range nextPositions {
		for _, nextHash := range next {
			g.allPositions[hash].nextPositions = append(g.allPositions[hash].nextPositions, g.allPositions[nextHash])
		}
	}

	return nil
}
