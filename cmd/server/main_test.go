package main

import (
	"bytes"
	"context"
	"net"
	"strings"
	"testing"
)

func noEnv(string) string { return "" }

func TestRun(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		label    string
		args     []string
		exitCode int
		stdout   string // expected prefix
		stderr   string // expected substring
	}{
		{label: "should print the version and exit", args: []string{"-version"}, exitCode: exitOK, stdout: "go-fizzbuzz dev (revision unknown)\n"},
		{label: "should print the usage on -h", args: []string{"-h"}, exitCode: exitOK, stderr: "Usage of fizzbuzz"},
		{label: "should reject an unknown flag", args: []string{"-port=80"}, exitCode: exitUsage, stderr: "flag provided but not defined"},
		{label: "should reject limits above the hard ceiling", args: []string{"-max-limit=100001"}, exitCode: exitUsage, stderr: "invalid max limit 100001"},
		{label: "should reject a zero string length", args: []string{"-max-str-length=0"}, exitCode: exitUsage, stderr: "invalid max string length 0"},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			if got := run(context.Background(), tc.args, noEnv, &stdout, &stderr); got != tc.exitCode {
				t.Fatalf("unexpected exit code (got: %d, expected: %d, stderr: %q)", got, tc.exitCode, stderr.String())
			}

			if !strings.HasPrefix(stdout.String(), tc.stdout) {
				t.Fatalf("unexpected stdout (got: %q, expected prefix: %q)", stdout.String(), tc.stdout)
			}

			if !strings.Contains(stderr.String(), tc.stderr) {
				t.Fatalf("unexpected stderr (got: %q, expected to contain: %q)", stderr.String(), tc.stderr)
			}
		})
	}
}

// TestRun_serves_until_cancelled checks the whole wiring: configuration,
// logger, generator, server start and graceful shutdown.
func TestRun_serves_until_cancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // the server shuts down as soon as it starts

	var stdout, stderr bytes.Buffer

	// port 0: any free port, so parallel test runs don't collide
	if got := run(ctx, []string{"-addr=127.0.0.1:0", "-log-format=json"}, noEnv, &stdout, &stderr); got != exitOK {
		t.Fatalf("unexpected exit code (got: %d, expected: %d, stderr: %q)", got, exitOK, stderr.String())
	}

	if !strings.Contains(stderr.String(), `"msg":"application start"`) {
		t.Fatalf("missing JSON start log in %q", stderr.String())
	}
}

func TestRun_fails_if_the_address_is_in_use(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("unexpected error while listening: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	var stdout, stderr bytes.Buffer

	args := []string{"-addr=" + listener.Addr().String()}
	if got := run(context.Background(), args, noEnv, &stdout, &stderr); got != exitError {
		t.Fatalf("unexpected exit code (got: %d, expected: %d, stderr: %q)", got, exitError, stderr.String())
	}

	if !strings.Contains(stderr.String(), "address already in use") {
		t.Fatalf("missing error in %q", stderr.String())
	}
}
