package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/peczenyj/go-fizzbuzz/internal/api"
	"github.com/peczenyj/go-fizzbuzz/internal/config"
	"github.com/peczenyj/go-fizzbuzz/internal/fizzbuzz"
	"github.com/peczenyj/go-fizzbuzz/internal/stats"
)

const (
	programName = "fizzbuzz"

	defaultReadHeaderTimeout = 5 * time.Second
	defaultReadTimeout       = 10 * time.Second
	defaultWriteTimeout      = 10 * time.Second
	defaultIdleTimeout       = 120 * time.Second
	defaultMaxHeaderBytes    = 16 << 10 // 16 KiB
)

var (
	version  = "dev"
	revision = "unknown"
)

// Exit codes returned by run.
const (
	exitOK    = 0
	exitError = 1 // the server failed
	exitUsage = 2 // invalid flags or environment, as for flag errors
)

func main() {
	os.Exit(run(context.Background(), os.Args[1:], os.Getenv, os.Stdout, os.Stderr))
}

// run reads the configuration, then serves until ctx is cancelled or SIGINT
// or SIGTERM arrives. It returns the process exit code.
func run(ctx context.Context, args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	cfg, err := config.Parse(programName, args, getenv, stderr)

	switch {
	case errors.Is(err, flag.ErrHelp):
		return exitOK
	case err != nil:
		return exitUsage // already reported by config.Parse
	case cfg.ShowVersion:
		fmt.Fprintf(stdout, "go-fizzbuzz %s (revision %s)\n", version, revision)

		return exitOK
	}

	logger := cfg.NewLogger(stderr)
	slog.SetDefault(logger)

	gen, err := fizzbuzz.NewGenerator(cfg.MaxLimit, cfg.MaxStringLength)
	if err != nil {
		logger.Error("invalid configuration", slog.Any("error", err))

		return exitUsage
	}

	logger.Info("application start",
		slog.String("version", version),
		slog.String("revision", revision),
		slog.String("log_level", cfg.LogLevel.String()),
		slog.String("log_format", cfg.LogFormat),
		slog.Int("max_limit", cfg.MaxLimit),
		slog.Int("max_str_length", cfg.MaxStringLength),
	)

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	handler := api.New(gen, stats.NewCounter[fizzbuzz.Params]())

	if err := RunServer(ctx, cfg, handler, logger); err != nil {
		logger.Error("server finish with error", slog.Any("error", err))

		return exitError
	}

	return exitOK
}

var errServerStopped = errors.New("server stopped")

// RunServer serves handler on cfg.Addr until ctx is cancelled, then shuts
// down gracefully within cfg.ShutdownTimeout.
func RunServer(ctx context.Context, cfg config.Config, handler http.Handler, logger *slog.Logger) error {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(errServerStopped)

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler,
		ReadHeaderTimeout: defaultReadHeaderTimeout,
		ReadTimeout:       defaultReadTimeout,
		WriteTimeout:      defaultWriteTimeout,
		IdleTimeout:       defaultIdleTimeout,
		MaxHeaderBytes:    defaultMaxHeaderBytes,
	}

	done := make(chan error, 1)

	go prepareServerShutdown(ctx, server, cfg.ShutdownTimeout, logger, done)

	logger.Info("starting server, will listen and serve", slog.String("addr", server.Addr))

	// this will return immediately during graceful shutdown!
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("unexpected error from http listen and serve: %w", err)
	}

	err := <-done

	return err
}

func prepareServerShutdown(ctx context.Context, server *http.Server, timeout time.Duration, logger *slog.Logger, done chan error) {
	defer close(done)

	<-ctx.Done()

	if cause := context.Cause(ctx); !errors.Is(cause, errServerStopped) {
		logger.Info("closing server", slog.Any("cause", cause))
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), timeout)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		done <- fmt.Errorf("http server shutdown error: %w", err)
	}
}
