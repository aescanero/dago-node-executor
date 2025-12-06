package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// HealthServer provides health check endpoints
type HealthServer struct {
	worker *Worker
	logger *zap.Logger
	server *http.Server
}

// NewHealthServer creates a new health server
func NewHealthServer(worker *Worker, port int, logger *zap.Logger) *HealthServer {
	mux := http.NewServeMux()

	hs := &HealthServer{
		worker: worker,
		logger: logger,
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
	}

	mux.HandleFunc("/health", hs.handleHealth)
	mux.HandleFunc("/ready", hs.handleReady)

	return hs
}

// Start starts the health server
func (hs *HealthServer) Start() error {
	hs.logger.Info("starting health server", zap.String("addr", hs.server.Addr))

	go func() {
		if err := hs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			hs.logger.Error("health server error", zap.Error(err))
		}
	}()

	return nil
}

// Stop stops the health server
func (hs *HealthServer) Stop(ctx context.Context) error {
	return hs.server.Shutdown(ctx)
}

// handleHealth handles health check requests
func (hs *HealthServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	isHealthy := hs.worker.IsHealthy()

	status := "healthy"
	statusCode := http.StatusOK

	if !isHealthy {
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	response := map[string]interface{}{
		"status":         status,
		"worker_id":      hs.worker.id,
		"last_processed": hs.worker.GetLastProcessed(),
		"timestamp":      time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(response)
}

// handleReady handles readiness check requests
func (hs *HealthServer) handleReady(w http.ResponseWriter, r *http.Request) {
	// Worker is ready if it's healthy
	if hs.worker.IsHealthy() {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("not ready"))
	}
}
