package api

import (
	"encoding/json"
	"fmt"
	"mangascraper/internal/db"
	"mangascraper/internal/scraper"
	"net/http"
	"strings"
)

// Server represents the HTTP server
type Server struct {
	port string
}

// NewServer creates a new HTTP server
func NewServer(port string) *Server {
	if port == "" {
		port = "8080"
	}
	return &Server{port: port}
}

// DownloadRequest represents a request to download a manga
type DownloadRequest struct {
	Title string `json:"title"`
}

// DownloadResponse represents the response from a download request
type DownloadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	MangaID int    `json:"manga_id,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// Start starts the HTTP server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/manga/download", s.handleDownloadManga)
	mux.HandleFunc("/api/manga", s.handleListManga)
	mux.HandleFunc("/api/manga/", s.handleMangaDetail)

	fmt.Printf("HTTP server starting on port %s\n", s.port)
	fmt.Printf("Endpoints:\n")
	fmt.Printf("  POST /api/manga/download - Download a manga (JSON body: {\"title\": \"manga-slug\"})\n")
	fmt.Printf("  GET  /api/manga           - List all manga\n")
	fmt.Printf("  GET  /api/manga/{slug}    - Get manga details and chapters\n")
	fmt.Printf("  GET  /api/health          - Health check\n")

	return http.ListenAndServe(":"+s.port, mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	response := map[string]string{
		"status":  "ok",
		"service": "mangascraper",
	}
	s.sendJSON(w, http.StatusOK, response)
}

func (s *Server) handleDownloadManga(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed, use POST")
		return
	}

	var req DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid JSON body: "+err.Error())
		return
	}

	if req.Title == "" {
		s.sendError(w, http.StatusBadRequest, "Title is required")
		return
	}

	// Get cookies first
	if err := scraper.GetCookies(); err != nil {
		s.sendError(w, http.StatusServiceUnavailable, "Failed to get cookies: "+err.Error())
		return
	}

	// Start download in a goroutine so it doesn't block the response
	go func(title string) {
		if err := scraper.DownloadManga(title); err != nil {
			fmt.Printf("Error downloading manga %s: %v\n", title, err)
		}
	}(req.Title)

	// Get the manga ID that was just created
	manga, err := db.GetMangaBySlug(req.Title)
	mangaID := 0
	if err == nil && manga != nil {
		mangaID = manga.MangaID
	}

	response := DownloadResponse{
		Success: true,
		Message: "Download started for: " + req.Title,
		MangaID: mangaID,
	}
	s.sendJSON(w, http.StatusAccepted, response)
}

func (s *Server) handleListManga(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed, use GET")
		return
	}

	// Check if it's a detail request
	path := strings.TrimPrefix(r.URL.Path, "/api/manga")
	if path != "" && path != "/" {
		s.handleMangaDetail(w, r)
		return
	}

	mangaList, err := db.GetAllManga()
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get manga list: "+err.Error())
		return
	}

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    mangaList,
		"count":   len(mangaList),
	})
}

func (s *Server) handleMangaDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed, use GET")
		return
	}

	// Extract slug from path: /api/manga/{slug}
	path := strings.TrimPrefix(r.URL.Path, "/api/manga/")
	if path == "" {
		s.sendError(w, http.StatusBadRequest, "Manga slug is required")
		return
	}

	slug := strings.TrimSpace(path)

	// Get manga
	manga, err := db.GetMangaBySlug(slug)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get manga: "+err.Error())
		return
	}

	if manga == nil {
		s.sendError(w, http.StatusNotFound, "Manga not found: "+slug)
		return
	}

	// Get chapters
	chapters, err := db.GetChaptersByManga(manga.MangaID)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get chapters: "+err.Error())
		return
	}

	response := map[string]interface{}{
		"success":  true,
		"manga":    manga,
		"chapters": chapters,
	}
	s.sendJSON(w, http.StatusOK, response)
}

func (s *Server) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) sendError(w http.ResponseWriter, status int, message string) {
	s.sendJSON(w, status, ErrorResponse{
		Success: false,
		Error:   message,
	})
}
