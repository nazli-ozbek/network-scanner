package api

import (
	"encoding/json"
	"net/http"
	"network-scanner/logger"
	"network-scanner/model"
	"network-scanner/service"

	"github.com/gorilla/mux"
)

type RangeHandler struct {
	service *service.RangeService
	logger  logger.Logger
}

func NewRangeHandler(service *service.RangeService, logger logger.Logger) *RangeHandler {
	return &RangeHandler{service: service, logger: logger}
}

// ListRanges godoc
// @Summary Get all saved IP ranges
// @Produce json
// @Success 200 {array} model.IPRange
// @Router /ranges [get]
func (h *RangeHandler) ListRanges(w http.ResponseWriter, r *http.Request) {
	ranges, err := h.service.List()
	if err != nil {
		h.logger.Error("Failed to fetch ranges:", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(ranges)
}

// AddRange godoc
// @Summary Add a new IP range
// @Accept json
// @Produce json
// @Param input body model.IPRange true "IP Range"
// @Success 201 {string} string "Created"
// @Failure 400 {string} string "Invalid input"
// @Router /ranges [post]
func (h *RangeHandler) AddRange(w http.ResponseWriter, r *http.Request) {
	var input model.IPRange
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.ID == "" || input.Name == "" || input.Range == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	if err := h.service.Save(input); err != nil {
		h.logger.Error("Failed to save IP range:", err)
		http.Error(w, "Invalid CIDR or server error", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Created"))
}

// DeleteRange godoc
// @Summary Delete an IP range by ID
// @Param id path string true "IP Range ID"
// @Success 200 {string} string "Deleted"
// @Failure 404 {string} string "Not found"
// @Router /ranges/{id} [delete]
func (h *RangeHandler) DeleteRange(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.service.Delete(id); err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	w.Write([]byte("Deleted"))
}
