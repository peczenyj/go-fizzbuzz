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
// and an error wrapping ErrInvalid for an invalid value. With -version, the
// other values are not validated.
//
// Every invalid environment variable is reported, even when a flag replaces
// it. Flag parsing stops at an invalid flag or at the first argument, and
// the flag values are then left unchecked, since the flags after that point
// could have replaced them; otherwise every invalid flag value is reported.
//
// MaxLimit and MaxStringLength are only parsed here: their bounds belong to
// fizzbuzz.NewGenerator.
func Parse(name string, args []string, getenv func(string) string, output io.Writer) (Config, error) {
	cfg := Default()

	envErr := cfg.readEnv(getenv)

	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(output)

	fs.StringVar(&cfg.Addr, "addr", cfg.Addr, "listen address (env "+EnvAddr+")")
	fs.TextVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "minimum log level: debug, info, warn or error (env "+EnvLogLevel+")")
	fs.StringVar(&cfg.LogFormat, "log-format", cfg.LogFormat, "log format: text or json (env "+EnvLogFormat+")")
	fs.IntVar(&cfg.MaxLimit, "max-limit", cfg.MaxLimit, "largest accepted limit, up to 100000 (env "+EnvMaxLimit+")")
	fs.IntVar(&cfg.MaxStringLength, "max-str-length", cfg.MaxStringLength, "largest accepted str1/str2 in bytes, up to 255 (env "+EnvMaxStringLength+")")
	fs.DurationVar(&cfg.ShutdownTimeout, "shutdown-timeout", cfg.ShutdownTimeout, "grace period for in-flight requests on shutdown (env "+EnvShutdownTimeout+")")
	fs.BoolVar(&cfg.ShowVersion, "version", false, "print the version and exit")

	report := func(err error) error {
		if err != nil {
			fmt.Fprintf(output, "%s: %v\n", name, err)
		}

		return err
	}

	var err error

	switch parseErr := fs.Parse(args); {
	case errors.Is(parseErr, flag.ErrHelp):
		return Config{}, parseErr
	case parseErr != nil:
		// flag has already reported parseErr and the usage.
		return Config{}, errors.Join(fmt.Errorf("%w: %w", ErrInvalid, parseErr), report(envErr))
	case cfg.ShowVersion:
		return cfg, nil
	case fs.NArg() > 0:
		err = fmt.Errorf("%w: unexpected arguments %q", ErrInvalid, fs.Args())
	default:
		err = cfg.validate()
	}

	if err := report(errors.Join(envErr, err)); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// readEnv overrides the defaults with the environment variables that are set.
// It checks each value as validate does, and reports every invalid one; an
// invalid value leaves the built-in default in place.
func (c *Config) readEnv(getenv func(string) string) error {
	var errs []error

	parse := func(name string, set func(string) error) {
		if v := getenv(name); v != "" {
			if err := set(v); err != nil {
				errs = append(errs, fmt.Errorf("%w: %s=%q: %w", ErrInvalid, name, v, err))
			}
		}
	}

	parse(EnvAddr, setIfValid(&c.Addr, checked(asString, checkAddr)))
	parse(EnvLogLevel, setIfValid(&c.LogLevel, parseLevel))
	parse(EnvLogFormat, setIfValid(&c.LogFormat, checked(asString, checkLogFormat)))
	parse(EnvMaxLimit, setIfValid(&c.MaxLimit, parseInt))
	parse(EnvMaxStringLength, setIfValid(&c.MaxStringLength, parseInt))
	parse(EnvShutdownTimeout, setIfValid(&c.ShutdownTimeout, checked(time.ParseDuration, checkShutdownTimeout)))

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

// checked returns parse followed by check, for a value that must parse and
// then pass check.
func checked[T any](parse func(string) (T, error), check func(T) error) func(string) (T, error) {
	return func(v string) (T, error) {
		x, err := parse(v)
		if err == nil {
			err = check(x)
		}

		return x, err
	}
}

func asString(v string) (string, error) { return v, nil }

// parseLevel parses a log level name, such as "debug" or "warn".
func parseLevel(v string) (level slog.Level, err error) {
	err = level.UnmarshalText([]byte(v))

	return level, err
}

// parseInt parses an integer with the same syntax as flag.IntVar: base
// prefixes (0x, 0o, 0b), leading-zero octal and underscores.
func parseInt(v string) (int, error) {
	n, err := strconv.ParseInt(v, 0, strconv.IntSize)

	return int(n), err
}

// checkAddr checks a host:port address, and that its port could be listened
// on: a number up to 65535 or a known service name.
func checkAddr(addr string) error {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}

	_, err = net.LookupPort("tcp", port)

	return err
}

func checkLogFormat(format string) error {
	if format != LogFormatText && format != LogFormatJSON {
		return fmt.Errorf("must be %q or %q", LogFormatText, LogFormatJSON)
	}

	return nil
}

func checkShutdownTimeout(timeout time.Duration) error {
	if timeout <= 0 {
		return errors.New("must be positive")
	}

	return nil
}

// validate checks the values that parsing alone cannot. It reports every
// invalid value, not just the first one.
func (c *Config) validate() error {
	var errs []error

	if err := checkAddr(c.Addr); err != nil {
		errs = append(errs, fmt.Errorf("%w: addr %q: %w", ErrInvalid, c.Addr, err))
	}

	if err := checkLogFormat(c.LogFormat); err != nil {
		errs = append(errs, fmt.Errorf("%w: log format %q: %w", ErrInvalid, c.LogFormat, err))
	}

	if err := checkShutdownTimeout(c.ShutdownTimeout); err != nil {
		errs = append(errs, fmt.Errorf("%w: shutdown timeout %v: %w", ErrInvalid, c.ShutdownTimeout, err))
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
