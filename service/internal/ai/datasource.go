package ai

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// queryTimeout bounds a single generated query. The model occasionally writes
// an accidental cross join; the database should give up rather than the chat
// hanging on the "typing" indicator.
const queryTimeout = 10 * time.Second

// DataSource executes assistant-generated SELECTs against the curated `ai`
// schema. Every query runs inside one read-only transaction that has first
// switched to the unprivileged `ai_readonly` role, so PostgreSQL — not the
// string checks in sqlguard.go — is what actually keeps the assistant away from
// password hashes, refresh tokens, HR records and private chat messages.
type DataSource struct {
	db *sqlx.DB
}

func NewDataSource(db *sqlx.DB) *DataSource {
	if db == nil {
		return nil
	}
	return &DataSource{db: db}
}

// ResultSet is a query result rendered as text, ready to hand back to a model.
type ResultSet struct {
	Columns   []string
	Rows      [][]string
	Truncated bool
}

func (r *ResultSet) Empty() bool { return r == nil || len(r.Rows) == 0 }

// Markdown renders the result as a Markdown table so the model reads it in the
// same shape it is expected to answer in.
func (r *ResultSet) Markdown() string {
	if r.Empty() {
		return "(no rows)"
	}
	var b strings.Builder
	b.WriteString("| " + strings.Join(r.Columns, " | ") + " |\n")
	b.WriteString("|" + strings.Repeat(" --- |", len(r.Columns)) + "\n")
	for _, row := range r.Rows {
		b.WriteString("| " + strings.Join(row, " | ") + " |\n")
	}
	if r.Truncated {
		fmt.Fprintf(&b, "\n(showing the first %d rows)\n", maxResultRows)
	}
	return b.String()
}

// Query vets, bounds and runs a generated query.
func (d *DataSource) Query(ctx context.Context, generated string) (*ResultSet, error) {
	query, err := sanitizeQuery(generated)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	tx, err := d.db.BeginTxx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck // read-only; nothing to commit

	// Drop to the unprivileged role for the rest of the transaction. SET LOCAL
	// cannot outlive it, and the generated query cannot undo it: sanitizeQuery
	// rejects both SET and multi-statement input.
	if _, err := tx.ExecContext(ctx, "SET LOCAL ROLE ai_readonly"); err != nil {
		return nil, fmt.Errorf("ai_readonly role unavailable (run migration 20260713000009): %w", err)
	}
	if _, err := tx.ExecContext(ctx, "SET LOCAL statement_timeout = '10s'"); err != nil {
		return nil, err
	}
	// Unqualified names resolve to the curated views, so the model may write
	// either `sales_orders` or `ai.sales_orders`.
	if _, err := tx.ExecContext(ctx, "SET LOCAL search_path = ai"); err != nil {
		return nil, err
	}

	rows, err := tx.QueryxContext(ctx, wrapQuery(query))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	out := &ResultSet{Columns: cols}
	for rows.Next() {
		if len(out.Rows) >= maxResultRows {
			out.Truncated = true
			break
		}
		raw, err := rows.SliceScan()
		if err != nil {
			return nil, err
		}
		row := make([]string, len(raw))
		for i, v := range raw {
			row[i] = formatValue(v)
		}
		out.Rows = append(out.Rows, row)
	}
	return out, rows.Err()
}

// formatValue renders one cell for the model: compact, unambiguous, and free of
// the pipes and newlines that would break the Markdown table it lands in.
func formatValue(v interface{}) string {
	var s string
	switch t := v.(type) {
	case nil:
		return "null"
	case []byte:
		s = string(t)
	case string:
		s = t
	case time.Time:
		// Dates carry no useful clock component in this schema.
		if t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 {
			return t.Format("2006-01-02")
		}
		return t.Format("2006-01-02 15:04")
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	default:
		s = fmt.Sprint(t)
	}
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 120 {
		s = s[:120] + "…"
	}
	return s
}
