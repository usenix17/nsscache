package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"nsscache-http/cache"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	cache *cache.Cache
}

// New creates a new Handler
func New(c *cache.Cache) *Handler {
	return &Handler{cache: c}
}

// PasswdJSON returns users as JSON
func (h *Handler) PasswdJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(h.cache.GetUsers()); err != nil {
		log.Printf("error encoding passwd json: %v", err)
	}
}

// PasswdFlat returns users in passwd file format
func (h *Handler) PasswdFlat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	users := h.cache.GetUsers()
	var sb strings.Builder
	for _, u := range users {
		sb.WriteString(u.ToPasswdLine())
		sb.WriteByte('\n')
	}
	w.Write([]byte(sb.String()))
}

// GroupJSON returns groups as JSON
func (h *Handler) GroupJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(h.cache.GetGroups()); err != nil {
		log.Printf("error encoding group json: %v", err)
	}
}

// GroupFlat returns groups in group file format
func (h *Handler) GroupFlat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	groups := h.cache.GetGroups()
	var sb strings.Builder
	for _, g := range groups {
		sb.WriteString(g.ToGroupLine())
		sb.WriteByte('\n')
	}
	w.Write([]byte(sb.String()))
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Users     int       `json:"users"`
	Groups    int       `json:"groups"`
	LastFetch time.Time `json:"last_fetch"`
	CacheAge  string    `json:"cache_age"`
}

// Health returns cache health status
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	users, groups, lastFetch := h.cache.Stats()

	resp := HealthResponse{
		Status:    "ok",
		Users:     users,
		Groups:    groups,
		LastFetch: lastFetch,
		CacheAge:  time.Since(lastFetch).Round(time.Second).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("error encoding health json: %v", err)
	}
}
