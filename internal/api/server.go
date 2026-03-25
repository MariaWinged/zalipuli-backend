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
	storage storage.LevelRepository
}

func NewApi(s storage.LevelRepository) api.ServerInterface {
	ws.FillConstants()
	return &ZalipuliApi{storage: s}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (h *ZalipuliApi) GetHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, api.HealthResponse{Status: "ok"})
}

func (h *ZalipuliApi) PostLevelsStart(w http.ResponseWriter, r *http.Request) {
	var req api.StartLevelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Message: "invalid request body"})
		return
	}

	var level games.Level
	switch req.GameName {
	case api.Watersort:
		level = ws.NewWaterSortLevel(h.storage)
	default:
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Message: "unknown game name"})
		return
	}

	h.storage.Save(level)
	startState, err := level.StartLevelState()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, api.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, api.LevelResponse{
		Id:              level.Id(),
		StartLevelState: *startState,
		MinSteps:        nil,
		Status:          api.New,
	})
}

func (h *ZalipuliApi) GetLevel(w http.ResponseWriter, _ *http.Request, levelId string) {
	level, err := h.storage.Get(levelId)
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

func (h *ZalipuliApi) FinishLevel(w http.ResponseWriter, r *http.Request) {
	var req api.FinishLevelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Message: "invalid request body"})
		return
	}

	if err := h.storage.Delete(req.LevelId); err != nil {
		writeJSON(w, http.StatusNotFound, api.ErrorResponse{Message: "level not found"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ZalipuliApi) PostLevelsHint(w http.ResponseWriter, r *http.Request) {
	var req api.HintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{Message: "invalid request body"})
		return
	}

	level, err := h.storage.Get(req.LevelId)
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
