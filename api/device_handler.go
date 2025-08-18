package api

import (
	"encoding/json"
	"net/http"
	"network-scanner/logger"
	"network-scanner/service"
)

type DeviceHandler struct {
	scanner *service.ScannerService
	logger  logger.Logger
}

func NewDeviceHandler(scanner *service.ScannerService, logger logger.Logger) *DeviceHandler {
	return &DeviceHandler{scanner: scanner, logger: logger}
}

// GetDevices godoc
// @Summary Get all discovered devices
// @Description Returns the list of scanned devices
// @Produce json
// @Success 200 {array} model.Device
// @Router /devices [get]
func (h *DeviceHandler) GetDevices(w http.ResponseWriter, r *http.Request) {
	devices := h.scanner.GetDevices()
	h.logger.Info("Fetched ", len(devices), " devices")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices)
}

// ClearDevices godoc
// @Summary Clear all stored devices (DEV ONLY)
// @Description Deletes all devices from in-memory store
// @Success 200 {string} string "Devices cleared"
// @Router /clear [delete]
func (h *DeviceHandler) ClearDevices(w http.ResponseWriter, r *http.Request) {
	h.logger.Warn("Clearing all device records via API")
	h.scanner.Clear()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Devices cleared"))
}
