// Command server is the single deployable artifact: it serves the REST/OpenAPI
// API, the websocket endpoint, and the built frontend (embedded at build
// time) all from one binary.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/deranjer/loopira/internal/api"
	"github.com/deranjer/loopira/internal/db"
	"github.com/deranjer/loopira/internal/jobs"
	"github.com/deranjer/loopira/internal/webapp"
	"github.com/deranjer/loopira/internal/ws"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	databaseURL := getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/loopira?sslmode=disable")
	addr := getenv("ADDR", ":8080")

	if err := db.Migrate(databaseURL); err != nil {
		return err
	}
	slog.Info("migrations applied")

	pool, err := db.Connect(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	slog.Info("connected to database")

	if err := db.EnsureSeedData(ctx, db.New(pool)); err != nil {
		return err
	}

	hub := ws.NewHub()

	scheduler := jobs.NewScheduler()
	scheduler.Register(jobs.Job{
		Name:     "cycle-rollover",
		Interval: 24 * time.Hour,
		Run: func(ctx context.Context) error {
			// Rolls incomplete issues from ended cycles into the next cycle.
			// Implemented once cycles are built.
			return nil
		},
	})
	scheduler.Start(ctx)

	srv := api.New(pool, hub)

	frontend, err := webapp.Handler()
	if err != nil {
		return err
	}
	srv.Router.NotFound(frontend.ServeHTTP)

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           srv.Router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("shutdown error", "err", err)
		}
	}()

	slog.Info("listening", "addr", addr)
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
