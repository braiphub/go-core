package gohorizon

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed ui/dist/*
var uiFS embed.FS

// getUIFS returns the embedded UI filesystem
func getUIFS() (http.FileSystem, error) {
	subFS, err := fs.Sub(uiFS, "ui/dist")
	if err != nil {
		return nil, err
	}
	return http.FS(subFS), nil
}

// spaHandler serves the SPA with proper fallback to index.html
type spaHandler struct {
	handler http.Handler
	fs      http.FileSystem
}

func newSPAHandler(fs http.FileSystem) *spaHandler {
	return &spaHandler{
		handler: http.FileServer(fs),
		fs:      fs,
	}
}

func (h *spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Check if the path starts with /api - don't handle API routes here
	if strings.HasPrefix(path, "/api") {
		http.NotFound(w, r)
		return
	}

	// Try to open the file
	f, err := h.fs.Open(path)
	if err != nil {
		// File doesn't exist, serve index.html for SPA routing
		r.URL.Path = "/"
	} else {
		f.Close()
	}

	h.handler.ServeHTTP(w, r)
}
