package config_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/peczenyj/go-fizzbuzz/internal/config"
)

// env returns a getenv function backed by a map.
func env(vars map[string]string) func(string) string {
	return func(name string) string { return vars[name] }
}

func TestParse(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		label    string
		args     []string
		env      map[string]string
		expected func(*config.Config) // changes applied to config.Default()
	}{
		{
			label:    "should return the defaults without flags or environment",
			expected: func(*config.Config) {},
		},
		{
			label: "should read every setting from the environment",
			env: map[string]string{
				config.EnvAddr:            "127.0.0.1:9090",
				config.EnvLogLevel:        "debug",
				config.EnvLogFormat:       "json",
				config.EnvMaxLimit:        "100",
				config.EnvMaxStringLength: "8",
				config.EnvShutdownTimeout: "3s",
			},
			expected: func(c *config.Config) {
				c.Addr = "127.0.0.1:9090"
				c.LogLevel = slog.LevelDebug
				c.LogFormat = config.LogFormatJSON
				c.MaxLimit = 100
				c.MaxStringLength = 8
				c.ShutdownTimeout = 3 * time.Second
			},
		},
		{
			label: "should read every setting from the flags",
			args: []string{
				"-addr=:9090", "-log-level=warn", "-log-format=json",
				"-max-limit=2048", "-max-str-length=128", "-shutdown-timeout=1m",
			},
			expected: func(c *config.Config) {
				c.Addr = ":9090"
				c.LogLevel = slog.LevelWarn
				c.LogFormat = config.LogFormatJSON
				c.MaxLimit = 2048
				c.MaxStringLength = 128
				c.ShutdownTimeout = time.Minute
			},
		},
		{
			label: "should prefer a flag over the environment",
			args:  []string{"-log-level=error", "-addr=:7070"},
			env:   map[string]string{config.EnvLogLevel: "debug", config.EnvAddr: ":9090"},
			expected: func(c *config.Config) {
				c.LogLevel = slog.LevelError
				c.Addr = ":7070"
			},
		},
		{
			// the same syntax as the flags: base prefixes, leading-zero octal and underscores
			label: "should read integers from the environment like the flags",
			env:   map[string]string{config.EnvMaxLimit: "0x400", config.EnvMaxStringLength: "010"},
			expected: func(c *config.Config) {
				c.MaxLimit = 1024
				c.MaxStringLength = 8
			},
		},
		{
			label:    "should accept an address with port 0",
			args:     []string{"-addr=localhost:0"},
			expected: func(c *config.Config) { c.Addr = "localhost:0" },
		},
		{
			label:    "should accept a case-insensitive log level",
			env:      map[string]string{config.EnvLogLevel: "WARN"},
			expected: func(c *config.Config) { c.LogLevel = slog.LevelWarn },
		},
		{
			label:    "should set ShowVersion with -version",
			args:     []string{"-version"},
			expected: func(c *config.Config) { c.ShowVersion = true },
		},
		{
			label:    "should skip validation with -version",
			args:     []string{"-version", "-log-format=xml"},
			expected: func(c *config.Config) { c.ShowVersion = true; c.LogFormat = "xml" },
		},
		{
			label:    "should ignore an invalid environment with -version",
			args:     []string{"-version"},
			env:      map[string]string{config.EnvMaxLimit: "many", config.EnvLogFormat: "xml"},
			expected: func(c *config.Config) { c.ShowVersion = true },
		},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			t.Parallel()

			var output bytes.Buffer

			got, err := config.Parse("fizzbuzz", tc.args, env(tc.env), &output)
			if err != nil {
				t.Fatalf("unexpected error: %v (output: %q)", err, output.String())
			}

			expected := config.Default()
			tc.expected(&expected)

			if got != expected {
				t.Fatalf("unexpected config (got: %+v, expected: %+v)", got, expected)
			}

			if output.Len() != 0 {
				t.Fatalf("unexpected output: %q", output.String())
			}
		})
	}
}

func TestParse_invalid(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		label  string
		args   []string
		env    map[string]string
		output string // expected in the reported error
	}{
		{label: "unknown flag", args: []string{"-port=80"}, output: "flag provided but not defined: -port"},
		{label: "unknown flag with -version", args: []string{"-version", "-port=80"}, output: "flag provided but not defined: -port"},
		{label: "positional argument", args: []string{"serve"}, output: `unexpected arguments ["serve"]`},
		{label: "invalid log level flag", args: []string{"-log-level=verbose"}, output: `invalid value "verbose" for flag -log-level`},
		{label: "invalid log format flag", args: []string{"-log-format=xml"}, output: `log format "xml"`},
		{label: "invalid integer flag", args: []string{"-max-limit=many"}, output: `invalid value "many" for flag -max-limit`},
		{label: "invalid duration flag", args: []string{"-shutdown-timeout=10"}, output: `invalid value "10" for flag -shutdown-timeout`},
		{label: "zero shutdown timeout", args: []string{"-shutdown-timeout=0s"}, output: "shutdown timeout 0s: must be positive"},
		{label: "address without port", args: []string{"-addr=localhost"}, output: `addr "localhost"`},
		{label: "port out of range", args: []string{"-addr=:99999"}, output: `addr ":99999"`},
		{label: "unknown port", args: []string{"-addr=:nope"}, output: `addr ":nope"`},
		{label: "port out of range env", env: map[string]string{config.EnvAddr: ":65536"}, output: `FIZZBUZZ_ADDR=":65536"`},
		{label: "invalid octal integer env", env: map[string]string{config.EnvMaxLimit: "08"}, output: `FIZZBUZZ_MAX_LIMIT="08"`},
		{label: "invalid log level env", env: map[string]string{config.EnvLogLevel: "verbose"}, output: `FIZZBUZZ_LOG_LEVEL="verbose"`},
		{label: "invalid log format env", env: map[string]string{config.EnvLogFormat: "xml"}, output: `FIZZBUZZ_LOG_FORMAT="xml": must be "text" or "json"`},
		{label: "address without port env", env: map[string]string{config.EnvAddr: "localhost"}, output: `FIZZBUZZ_ADDR="localhost"`},
		{label: "zero shutdown timeout env", env: map[string]string{config.EnvShutdownTimeout: "0s"}, output: `FIZZBUZZ_SHUTDOWN_TIMEOUT="0s": must be positive`},
		{label: "invalid integer env", env: map[string]string{config.EnvMaxLimit: "many"}, output: `FIZZBUZZ_MAX_LIMIT="many"`},
		{label: "invalid duration env", env: map[string]string{config.EnvShutdownTimeout: "soon"}, output: `FIZZBUZZ_SHUTDOWN_TIMEOUT="soon"`},
		{
			label:  "invalid env is not hidden by a valid flag",
			args:   []string{"-max-limit=10"},
			env:    map[string]string{config.EnvMaxLimit: "many"},
			output: `FIZZBUZZ_MAX_LIMIT="many"`,
		},
		{
			label:  "invalid env value is not hidden by a valid flag",
			args:   []string{"-log-format=json"},
			env:    map[string]string{config.EnvLogFormat: "xml"},
			output: `FIZZBUZZ_LOG_FORMAT="xml"`,
		},
	}

	for _, tc := range testcases {
		t.Run("should reject "+tc.label, func(t *testing.T) {
			t.Parallel()

			var output bytes.Buffer

			_, err := config.Parse("fizzbuzz", tc.args, env(tc.env), &output)
			if !errors.Is(err, config.ErrInvalid) {
				t.Fatalf("unexpected error (got: %v, expected: %v)", err, config.ErrInvalid)
			}

			if !strings.Contains(output.String(), tc.output) {
				t.Fatalf("error not reported (got: %q, expected to contain: %q)", output.String(), tc.output)
			}
		})
	}
}

func TestParse_reports(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		label    string
		args     []string
		env      map[string]string
		expected []string // each reported once, and in the returned error
		unwanted []string // not reported
	}{
		{
			label:    "should report every invalid env",
			env:      map[string]string{config.EnvMaxLimit: "many", config.EnvShutdownTimeout: "soon"},
			expected: []string{`FIZZBUZZ_MAX_LIMIT="many"`, `FIZZBUZZ_SHUTDOWN_TIMEOUT="soon"`},
		},
		{
			label: "should report every error",
			args:  []string{"-addr=localhost", "-shutdown-timeout=0s", "-log-format=yaml"},
			env:   map[string]string{config.EnvMaxLimit: "many", config.EnvLogFormat: "xml"},
			expected: []string{
				`FIZZBUZZ_MAX_LIMIT="many"`,
				`FIZZBUZZ_LOG_FORMAT="xml"`,
				`addr "localhost"`,
				`log format "yaml"`,
				"shutdown timeout 0s: must be positive",
			},
		},
		{
			label: "should report the invalid env with an invalid flag",
			args:  []string{"-port=80"},
			env:   map[string]string{config.EnvMaxLimit: "many", config.EnvLogFormat: "xml"},
			expected: []string{
				"flag provided but not defined: -port",
				`FIZZBUZZ_MAX_LIMIT="many"`,
				`FIZZBUZZ_LOG_FORMAT="xml"`,
			},
		},
		{
			// the flags after the invalid one are not parsed, so a flag
			// before it may still be replaced
			label:    "should not validate the flags before an invalid flag",
			args:     []string{"-addr=localhost", "-port=80", "-addr=:9090"},
			expected: []string{"flag provided but not defined: -port"},
			unwanted: []string{`addr "localhost"`},
		},
		{
			// "-addr=:9090" after "serve" is an argument, not a flag
			label:    "should not validate the flags before unexpected arguments",
			args:     []string{"-addr=localhost", "serve", "-addr=:9090"},
			env:      map[string]string{config.EnvLogFormat: "xml"},
			expected: []string{`unexpected arguments ["serve" "-addr=:9090"]`, `FIZZBUZZ_LOG_FORMAT="xml"`},
			unwanted: []string{`addr "localhost"`},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			t.Parallel()

			var output bytes.Buffer

			_, err := config.Parse("fizzbuzz", tc.args, env(tc.env), &output)
			if !errors.Is(err, config.ErrInvalid) {
				t.Fatalf("unexpected error (got: %v, expected: %v)", err, config.ErrInvalid)
			}

			for _, want := range tc.expected {
				if n := strings.Count(output.String(), want); n != 1 {
					t.Fatalf("error reported %d times, expected once (got: %q, expected: %q)", n, output.String(), want)
				}

				if !strings.Contains(err.Error(), want) {
					t.Fatalf("error not returned (got: %q, expected to contain: %q)", err.Error(), want)
				}
			}

			for _, unwanted := range tc.unwanted {
				if strings.Contains(output.String(), unwanted) {
					t.Fatalf("unexpected error reported (got: %q, not expected: %q)", output.String(), unwanted)
				}
			}
		})
	}
}

func TestParse_help(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	_, err := config.Parse("fizzbuzz", []string{"-h"}, env(nil), &output)
	if !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("unexpected error (got: %v, expected: %v)", err, flag.ErrHelp)
	}

	// the usage names every flag and its environment variable
	for _, want := range []string{"-addr", config.EnvAddr, "-log-level", "-version"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("usage does not mention %s: %q", want, output.String())
		}
	}
}

func TestParse_help_keeps_defaults_with_invalid_env(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	_, err := config.Parse("fizzbuzz", []string{"-h"}, env(map[string]string{
		config.EnvAddr:            "localhost",
		config.EnvLogLevel:        "verbose",
		config.EnvLogFormat:       "xml",
		config.EnvMaxLimit:        "99999999999999999999", // out of range: Atoi returns math.MaxInt with its error
		config.EnvMaxStringLength: "-99999999999999999999",
		config.EnvShutdownTimeout: "-5s", // parses, but must be positive
	}), &output)
	if !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("unexpected error (got: %v, expected: %v)", err, flag.ErrHelp)
	}

	// an invalid value must not replace the built-in default shown in the usage
	for _, want := range []string{
		`(default ":8080")`, "(default INFO)", `(default "text")`, "(default 1024)", "(default 64)", "(default 10s)",
	} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("usage does not mention %s: %q", want, output.String())
		}
	}
}

func TestConfig_NewLogger(t *testing.T) {
	t.Parallel()

	t.Run("should write JSON and filter by level", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		cfg := config.Default()
		cfg.LogFormat = config.LogFormatJSON
		cfg.LogLevel = slog.LevelWarn

		logger := cfg.NewLogger(&buf)
		logger.Info("hidden")
		logger.Warn("shown", slog.Int("n", 1))

		var record map[string]any
		if err := json.Unmarshal(buf.Bytes(), &record); err != nil {
			t.Fatalf("output is not a single JSON record: %v (%q)", err, buf.String())
		}

		if record["msg"] != "shown" || record["level"] != "WARN" {
			t.Fatalf("unexpected record: %v", record)
		}
	})

	t.Run("should write text by default", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		config.Default().NewLogger(&buf).Info("hello")

		if !strings.Contains(buf.String(), "level=INFO msg=hello") {
			t.Fatalf("unexpected text output: %q", buf.String())
		}
	})
}
