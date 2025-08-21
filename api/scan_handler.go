package api

import (
	"encoding/json"
	"net/http"
	"network-scanner/logger"
	"network-scanner/service"
)

type ScanHandler struct {
	scanner *service.ScannerService
	logger  logger.Logger
}

func NewScanHandler(scanner *service.ScannerService, logger logger.Logger) *ScanHandler {
	return &ScanHandler{scanner: scanner, logger: logger}
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
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "Scan started"})
}
