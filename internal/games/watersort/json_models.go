package watersort

import (
	"encoding/json"
	"zalipuli/pkg/api"
)

type position struct {
	Hash          string   `json:"hash"`
	NextPositions []string `json:"next_positions"`
	IsSuccessWay  bool     `json:"is_success_way"`
}

type graph struct {
	StartPosition string                     `json:"start_position"`
	AllPositions  map[string]json.RawMessage `json:"all_positions"`
	VialsCount    int                        `json:"vials_count"`
	IsBuilt       bool                       `json:"is_built"`
	MinStepsCount int                        `json:"min_steps_count"`
}

type level struct {
	Id          string          `json:"id"`
	ColorsCount int             `json:"colors_count"`
	Graph       json.RawMessage `json:"graph"`
	GraphPtr    uintptr         `json:"graph_ptr"`
	IsCorrect   bool            `json:"is_correct"`
	StartState  api.Vials       `json:"start_state"`
}
