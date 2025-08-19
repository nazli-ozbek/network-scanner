package api

import (
	"encoding/json"
	"net/http"
	"network-scanner/logger"
	"network-scanner/repository"
	"strconv"
	"strings"
)

type ScanHistoryHandler struct {
	repo   repository.ScanHistoryRepository
	logger logger.Logger
}

func NewScanHistoryHandler(repo repository.ScanHistoryRepository, logger logger.Logger) *ScanHistoryHandler {
	return &ScanHistoryHandler{repo: repo, logger: logger}
}

// GetScanHistory godoc
// @Summary Get scan history
// @Description Returns the list of past scans
// @Produce json
// @Success 200 {array} model.ScanHistory
// @Router /scan-history [get]
func (h *ScanHistoryHandler) GetScanHistory(w http.ResponseWriter, r *http.Request) {
	hist, err := h.repo.GetAll()
	if err != nil {
		http.Error(w, "Failed to load history", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(hist)
}

// DeleteScanHistory godoc
// @Summary Delete a scan history record
// @Param id path int true "history id"
// @Success 200 {string} string "deleted"
// @Failure 400 {string} string "invalid id"
// @Failure 404 {string} string "not found"
// @Router /scan-history/{id} [delete]
func (h *ScanHistoryHandler) DeleteScanHistory(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if _, err := h.repo.GetByID(id); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err := h.repo.Delete(id); err != nil {
		http.Error(w, "delete failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("deleted"))
}

// ClearScanHistory godoc
// @Summary Delete all scan history records
// @Success 200 {string} string "cleared"
// @Router /scan-history [delete]
func (h *ScanHistoryHandler) ClearScanHistory(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.Clear(); err != nil {
		http.Error(w, "clear failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("cleared"))
}
