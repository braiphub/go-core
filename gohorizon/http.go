package gohorizon

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// HTTPConfig defines HTTP server configuration
type HTTPConfig struct {
	Enabled  bool       `json:"enabled"`
	Addr     string     `json:"addr"`
	BasePath string     `json:"base_path"`
	Auth     AuthConfig `json:"auth"`
}

// AuthConfig defines authentication for dashboard
type AuthConfig struct {
	Enabled  bool   `json:"enabled"`
	Type     string `json:"type"` // "basic", "token"
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

// DefaultHTTPConfig returns sensible defaults
func DefaultHTTPConfig() HTTPConfig {
	return HTTPConfig{
		Enabled:  true,
		Addr:     ":8080",
		BasePath: "/horizon",
	}
}

// HTTPServer handles HTTP requests for the Horizon dashboard
type HTTPServer struct {
	config  HTTPConfig
	horizon *Horizon
	mux     *http.ServeMux
	server  *http.Server
}

// NewHTTPServer creates a new HTTP server
func NewHTTPServer(horizon *Horizon, config HTTPConfig) *HTTPServer {
	s := &HTTPServer{
		config:  config,
		horizon: horizon,
		mux:     http.NewServeMux(),
	}

	s.setupRoutes()
	return s
}

func (s *HTTPServer) setupRoutes() {
	base := s.config.BasePath
	if base == "" {
		base = "/horizon"
	}

	// API routes
	s.mux.HandleFunc(base+"/api/stats", s.withAuth(s.handleStats))
	s.mux.HandleFunc(base+"/api/queues", s.withAuth(s.handleQueues))
	s.mux.HandleFunc(base+"/api/workload", s.withAuth(s.handleWorkload))
	s.mux.HandleFunc(base+"/api/supervisors", s.withAuth(s.handleSupervisors))
	s.mux.HandleFunc(base+"/api/jobs/recent", s.withAuth(s.handleRecentJobs))
	s.mux.HandleFunc(base+"/api/jobs/failed", s.withAuth(s.handleFailedJobs))
	s.mux.HandleFunc(base+"/api/jobs/retry", s.withAuth(s.handleRetryJob))
	s.mux.HandleFunc(base+"/api/jobs/retry-all", s.withAuth(s.handleRetryAllJobs))
	s.mux.HandleFunc(base+"/api/jobs/flush", s.withAuth(s.handleFlushJobs))
	s.mux.HandleFunc(base+"/api/metrics/snapshots", s.withAuth(s.handleSnapshots))

	// Serve embedded UI dashboard
	uiFS, err := getUIFS()
	if err == nil {
		uiHandler := newSPAHandler(uiFS)
		s.mux.Handle(base+"/", s.withAuthHandler(http.StripPrefix(base, uiHandler)))
	}
}

// Start begins serving HTTP requests
func (s *HTTPServer) Start(ctx context.Context) error {
	s.server = &http.Server{
		Addr:    s.config.Addr,
		Handler: s.mux,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		return s.Stop(context.Background())
	case err := <-errCh:
		if err != http.ErrServerClosed {
			return err
		}
		return nil
	}
}

// Stop gracefully shuts down the server
func (s *HTTPServer) Stop(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// Handler returns the HTTP handler for embedding
func (s *HTTPServer) Handler() http.Handler {
	return s.mux
}

// Middleware

func (s *HTTPServer) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.config.Auth.Enabled {
			next(w, r)
			return
		}

		switch s.config.Auth.Type {
		case "basic":
			user, pass, ok := r.BasicAuth()
			if !ok || user != s.config.Auth.Username || pass != s.config.Auth.Password {
				w.Header().Set("WWW-Authenticate", `Basic realm="Horizon"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		case "token":
			token := r.Header.Get("Authorization")
			if token != "Bearer "+s.config.Auth.Token {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		next(w, r)
	}
}

func (s *HTTPServer) withAuthHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.config.Auth.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		switch s.config.Auth.Type {
		case "basic":
			user, pass, ok := r.BasicAuth()
			if !ok || user != s.config.Auth.Username || pass != s.config.Auth.Password {
				w.Header().Set("WWW-Authenticate", `Basic realm="Horizon"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		case "token":
			token := r.Header.Get("Authorization")
			if token != "Bearer "+s.config.Auth.Token {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// Response types

// StatsResponse contains overall statistics
type StatsResponse struct {
	Status         string          `json:"status"`
	JobsPerMinute  float64         `json:"jobs_per_minute"`
	TotalProcessed int64           `json:"total_processed"`
	TotalFailed    int64           `json:"total_failed"`
	TotalPending   int64           `json:"total_pending"`
	TotalWorkers   int             `json:"total_workers"`
	Queues         []*QueueMetrics `json:"queues"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// WorkloadResponse contains current workload by queue
type WorkloadResponse struct {
	Queues []QueueWorkload `json:"queues"`
}

// QueueWorkload represents workload for a queue
type QueueWorkload struct {
	Name    string `json:"name"`
	Length  int64  `json:"length"`
	Wait    string `json:"wait"`
	Workers int    `json:"workers"`
}

// SupervisorsResponse contains all supervisors
type SupervisorsResponse struct {
	Supervisors []SupervisorInfo `json:"supervisors"`
}

// SupervisorInfo contains supervisor details
type SupervisorInfo struct {
	Name    string   `json:"name"`
	Status  string   `json:"status"`
	Workers int      `json:"workers"`
	Queues  []string `json:"queues"`
	Balance string   `json:"balance"`
}

// Handlers

func (s *HTTPServer) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := s.horizon.metrics.GetStats(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Add worker count
	stats.TotalWorkers = 0
	for _, sup := range s.horizon.supervisors {
		stats.TotalWorkers += sup.WorkerCount()
	}

	s.writeJSON(w, stats)
}

func (s *HTTPServer) handleQueues(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics, err := s.horizon.metrics.GetAllQueuesMetrics(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.writeJSON(w, map[string]interface{}{
		"queues": metrics,
	})
}

func (s *HTTPServer) handleWorkload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics, err := s.horizon.metrics.GetAllQueuesMetrics(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	workload := WorkloadResponse{
		Queues: make([]QueueWorkload, 0, len(metrics)),
	}

	for _, m := range metrics {
		workload.Queues = append(workload.Queues, QueueWorkload{
			Name:   m.Queue,
			Length: m.PendingJobs + m.DelayedJobs,
			Wait:   m.WaitTime.String(),
		})
	}

	s.writeJSON(w, workload)
}

func (s *HTTPServer) handleSupervisors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := SupervisorsResponse{
		Supervisors: make([]SupervisorInfo, 0),
	}

	for name, sup := range s.horizon.supervisors {
		response.Supervisors = append(response.Supervisors, SupervisorInfo{
			Name:    name,
			Status:  string(sup.Status()),
			Workers: sup.WorkerCount(),
			Queues:  sup.Config().Queues,
			Balance: string(sup.Config().Balance),
		})
	}

	s.writeJSON(w, response)
}

func (s *HTTPServer) handleRecentJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := int64(50)
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 64); err == nil {
			limit = parsed
		}
	}

	jobs, err := s.horizon.metrics.GetRecentJobs(r.Context(), limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.writeJSON(w, map[string]interface{}{
		"jobs": jobs,
	})
}

func (s *HTTPServer) handleFailedJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := int64(50)
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 64); err == nil {
			limit = parsed
		}
	}

	jobs, err := s.horizon.failedStore.All(r.Context(), limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	count, _ := s.horizon.failedStore.Count(r.Context())

	s.writeJSON(w, map[string]interface{}{
		"jobs":        jobs,
		"total_count": count,
	})
}

func (s *HTTPServer) handleRetryJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.horizon.failedStore.Retry(r.Context(), req.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.writeJSON(w, map[string]interface{}{
		"success": true,
	})
}

func (s *HTTPServer) handleRetryAllJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	count, err := s.horizon.failedStore.RetryAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.writeJSON(w, map[string]interface{}{
		"success": true,
		"count":   count,
	})
}

func (s *HTTPServer) handleFlushJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := s.horizon.failedStore.Flush(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.writeJSON(w, map[string]interface{}{
		"success": true,
	})
}

func (s *HTTPServer) handleSnapshots(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Default to last hour
	to := time.Now()
	from := to.Add(-time.Hour)
	limit := int64(60)

	if f := r.URL.Query().Get("from"); f != "" {
		if ts, err := strconv.ParseInt(f, 10, 64); err == nil {
			from = time.Unix(ts, 0)
		}
	}
	if t := r.URL.Query().Get("to"); t != "" {
		if ts, err := strconv.ParseInt(t, 10, 64); err == nil {
			to = time.Unix(ts, 0)
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 64); err == nil {
			limit = parsed
		}
	}

	snapshots, err := s.horizon.metrics.GetSnapshots(r.Context(), from, to, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.writeJSON(w, map[string]interface{}{
		"snapshots": snapshots,
	})
}

func (s *HTTPServer) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
