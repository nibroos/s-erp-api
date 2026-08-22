package ai

import (
	"strings"
	"testing"
)

func TestSanitizeQueryAcceptsPlainSelects(t *testing.T) {
	cases := []string{
		"SELECT count(*) FROM sales_orders",
		"select to_char(order_at,'YYYY-MM') AS m, sum(grand_total) FROM sales_orders GROUP BY m ORDER BY 2 DESC LIMIT 1",
		"SELECT * FROM ai.products WHERE created_at > now() - interval '30 days'",
		"WITH monthly AS (SELECT 1 AS n) SELECT n FROM monthly",
		// A trailing semicolon is habitual; it should be trimmed, not rejected.
		"SELECT 1;",
		// Column names that merely contain forbidden keywords must pass.
		"SELECT created_at, updated_at FROM sales_orders ORDER BY created_at DESC",
		"SELECT id FROM sales_orders OFFSET 5 LIMIT 5",
		"SELECT do_at, direction FROM inventories",
	}
	for _, q := range cases {
		if _, err := sanitizeQuery(q); err != nil {
			t.Errorf("rejected a valid query %q: %v", q, err)
		}
	}
}

func TestSanitizeQueryRejectsEverythingElse(t *testing.T) {
	cases := map[string]string{
		"write":               "DELETE FROM sales_orders",
		"insert":              "INSERT INTO customers (name) VALUES ('x')",
		"ddl":                 "DROP TABLE sales_orders",
		"chained statement":   "SELECT 1; DROP TABLE customers",
		"chained write":       "SELECT 1; UPDATE users SET password='x'",
		"role escape":         "SELECT 1; RESET ROLE",
		"inline role escape":  "SELECT set_config('role','postgres',false)",
		"file read":           "SELECT pg_read_file('/etc/passwd')",
		"copy out":            "COPY (SELECT 1) TO '/tmp/x'",
		"line comment":        "SELECT 1 -- DROP TABLE customers",
		"block comment":       "SELECT /* sneaky */ 1",
		"credential table":    "SELECT * FROM pg_authid",
		"settings":            "SELECT current_setting('is_superuser')",
		"transaction control": "COMMIT",
		"empty":               "   ",
		"prose":               "I cannot answer that",
	}
	for name, q := range cases {
		if got, err := sanitizeQuery(q); err == nil {
			t.Errorf("%s: accepted %q (returned %q)", name, q, got)
		}
	}
}

// Models wrap SQL in a fence however firmly they are told not to.
func TestSanitizeQueryUnwrapsCodeFences(t *testing.T) {
	got, err := sanitizeQuery("```sql\nSELECT count(*) FROM customers\n```")
	if err != nil {
		t.Fatal(err)
	}
	if got != "SELECT count(*) FROM customers" {
		t.Fatalf("unexpected unwrap: %q", got)
	}
}

// The wrap is what makes statement chaining a syntax error rather than a second
// statement, and what guarantees the row cap regardless of the model's LIMIT.
func TestWrapQueryBoundsResults(t *testing.T) {
	got := wrapQuery("SELECT 1")
	if !strings.HasPrefix(got, "SELECT * FROM (") {
		t.Fatalf("query was not nested: %q", got)
	}
	if !strings.Contains(got, "LIMIT 200") {
		t.Fatalf("row cap missing: %q", got)
	}
}
