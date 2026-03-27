package ws

import (
	"errors"
	"math/rand"
	"zalipuli/internal/games"
	"zalipuli/internal/storage"
	"zalipuli/pkg/api"
)

type Graph struct {
	storage storage.PositionRepository
	errChan chan error
}

func NewGraph(s storage.PositionRepository) *Graph {
	FillConstants()
	return &Graph{storage: s, errChan: make(chan error)}
}

func (g Graph) GameName() api.GameName {
	return api.Watersort
}

func (g Graph) build(startPosition *Position) error {
	err := g.storage.DeletePosition(gameName, startPosition.Hash)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return err
	}

	queue := make([]*Position, 0)
	queue = append(queue, startPosition)
	allPositions := make(map[string]*Position)
	allPositions[startPosition.Hash] = startPosition

	minStepsQueue := make([]*Position, 0)

	for p := 0; p < len(queue); p++ {
		storagePos, err := g.getPosition(queue[p].Hash)
		if err == nil {
			queue[p].MinSteps = storagePos.MinSteps
			for _, nextHash := range storagePos.NextPositions {
				if nextPos, ok := allPositions[nextHash]; ok {
					queue[p].addNext(nextPos)
				} else {
					queue[p].NextPositions = append(queue[p].NextPositions, nextHash)
				}
			}

			minStepsQueue = append(minStepsQueue, queue[p])
			continue
		}

		for i := 0; i < len(queue[p].Vials); i++ {
			for j := 0; j < len(queue[p].Vials); j++ {
				if queue[p].canTransfuse(i, j) {
					newPos := queue[p].transfuse(i, j)
					if allPositions[newPos.Hash] == nil {
						allPositions[newPos.Hash] = newPos
						queue = append(queue, newPos)
					} else {
						newPos = allPositions[newPos.Hash]
					}
					queue[p].addNext(newPos)
				}
			}
		}
	}

	finalHash := FinalPositionsHash[len(startPosition.Vials)-3]
	if finalPos, ok := allPositions[finalHash]; ok {
		finalPos.MinSteps = 0
		minStepsQueue = append(minStepsQueue, finalPos)
	}

	minStepsPositons := make(map[string]bool)

	for p := 0; p < len(minStepsQueue); p++ {
		for _, prevHash := range minStepsQueue[p].PrevPositions {
			prevPos := allPositions[prevHash]
			if prevPos == nil {
				continue
			}
			prevPos.setMinStepsCount(minStepsQueue[p])
			if !minStepsPositons[prevPos.Hash] {
				minStepsPositons[prevPos.Hash] = true
				minStepsQueue = append(minStepsQueue, prevPos)
			}
		}
	}

	for _, pos := range allPositions {
		err := g.storage.SavePosition(gameName, pos.Hash, pos)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g Graph) getPosition(hash string) (*Position, error) {
	var position Position
	err := g.storage.GetPosition(gameName, hash, &position)
	if err != nil {
		return nil, err
	}

	err = position.restoreVials()
	if err != nil {
	}

	return &position, nil
}

func (g Graph) startBuild(position *Position) error {
	select {
	case err := <-g.errChan:
		return err
	default:
		go func() {
			err := g.build(position)
			if err != nil {
				g.errChan <- err
			}
		}()
	}
	return games.NotReadyErr
}

func (g Graph) GetMinSteps(state api.LevelState) (int, error) {
	position, err := newPositionFromLevelState(state)
	if err != nil {
		return 0, err
	}

	if storagePos, err := g.getPosition(position.Hash); err == nil {
		position = storagePos
		if !position.isSuccessWay() {
			return 0, games.NotSuccessWayErr
		}

		return position.minSteps(), nil
	}

	return 0, g.startBuild(position)
}

func (g Graph) GetRandomNextStep(state api.LevelState) (*api.HintResponse_Hint, error) {
	position, err := newPositionFromLevelState(state)
	if err != nil {
		return nil, err
	}

	if storagePos, err := g.getPosition(position.Hash); err == nil {
		position = storagePos

		if !position.isSuccessWay() {
			return nil, errors.New("no next level")
		}
		successNextPositions := make([]*Position, 0)
		for _, nextPosHash := range position.NextPositions {
			nextPos, err := g.getPosition(nextPosHash)
			if err != nil {
				if errors.Is(err, storage.ErrNotFound) {
					return nil, g.startBuild(position)
				}

				return nil, err
			}
			successNextPositions = append(successNextPositions, nextPos)
		}

		if len(successNextPositions) == 0 {
			return nil, games.NotSuccessWayErr
		}

		return g.getNext(state, position, successNextPositions[rand.Intn(len(successNextPositions))])
	}

	return nil, g.startBuild(position)
}

func (g Graph) getNext(state api.LevelState, pos *Position, nextPos *Position) (*api.HintResponse_Hint, error) {
	wsLevelState, err := state.AsWaterSortLevelState()
	if err != nil {
		return nil, err
	}

	fromVial, toVial := pos.getStepVials(nextPos)
	if fromVial == 0 && toVial == 0 {
		return nil, games.NotSuccessWayErr
	}

	vials := convertFromApiVials(wsLevelState.Vials)

	from, to := -1, -1

	for i, vial := range vials {
		if vial == fromVial && (from == -1 || rand.Intn(2) == 0) {
			from = i
		} else if vial == toVial && (to == -1 || rand.Intn(2) == 0) {
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

func (g Graph) IsFinal(state api.LevelState) bool {
	position, err := newPositionFromLevelState(state)
	if err != nil {
		return false
	}

	return position.isFinal()
}
