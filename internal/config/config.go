// Package config reads the server settings from command-line flags, with
// FIZZBUZZ_* environment variables as defaults: flag > env var > built-in default.
package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strconv"
	"time"

	"github.com/peczenyj/go-fizzbuzz/internal/fizzbuzz"
)

// Environment variables read by Parse.
const (
	EnvAddr            = "FIZZBUZZ_ADDR"
	EnvLogLevel        = "FIZZBUZZ_LOG_LEVEL"
	EnvLogFormat       = "FIZZBUZZ_LOG_FORMAT"
	EnvMaxLimit        = "FIZZBUZZ_MAX_LIMIT"
	EnvMaxStringLength = "FIZZBUZZ_MAX_STR_LENGTH"
	EnvShutdownTimeout = "FIZZBUZZ_SHUTDOWN_TIMEOUT"
)

// Log formats accepted by Parse.
const (
	LogFormatText = "text"
	LogFormatJSON = "json"
)

// ErrInvalid is returned, wrapped, for an invalid flag or environment value.
var ErrInvalid = errors.New("invalid configuration")

// Config holds the server settings.
type Config struct {
	Addr            string        // listen address, e.g. ":8080"
	LogLevel        slog.Level    // minimum level logged
	LogFormat       string        // LogFormatText or LogFormatJSON
	MaxLimit        int           // largest accepted limit
	MaxStringLength int           // largest accepted str1/str2, in bytes
	ShutdownTimeout time.Duration // grace period for in-flight requests
	ShowVersion     bool          // print the version and exit
}

// Default returns the built-in settings.
func Default() Config {
	return Config{
		Addr:            ":8080",
		LogLevel:        slog.LevelInfo,
		LogFormat:       LogFormatText,
		MaxLimit:        fizzbuzz.DefaultMaxLimit,
		MaxStringLength: fizzbuzz.DefaultMaxStringLength,
		ShutdownTimeout: 10 * time.Second,
	}
}

// Parse reads args (without the program name), using getenv for defaults.
// Errors and usage are written to output. It returns flag.ErrHelp for -h,
// and an error wrapping ErrInvalid for an invalid value; every invalid value
// is reported, not just the first one. With -version, the other values are
// not validated.
//
// MaxLimit and MaxStringLength are only parsed here: their bounds belong to
// fizzbuzz.NewGenerator.
func Parse(name string, args []string, getenv func(string) string, output io.Writer) (Config, error) {
	cfg := Default()

	errs := []error{cfg.readEnv(getenv)}

	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(output)

	fs.StringVar(&cfg.Addr, "addr", cfg.Addr, "listen address (env "+EnvAddr+")")
	fs.TextVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "minimum log level: debug, info, warn or error (env "+EnvLogLevel+")")
	fs.StringVar(&cfg.LogFormat, "log-format", cfg.LogFormat, "log format: text or json (env "+EnvLogFormat+")")
	fs.IntVar(&cfg.MaxLimit, "max-limit", cfg.MaxLimit, "largest accepted limit, up to 100000 (env "+EnvMaxLimit+")")
	fs.IntVar(&cfg.MaxStringLength, "max-str-length", cfg.MaxStringLength, "largest accepted str1/str2 in bytes, up to 255 (env "+EnvMaxStringLength+")")
	fs.DurationVar(&cfg.ShutdownTimeout, "shutdown-timeout", cfg.ShutdownTimeout, "grace period for in-flight requests on shutdown (env "+EnvShutdownTimeout+")")
	fs.BoolVar(&cfg.ShowVersion, "version", false, "print the version and exit")

	// flag reports its own error and the usage, so flagErr is only returned.
	var flagErr error

	switch err := fs.Parse(args); {
	case errors.Is(err, flag.ErrHelp):
		return Config{}, err
	case err != nil:
		// The flags after the invalid one are not parsed: validating cfg
		// could reject values those flags would have replaced.
		flagErr = fmt.Errorf("%w: %w", ErrInvalid, err)
	case cfg.ShowVersion:
		return cfg, nil
	default:
		if fs.NArg() > 0 {
			errs = append(errs, fmt.Errorf("%w: unexpected arguments %q", ErrInvalid, fs.Args()))
		}

		errs = append(errs, cfg.validate())
	}

	err := errors.Join(errs...)
	if err != nil {
		fmt.Fprintf(output, "%s: %v\n", name, err)
	}

	if err = errors.Join(flagErr, err); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// readEnv overrides the defaults with the environment variables that are set.
// It reports every invalid value, not just the first one.
func (c *Config) readEnv(getenv func(string) string) error {
	var errs []error

	parse := func(name string, set func(string) error) {
		if v := getenv(name); v != "" {
			if err := set(v); err != nil {
				errs = append(errs, fmt.Errorf("%w: %s=%q: %w", ErrInvalid, name, v, err))
			}
		}
	}

	parse(EnvAddr, func(v string) error { c.Addr = v; return nil })
	parse(EnvLogLevel, setIfValid(&c.LogLevel, parseLevel))
	parse(EnvLogFormat, func(v string) error { c.LogFormat = v; return nil })
	parse(EnvMaxLimit, setIfValid(&c.MaxLimit, strconv.Atoi))
	parse(EnvMaxStringLength, setIfValid(&c.MaxStringLength, strconv.Atoi))
	parse(EnvShutdownTimeout, setIfValid(&c.ShutdownTimeout, time.ParseDuration))

	return errors.Join(errs...)
}

// setIfValid returns a setter that stores the parsed value in dst only on success,
// so an invalid value leaves the built-in default in place.
func setIfValid[T any](dst *T, parse func(string) (T, error)) func(string) error {
	return func(v string) error {
		x, err := parse(v)
		if err == nil {
			*dst = x
		}

		return err
	}
}

// parseLevel parses a log level name, such as "debug" or "warn".
func parseLevel(v string) (level slog.Level, err error) {
	err = level.UnmarshalText([]byte(v))

	return level, err
}

// validate checks the values that parsing alone cannot. It reports every
// invalid value, not just the first one.
func (c *Config) validate() error {
	var errs []error

	if _, _, err := net.SplitHostPort(c.Addr); err != nil {
		errs = append(errs, fmt.Errorf("%w: addr %q: %w", ErrInvalid, c.Addr, err))
	}

	if c.LogFormat != LogFormatText && c.LogFormat != LogFormatJSON {
		errs = append(errs, fmt.Errorf("%w: log format %q: must be %q or %q", ErrInvalid, c.LogFormat, LogFormatText, LogFormatJSON))
	}

	if c.ShutdownTimeout <= 0 {
		errs = append(errs, fmt.Errorf("%w: shutdown timeout %v: must be positive", ErrInvalid, c.ShutdownTimeout))
	}

	return errors.Join(errs...)
}

// NewLogger returns a logger writing to w in the configured format and level.
func (c Config) NewLogger(w io.Writer) *slog.Logger {
	opts := &slog.HandlerOptions{Level: c.LogLevel}

	if c.LogFormat == LogFormatJSON {
		return slog.New(slog.NewJSONHandler(w, opts))
	}

	return slog.New(slog.NewTextHandler(w, opts))
}
