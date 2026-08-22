#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────────────────
# One pass that produces everything the pipeline needs:
#   reports/junit/go.xml            JUnit, for the Jenkins test report
#   service/coverage.out            native Go profile, for SonarQube
#   coverage/cobertura-coverage.xml Cobertura, for the new-code coverage gate
#
# It has to be one command because the Tests stage prefers `commands.coverage`
# over `commands.test` — running them separately means whichever one loses is
# never executed, and its outputs silently do not exist.
# ─────────────────────────────────────────────────────────────────────────────
set -Eeuo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
mkdir -p "$ROOT/reports/junit" "$ROOT/coverage"

cd "$ROOT/service"

# -v is required: go-junit-report builds cases from the per-test "=== RUN" and
# "--- PASS" lines, which go test only emits in verbose mode. Without it the XML
# has zero cases and the pipeline reports "no tests found" for a passing suite.
#
# PIPESTATUS is checked explicitly — a pipeline's exit status is the LAST
# command's, so a failing `go test` piped into a succeeding reporter would look
# like a pass.
set +e
go test -v ./... -coverprofile=coverage.out -covermode=atomic 2>&1 \
  | tee "$ROOT/reports/go-test.log" \
  | go-junit-report -set-exit-code > "$ROOT/reports/junit/go.xml"
rc=${PIPESTATUS[0]}
set -e

# Cobertura for diff-cover. Only meaningful if the profile was written.
if [ -s coverage.out ]; then
  gocover-cobertura < coverage.out > "$ROOT/coverage/cobertura-coverage.xml"
  echo "coverage: $(go tool cover -func=coverage.out | tail -1 | awk '{print $3}') overall"
else
  echo "WARNING: no coverage profile produced" >&2
fi

cases=$(grep -c '<testcase' "$ROOT/reports/junit/go.xml" 2>/dev/null || echo 0)
echo "junit: ${cases} test case(s) in reports/junit/go.xml"

exit "$rc"
