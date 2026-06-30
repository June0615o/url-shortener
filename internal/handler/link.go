package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/panhao/url-shortener/internal/model"
	"github.com/panhao/url-shortener/internal/service"
)

type LinkHandler struct {
	linkService *service.LinkService
}

func NewLinkHandler(linkService *service.LinkService) *LinkHandler {
	return &LinkHandler{linkService: linkService}
}

// Create handles POST /api/v1/links
func (h *LinkHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateLinkReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	var userID *int64
	if uid, ok := r.Context().Value("user_id").(int64); ok {
		userID = &uid
	}

	resp, err := h.linkService.Create(r.Context(), req, userID)
	if err != nil {
		log.Printf("Error creating link: %v", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// Get handles GET /api/v1/links/:code
func (h *LinkHandler) Get(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "short code is required"})
		return
	}

	link, err := h.linkService.Get(r.Context(), code)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}
	if link == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Link not found"})
		return
	}

	// Check ownership for authenticated requests
	if uid, ok := r.Context().Value("user_id").(int64); ok {
		if link.UserID != nil && *link.UserID != uid {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "Access denied"})
			return
		}
	}

	resp := model.LinkResp{
		Link:     *link,
		ShortURL: r.Host + "/" + link.ShortCode,
	}
	writeJSON(w, http.StatusOK, resp)
}

// List handles GET /api/v1/links
func (h *LinkHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Authentication required"})
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	resp, err := h.linkService.List(r.Context(), userID, page, pageSize)
	if err != nil {
		log.Printf("Error listing links: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// Update handles PATCH /api/v1/links/:code
func (h *LinkHandler) Update(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "short code is required"})
		return
	}

	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Authentication required"})
		return
	}

	// Verify ownership
	_, err := h.linkService.GetUserLink(r.Context(), code, userID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Link not found or access denied"})
		return
	}

	var req model.UpdateLinkReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	link, err := h.linkService.Update(r.Context(), code, req)
	if err != nil {
		log.Printf("Error updating link: %v", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	resp := model.LinkResp{
		Link:     *link,
		ShortURL: r.Host + "/" + link.ShortCode,
	}
	writeJSON(w, http.StatusOK, resp)
}

// Delete handles DELETE /api/v1/links/:code
func (h *LinkHandler) Delete(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "short code is required"})
		return
	}

	_, ok := r.Context().Value("user_id").(int64)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Authentication required"})
		return
	}

	err := h.linkService.Delete(r.Context(), code)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}

// GetStats handles GET /api/v1/links/:code/stats
func (h *LinkHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "short code is required"})
		return
	}

	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Authentication required"})
		return
	}

	stats, err := h.linkService.GetStats(r.Context(), code, userID)
	if err != nil {
		log.Printf("Error getting stats: %v", err)
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("Error encoding JSON response: %v", err)
		}
	}
}
