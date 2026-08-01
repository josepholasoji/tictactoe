package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oyewunmi/tictactoe/internal/config"
	"github.com/oyewunmi/tictactoe/internal/db"
	"github.com/oyewunmi/tictactoe/internal/hub"
	"github.com/oyewunmi/tictactoe/internal/logging"
	"github.com/oyewunmi/tictactoe/internal/matchmaking"
	"github.com/oyewunmi/tictactoe/internal/session"
	"github.com/oyewunmi/tictactoe/internal/store"
	"github.com/oyewunmi/tictactoe/internal/ws"
)

const heartbeatSweepInterval = 10 * time.Second

// connectWithRetry will continue to retry until Postgres is ready, or until the context is canceled.
func connectWithRetry(ctx context.Context, dsn string, log *slog.Logger) (*db.Repository, error) {
	const maxAttempts = 15
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		repo, err := db.Connect(ctx, dsn)
		if err == nil {
			return repo, nil
		}
		lastErr = err
		log.Warn("postgres not ready, retrying", "attempt", attempt, "max_attempts", maxAttempts, "error", err)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return nil, lastErr
}

// handleHealth is a simple health check endpoint that returns 200 OK.
func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func runHeartbeatSweep(ctx context.Context, gateway *ws.Gateway) {
	ticker := time.NewTicker(heartbeatSweepInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			gateway.SweepStaleConnections()
		}
	}
}

func main() {
	cfg := config.Load()
	log := logging.New(cfg.LogLevel)
	slog.SetDefault(log)

	//
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	repo, err := connectWithRetry(ctx, cfg.PostgresDSN(), log)
	if err != nil {
		log.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer repo.Close()

	memStore := store.NewMemoryStore()
	connHub := hub.New()
	sessions := session.NewManager(memStore, connHub, repo, log)
	mm := matchmaking.NewService(memStore, connHub, sessions, log)
	gateway := ws.NewGateway(connHub, memStore, mm, sessions, repo, log)

	// start web/websocket server
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", gateway.HandleWS)
	mux.HandleFunc("/healthz", handleHealth)

	httpServer := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go runHeartbeatSweep(ctx, gateway)

	go func() {
		log.Info("server listening", "port", cfg.ServerPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
	}
}
