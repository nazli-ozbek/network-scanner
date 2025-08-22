package api

import (
	"encoding/json"
	"net/http"
	"network-scanner/logger"
	"network-scanner/model"
	"network-scanner/service"
	"strings"

	"github.com/gorilla/mux"
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
	w.Write([]byte("Devices cleared"))
}

// GetDeviceByID godoc
// @Summary Get device by ID
// @Param id path string true "Device ID"
// @Produce json
// @Success 200 {object} model.Device
// @Failure 404 {string} string "Not found"
// @Router /devices/{id} [get]
func (h *DeviceHandler) GetDeviceByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	dev, err := h.scanner.FindByID(id)
	if err != nil || dev == nil {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(dev)
}

// AddTag godoc
// @Summary Add a tag to a device
// @Param id path string true "Device ID"
// @Accept json
// @Produce json
// @Param input body map[string]string true "Tag to add"
// @Success 200 {string} string "Tag added"
// @Router /devices/{id}/tags [post]
func (h *DeviceHandler) AddTag(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["tag"] == "" {
		http.Error(w, "Invalid tag", http.StatusBadRequest)
		return
	}

	dev, err := h.scanner.FindByID(id)
	if err != nil || dev == nil {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	tag := strings.TrimSpace(body["tag"])
	for _, t := range dev.Tags {
		if t == tag {
			w.Write([]byte("Tag already exists"))
			return
		}
	}
	dev.Tags = append(dev.Tags, tag)
	_ = h.scanner.UpdateTags(id, dev.Tags)
	w.Write([]byte("Tag added"))
}

// RemoveTag godoc
// @Summary Remove a tag from a device
// @Param id path string true "Device ID"
// @Accept json
// @Produce json
// @Param input body map[string]string true "Tag to remove"
// @Success 200 {string} string "Tag removed"
// @Router /devices/{id}/tags [delete]
func (h *DeviceHandler) RemoveTag(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["tag"] == "" {
		http.Error(w, "Invalid tag", http.StatusBadRequest)
		return
	}

	dev, err := h.scanner.FindByID(id)
	if err != nil || dev == nil {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	tag := strings.TrimSpace(body["tag"])
	newTags := make([]string, 0)
	for _, t := range dev.Tags {
		if t != tag {
			newTags = append(newTags, t)
		}
	}
	_ = h.scanner.UpdateTags(id, newTags)
	w.Write([]byte("Tag removed"))
}

// SearchDevices godoc
// @Summary Search devices by IP, hostname, or tags
// @Param query query string true "Search query"
// @Produce json
// @Success 200 {array} model.Device
// @Router /devices/search [get]
func (h *DeviceHandler) SearchDevices(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("query")
	if q == "" {
		http.Error(w, "Missing query", http.StatusBadRequest)
		return
	}
	result, err := h.scanner.SearchDevices(q)
	if err != nil {
		http.Error(w, "Search error", http.StatusInternalServerError)
		return
	}
	if result == nil {
		result = []model.Device{}
	}
	json.NewEncoder(w).Encode(result)
}
