package api

import (
	"encoding/json"
	"fmt"
	"io"
	"mangascraper/internal/db"
	"mangascraper/internal/middleware"
	"mangascraper/internal/scraper"
	"net/http"
	"strings"
)

// InitScraper initializes the scraper by getting cookies from FlareSolverr
func InitScraper() error {
	fmt.Println("Initializing scraper - getting Cloudflare cookies...")
	if err := scraper.GetCookies(); err != nil {
		return fmt.Errorf("failed to get cookies: %w", err)
	}
	fmt.Println("Scraper initialized successfully")
	return nil
}

// Server represents the HTTP server
type Server struct {
	port string
}

// NewServer creates a new HTTP server
func NewServer(port string) *Server {
	if port == "" {
		port = "8081"
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
	mux.HandleFunc("/api/search", s.handleSearchManga)
	mux.HandleFunc("/api/token", s.handleGetToken)
	mux.HandleFunc("/api/chapter/", s.handleChapterPages)
	mux.HandleFunc("/api/manga/", s.handleMangaDetail)

	fmt.Printf("HTTP server starting on port %s\n", s.port)
	fmt.Printf("Endpoints:\n")
	fmt.Printf("  POST /api/manga/download    - Download a manga (JSON body: {\"title\": \"manga-slug\"})\n")
	fmt.Printf("  GET  /api/manga             - List all manga\n")
	fmt.Printf("  GET  /api/manga/{slug}      - Get manga details and chapters\n")
	fmt.Printf("  GET  /api/search?q=query    - Search manga on mangakakalot\n")
	fmt.Printf("  POST /api/token             - Get/refresh Cloudflare token\n")
	fmt.Printf("  GET  /api/chapter/{id}      - Get chapter pages\n")
	fmt.Printf("  GET  /api/health            - Health check\n")

	return http.ListenAndServe(":"+s.port, middleware.CORS(mux))
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

func (s *Server) handleChapterPages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed, use GET")
		return
	}

	// Extract chapter ID from path: /api/chapter/{id}
	path := strings.TrimPrefix(r.URL.Path, "/api/chapter/")
	if path == "" {
		s.sendError(w, http.StatusBadRequest, "Chapter ID is required")
		return
	}

	var chapterID int
	if _, err := fmt.Sscanf(path, "%d", &chapterID); err != nil {
		s.sendError(w, http.StatusBadRequest, "Invalid chapter ID")
		return
	}

	// Get chapter info
	chapter, err := db.GetChapterByID(chapterID)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get chapter: "+err.Error())
		return
	}

	if chapter == nil {
		s.sendError(w, http.StatusNotFound, "Chapter not found")
		return
	}

	// Get pages
	pages, err := db.GetPagesByChapter(chapterID)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get pages: "+err.Error())
		return
	}

	response := map[string]interface{}{
		"success":    true,
		"chapter_id": chapterID,
		"chapter":    chapter,
		"pages":      pages,
		"count":      len(pages),
	}
	s.sendJSON(w, http.StatusOK, response)
}

func (s *Server) handleGetToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed, use POST")
		return
	}

	fmt.Println("API: Getting fresh Cloudflare token...")
	if err := scraper.GetCookies(); err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to get token: "+err.Error())
		return
	}

	cfToken := scraper.GetCfClearanceToken()
	if cfToken == "" {
		s.sendError(w, http.StatusServiceUnavailable, "Token not available. FlareSolverr may have failed.")
		return
	}

	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Token retrieved successfully",
		"token_preview": cfToken[:min(20, len(cfToken))] + "...",
	})
}

func (s *Server) handleSearchManga(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed, use GET")
		return
	}

	// Get search query
	query := r.URL.Query().Get("q")
	if query == "" {
		s.sendError(w, http.StatusBadRequest, "Search query is required (use ?q=searchterm)")
		return
	}

	// Get cf_clearance token (it should already be available from the /api/token call)
	cfToken := scraper.GetCfClearanceToken()
	if cfToken == "" {
		s.sendError(w, http.StatusServiceUnavailable, "Cloudflare token not available. Please call POST /api/token first.")
		return
	}

	// Create HTTP client with headers to mimic browser
	client := &http.Client{}
	mangakakalotURL := fmt.Sprintf("https://www.mangakakalot.gg/home/search/json?searchword=%s", query)
	
	req, err := http.NewRequest("GET", mangakakalotURL, nil)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to create request: "+err.Error())
		return
	}
	
	// Add browser-like headers with cf_clearance cookie
	req.Header.Set("User-Agent", scraper.GetCurrentUserAgent())
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Referer", "https://www.mangakakalot.gg/")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Cookie", "cf_clearance="+cfToken)
	
	resp, err := client.Do(req)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to search: "+err.Error())
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, "Failed to read response: "+err.Error())
		return
	}

	fmt.Printf("Search response status: %d, body length: %d\n", resp.StatusCode, len(body))
	fmt.Printf("Response preview: %s\n", string(body[:min(200, len(body))]))

	// Check status code
	if resp.StatusCode != http.StatusOK {
		s.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Search API returned status %d", resp.StatusCode))
		return
	}

	// Parse JSON to validate it
	var results interface{}
	if err := json.Unmarshal(body, &results); err != nil {
		fmt.Printf("JSON unmarshal error: %v\n", err)
		s.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Invalid response from search API: %v", err))
		return
	}

	// Return results
	s.sendJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    results,
		"query":   query,
	})
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

// min returns the smaller of a or b
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
