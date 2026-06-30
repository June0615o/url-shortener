package handler

import (
	"net/http"
	"strconv"

	"github.com/panhao/url-shortener/internal/service"
)

type DashboardHandler struct {
	linkService *service.LinkService
}

func NewDashboardHandler(linkService *service.LinkService) *DashboardHandler {
	return &DashboardHandler{linkService: linkService}
}

// Overview handles GET /api/v1/dashboard/overview
func (h *DashboardHandler) Overview(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Authentication required"})
		return
	}

	overview, err := h.linkService.DashboardOverview(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, overview)
}

// Trend handles GET /api/v1/dashboard/trend
func (h *DashboardHandler) Trend(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Authentication required"})
		return
	}

	hours, _ := strconv.Atoi(r.URL.Query().Get("hours"))
	if hours < 1 || hours > 720 {
		hours = 24
	}

	data, err := h.linkService.DashboardTrend(r.Context(), userID, hours)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, data)
}

// Geo handles GET /api/v1/dashboard/geo
func (h *DashboardHandler) Geo(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Authentication required"})
		return
	}

	data, err := h.linkService.DashboardGeo(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, data)
}

// Devices handles GET /api/v1/dashboard/devices
func (h *DashboardHandler) Devices(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Authentication required"})
		return
	}

	data, err := h.linkService.DashboardDevices(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, data)
}
