package api

import (
	"encoding/json"
	"net/http"
	"network-scanner/logger"
	"network-scanner/repository"
	"network-scanner/service"
	"strconv"
)

type ScanHandler struct {
	scanner     *service.ScannerService
	logger      logger.Logger
	historyRepo repository.ScanHistoryRepository
}

func NewScanHandler(scanner *service.ScannerService, logger logger.Logger, history repository.ScanHistoryRepository) *ScanHandler {
	return &ScanHandler{scanner: scanner, logger: logger, historyRepo: history}
}

type ScanRequest struct {
	IPRange string `json:"ip_range"`
}

// StartScan godoc
// @Summary Initiate a network scan
// @Description Starts scanning the given CIDR range in the background
// @Accept json
// @Produce json
// @Param input body ScanRequest true "IP range to scan"
// @Success 202 {object} map[string]string
// @Failure 400 {string} string "Bad request"
// @Router /scan [post]
func (h *ScanHandler) StartScan(w http.ResponseWriter, r *http.Request) {
	var body ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.IPRange == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		h.logger.Error("Invalid scan request body")
		return
	}

	h.logger.Info("Received scan request for range: ", body.IPRange)
	h.scanner.StartScan(body.IPRange)

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "Scan started"})
}

// RepeatScan godoc
// @Summary Repeat a past scan
// @Description Starts a new scan using the ip_range of a history record
// @Param id query int true "history id"
// @Produce json
// @Success 202 {object} map[string]string
// @Failure 400 {string} string "invalid id"
// @Failure 404 {string} string "not found"
// @Router /scan/repeat [post]
func (h *ScanHandler) RepeatScan(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	hist, err := h.historyRepo.GetByID(id)
	if err != nil || hist == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	h.logger.Info("Repeat scan requested for history id=", id, " range=", hist.IPRange)
	h.scanner.StartScan(hist.IPRange)

	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "Scan started"})
}
