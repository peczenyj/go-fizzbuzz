#!/usr/bin/env bash
# Load test for a running go-fizzbuzz server, using hey
# (https://github.com/rakyll/hey; `apt install hey` or `go install github.com/rakyll/hey@latest`).
#
# Usage:  scripts/loadtest.sh [base-url]           (default http://localhost:8080)
# Env:    DURATION (default 30s), CONCURRENCY (default 50)
#
# Prints a Markdown table ready to paste into the README.
set -euo pipefail

BASE_URL="${1:-http://localhost:8080}"
DURATION="${DURATION:-30s}"
CONCURRENCY="${CONCURRENCY:-50}"

command -v hey >/dev/null || { echo "hey not found: apt install hey" >&2; exit 1; }
curl -fsS -o /dev/null "$BASE_URL/healthz" || { echo "server not reachable at $BASE_URL" >&2; exit 1; }

# Worst case: max limit, int1=int2=1 (every element is str1str2), and 64-byte
# strings that JSON-escape 6× ("<" → "<"): about 790 KB per response.
worst=$(printf '%%3C%.0s' {1..64})

scenarios=(
  "healthz|/healthz"
  "classic 1..100|/fizzbuzz?int1=3&int2=5&limit=100&str1=fizz&str2=buzz"
  "max limit 1024|/fizzbuzz?int1=3&int2=5&limit=1024&str1=fizz&str2=buzz"
  "worst case (~790 KB)|/fizzbuzz?int1=1&int2=1&limit=1024&str1=$worst&str2=$worst"
  "invalid (400)|/fizzbuzz?int1=0&int2=5&limit=100&str1=a&str2=b"
)

echo "hey -z $DURATION -c $CONCURRENCY against $BASE_URL"
echo "$(nproc) CPUs, $(grep -m1 'model name' /proc/cpuinfo | cut -d: -f2 | xargs), $(date -u +%Y-%m-%d)"
echo
echo "| Scenario | Req/s | p50 | p95 | p99 | Status |"
echo "|---|---:|---:|---:|---:|---|"

for s in "${scenarios[@]}"; do
  name="${s%%|*}"
  path="${s#*|}"
  echo "running: $name" >&2

  hey -z "$DURATION" -c "$CONCURRENCY" "$BASE_URL$path" | awk -v name="$name" '
    function ms(secs) { return sprintf("%.1f ms", secs * 1000) }
    /Requests\/sec:/      { rps = sprintf("%.0f", $2) }
    /^ +50%+ in/          { p50 = ms($3) }
    /^ +95%+ in/          { p95 = ms($3) }
    /^ +99%+ in/          { p99 = ms($3) }
    /^ +\[[0-9]+\]/       { codes = codes (codes ? ", " : "") substr($1, 2, 3) }
    /^Error distribution/ { errors = 1 }
    END {
      if (errors) codes = codes ", transport errors (see hey output)"
      printf "| %s | %s | %s | %s | %s | %s |\n", name, rps, p50, p95, p99, codes
    }'
done
