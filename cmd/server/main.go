package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/peczenyj/go-fizzbuzz/internal/api"
)

const (
	defaultShutdownTimeout   = 10 * time.Second
	defaultReadHeaderTimeout = 10 * time.Second
	defaultListenerAddress   = `:8080`
)

var (
	version  = "dev"
	revision = "unknown"
)

func main() {
	slog.Info("application start", slog.String("version", version), slog.String("revision", revision))

	os.Exit(MainWithExitCode(context.Background()))
}

// MainWithExitCode will run the server and prepare a context notification in case of signal to perform a graceful shutdown.
// returns the os exit code.
func MainWithExitCode(ctx context.Context) int {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := RunServer(ctx); err != nil {
		slog.Error("server finish with error", slog.Any("error", err))

		return 1
	}

	return 0
}

// RunServer start an http server and register the api handler.
func RunServer(ctx context.Context) error {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", api.Healthz)
	mux.HandleFunc("GET /fizzbuzz", api.FizzBuzz)

	server := &http.Server{
		Addr:              defaultListenerAddress,
		Handler:           mux,
		ReadHeaderTimeout: defaultReadHeaderTimeout,
	}

	done := make(chan error)

	go prepareServerShutdown(ctx, server, done)

	slog.Info("starting server, will listen and serve", slog.String("addr", server.Addr))

	// this will return immediately during graceful shutdown!
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("unexpected error from http listen and serve: %w", err)
	}

	err := <-done

	return err
}

func prepareServerShutdown(ctx context.Context, server *http.Server, done chan error) {
	defer close(done)

	<-ctx.Done()

	slog.Info("closing server", slog.Any("cause", context.Cause(ctx)))

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		done <- fmt.Errorf("http server shutdown error: %w", err)
	}
}
