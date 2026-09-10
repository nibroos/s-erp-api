package ai

import (
	"strings"
	"testing"
)

func TestStripCodeBlocks(t *testing.T) {
	in := "Here is how:\n\n```sql\nSELECT * FROM sales_orders;\n```\n\nRun that query."
	got := stripCodeBlocks(in)
	if strings.Contains(got, "SELECT") {
		t.Fatalf("query survived stripping: %q", got)
	}
	if !strings.Contains(got, "Here is how:") || !strings.Contains(got, "Run that query.") {
		t.Fatalf("prose was lost: %q", got)
	}
}

func TestStripCodeBlocksLeavesPlainTextAlone(t *testing.T) {
	in := "Go to Sales Order and filter by date."
	if got := stripCodeBlocks(in); got != in {
		t.Fatalf("plain text altered: %q", got)
	}
}

// An unterminated fence must not swallow the rest of the reply beyond its own
// block — everything after an opening fence is dropped, which is the safe side.
func TestStripCodeBlocksUnterminatedFence(t *testing.T) {
	got := stripCodeBlocks("Intro line.\n\n```\nSELECT 1;")
	if strings.Contains(got, "SELECT") {
		t.Fatalf("query survived stripping: %q", got)
	}
	if !strings.Contains(got, "Intro line.") {
		t.Fatalf("prose before the fence was lost: %q", got)
	}
}

func TestPrepareMessagesLaundersAssistantHistory(t *testing.T) {
	out := prepareMessages([]Message{
		{Role: "user", Content: "best selling month?"},
		{Role: "assistant", Content: "Here you go:\n```sql\nSELECT 1;\n```"},
		{Role: "user", Content: "and the best product?"},
	})
	if strings.Contains(out[1].Content, "SELECT") {
		t.Fatalf("assistant history was not laundered: %q", out[1].Content)
	}
	// Laundering must not append anything — the reminder is applied separately,
	// and must never reach a reply being written from real query results.
	if last := out[len(out)-1]; last.Content != "and the best product?" {
		t.Fatalf("final turn should be untouched, got %q", last.Content)
	}
}

// Once the assistant has replied, the newest user turn carries the style
// reminder — that recency is what holds the tone against a polluted history.
func TestWithStyleReminderAppendsAfterHistory(t *testing.T) {
	in := []Message{
		{Role: "user", Content: "best selling month?"},
		{Role: "assistant", Content: "some earlier answer"},
		{Role: "user", Content: "and the best product?"},
	}
	out := withStyleReminder(in)
	last := out[len(out)-1]
	if !strings.HasPrefix(last.Content, "and the best product?") || !strings.Contains(last.Content, styleReminder) {
		t.Fatalf("reminder missing from final user turn: %q", last.Content)
	}
	// The caller's slice must not be mutated: Assistant reuses it for the
	// data path, which must never carry the "no lookup was run" wording.
	if in[2].Content != "and the best product?" {
		t.Fatalf("input was mutated: %q", in[2].Content)
	}
}

// A first message needs no reminder: the system prompt alone handles it, and
// appending one nudges even a greeting toward an unwanted data caveat.
func TestWithStyleReminderSkipsFirstTurn(t *testing.T) {
	out := withStyleReminder([]Message{{Role: "user", Content: "hi"}})
	if len(out) != 1 || out[0].Content != "hi" {
		t.Fatalf("first turn should be untouched, got %+v", out)
	}
}
