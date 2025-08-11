package api

import (
	"encoding/json"
	"net/http"
	"network-scanner/service"
)

type Handler struct {
	scanner *service.ScannerService
}

func NewHandler(scanner *service.ScannerService) *Handler {
	return &Handler{scanner: scanner}
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
func (h *Handler) StartScan(w http.ResponseWriter, r *http.Request) {
	var body ScanRequest

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.IPRange == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	h.scanner.StartScan(body.IPRange)
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "Scan started"})
}

// GetDevices godoc
// @Summary Get all discovered devices
// @Description Returns the list of scanned devices
// @Produce json
// @Success 200 {array} model.Device
// @Router /devices [get]
func (h *Handler) GetDevices(w http.ResponseWriter, r *http.Request) {
	devices := h.scanner.GetDevices()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices)
}

// ClearDevices godoc
// @Summary Clear all stored devices (DEV ONLY)
// @Description Deletes all devices from in-memory store
// @Success 200 {string} string "Devices cleared"
// @Router /clear [delete]
func (h *Handler) ClearDevices(w http.ResponseWriter, r *http.Request) {
	h.scanner.Clear()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Devices cleared"))
}
