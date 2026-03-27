package api

import (
	"encoding/json"
	"net/http"
	"zalipuli/internal/games"
	"zalipuli/internal/storage"

	ws "zalipuli/internal/games/watersort"
	"zalipuli/pkg/api"
)

type ZalipuliApi struct {
	levelRepo    storage.LevelRepository
	positionRepo storage.PositionRepository
	gameGraphs   map[api.GameName]games.Graph
}

func NewApi(lp storage.LevelRepository, ps storage.PositionRepository) api.ServerInterface {
	ws.WaterSortGraph = ws.NewGraph(ps)

	graphs := map[api.GameName]games.Graph{
		api.Watersort: ws.WaterSortGraph,
	}

	return &ZalipuliApi{levelRepo: lp, positionRepo: ps, gameGraphs: graphs}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (a *ZalipuliApi) GetHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, api.HealthResponse{Status: "ok"})
}

func (a *ZalipuliApi) PostLevelsStart(w http.ResponseWriter, r *http.Request) {
	var req api.StartLevelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Message: "invalid request body"})
		return
	}

	var level games.Level
	var err error

	switch req.GameName {
	case api.Watersort:
		level, err = ws.NewLevel()
	default:
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Message: "unknown game name"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, api.ErrorResponse{Message: err.Error()})
		return
	}

	err = a.levelRepo.SaveLevel(level)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, api.ErrorResponse{Message: err.Error()})
		return
	}

	startState, err := level.StartLevelState()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, api.ErrorResponse{Message: err.Error()})
		return
	}

	minSteps, _ := level.MinSteps()

	writeJSON(w, http.StatusCreated, api.LevelResponse{
		Id:              level.Id(),
		StartLevelState: *startState,
		MinSteps:        minSteps,
		Status:          level.Status(),
	})
}

func (a *ZalipuliApi) GetLevel(w http.ResponseWriter, _ *http.Request, levelId string) {
	level, err := a.levelRepo.GetLevel(levelId)
	if err != nil {
		writeJSON(w, http.StatusNotFound, api.ErrorResponse{Message: "level not found"})
		return
	}

	minSteps, _ := level.MinSteps()
	levelState, err := level.StartLevelState()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, api.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, api.LevelResponse{
		Id:              level.Id(),
		MinSteps:        minSteps,
		StartLevelState: *levelState,
		Status:          level.Status(),
	})
}

func (a *ZalipuliApi) FinishLevel(w http.ResponseWriter, r *http.Request) {
	var req api.FinishLevelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Message: "invalid request body"})
		return
	}

	if err := a.levelRepo.DeleteLevel(req.LevelId); err != nil {
		writeJSON(w, http.StatusNotFound, api.ErrorResponse{Message: "level not found"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *ZalipuliApi) PostLevelsHint(w http.ResponseWriter, r *http.Request) {
	var req api.HintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Message: "invalid request body"})
		return
	}

	level, err := a.levelRepo.GetLevel(req.LevelId)
	if err != nil {
		writeJSON(w, http.StatusNotFound, api.ErrorResponse{Message: "level not found"})
		return
	}

	hint, err := level.Hint(req.LevelState)
	if err != nil {
		writeJSON(w, http.StatusOK, api.HintResponse{
			IsSuccess: false,
		})
	}

	writeJSON(w, http.StatusOK, api.HintResponse{
		IsSuccess: true,
		Hint:      hint,
	})
}
