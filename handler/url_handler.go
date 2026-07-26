package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/eqweqr/urlshortener/service"
)

type UrlHandler struct {
	service service.Service
	logger  *slog.Logger
}

func NewHandler(s service.Service, l *slog.Logger) *UrlHandler {
	return &UrlHandler{
		service: s,
		logger:  l,
	}
}

func (h *UrlHandler) Shorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("failed to decode body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	url, err := h.service.Shorten(req.URL)
	if err != nil {
		h.logger.Warn("failed to shorten url", "url", req.URL, "error", err)
		http.Error(w, fmt.Sprintf("Failed to shorten URL: %v", err), http.StatusInternalServerError)
		return
	}

	baseURL := getBaseURL(r)
	shortURL := fmt.Sprintf("%s/%s", baseURL, url)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"short_url":    shortURL,
		"original_url": req.URL,
	})
}

func getBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, r.Host)
}

func (h *UrlHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		http.Error(w, "Short code is required", http.StatusBadRequest)
		return
	}

	url, err := h.service.Resolve(path)
	if err != nil {
		h.logger.Warn("short url not found", "path", path)
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, url, http.StatusMovedPermanently)
}
