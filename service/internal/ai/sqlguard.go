package ai

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// maxResultRows caps what one generated query may return. It is enforced in SQL
// (so the database stops early) and again while scanning.
const maxResultRows = 200

var (
	ErrNotAQuery     = errors.New("the generated statement is not a read-only query")
	ErrMultiStatment = errors.New("only a single statement may be run")
	ErrForbiddenSQL  = errors.New("the generated query uses a forbidden construct")
)

// forbidden matches statement kinds and functions that must never appear in a
// generated query. The database would reject nearly all of them anyway — the
// role holds no privileges and the transaction is read-only — but failing here
// keeps a bad query out of the database entirely and gives the model a specific
// error to repair. Word boundaries matter: `\bupdate\b` must not fire on the
// `updated_at` column, and `\bcreate\b` must not fire on `created_at`.
var forbidden = regexp.MustCompile(`(?is)\b(insert|update|delete|drop|alter|truncate|create|grant|revoke|comment|copy|vacuum|analyze|reindex|cluster|lock|call|do|prepare|execute|deallocate|listen|notify|unlisten|discard|set|reset|begin|commit|rollback|savepoint|refresh|import|security|dblink|pg_read_file|pg_read_binary_file|pg_ls_dir|pg_stat_file|pg_sleep|pg_terminate_backend|pg_cancel_backend|pg_reload_conf|lo_import|lo_export|current_setting|set_config|pg_authid|pg_shadow|pg_user_mapping)\b`)

// leadingKeyword matches the statement kinds we accept.
var leadingKeyword = regexp.MustCompile(`(?is)^\s*(select|with)\b`)

// fencedBlock pulls SQL back out of a Markdown code fence, which models add
// even when told not to.
var fencedBlock = regexp.MustCompile("(?s)```(?:sql)?\\s*(.*?)```")

// sanitizeQuery extracts the SQL a model produced and rejects anything that is
// not a single plain read-only query.
func sanitizeQuery(raw string) (string, error) {
	query := strings.TrimSpace(raw)
	if m := fencedBlock.FindStringSubmatch(query); m != nil {
		query = strings.TrimSpace(m[1])
	}
	// A trailing semicolon is habitual and harmless; anything after one is not.
	query = strings.TrimRight(query, "; \t\r\n")
	if query == "" {
		return "", ErrNotAQuery
	}
	if strings.Contains(query, ";") {
		return "", ErrMultiStatment
	}
	// Line comments could hide the rest of the statement from a human reader,
	// and block comments can smuggle keywords past the checks below.
	if strings.Contains(query, "--") || strings.Contains(query, "/*") {
		return "", ErrForbiddenSQL
	}
	if !leadingKeyword.MatchString(query) {
		return "", ErrNotAQuery
	}
	if match := forbidden.FindString(query); match != "" {
		return "", fmt.Errorf("%w: %s", ErrForbiddenSQL, strings.ToUpper(match))
	}
	return query, nil
}

// wrapQuery nests a vetted query inside a bounded outer SELECT. Beyond forcing
// a row cap this makes statement chaining structurally impossible: a semicolon
// inside the parentheses is a syntax error rather than a second statement,
// which matters because lib/pq sends argument-less queries over the simple
// protocol, where multiple statements would otherwise be accepted.
func wrapQuery(query string) string {
	return fmt.Sprintf("SELECT * FROM (\n%s\n) AS ai_result LIMIT %d", query, maxResultRows)
}
