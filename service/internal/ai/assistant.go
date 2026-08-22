package ai

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
)

// Assistant answers a chat turn, consulting the ERP database when the question
// is about stored records.
//
// It runs two passes rather than native tool-calling: the first turns the
// question into a single SELECT (or declines with NO_QUERY), the second writes
// the reply from the rows that came back. Two passes work on any
// OpenAI-compatible endpoint regardless of whether the served model supports
// tool schemas, and they keep the generated SQL somewhere it can be vetted and
// logged before it reaches the database.
//
// With no DataSource configured the Assistant degrades to a plain chat reply.
type Assistant struct {
	provider Provider
	data     *DataSource
}

func NewAssistant(provider Provider, data *DataSource) *Assistant {
	return &Assistant{provider: provider, data: data}
}

// SupportsImages reports whether the configured provider can read images.
func (a *Assistant) SupportsImages() bool { return a.provider.SupportsImages() }

func (a *Assistant) Name() string {
	if a.data == nil {
		return a.provider.Name() + " (no database access)"
	}
	return a.provider.Name() + " + ERP data"
}

// noQuery is the planner's way of saying the question needs no lookup.
const noQuery = "NO_QUERY"

// Reply produces the assistant's answer for a conversation.
func (a *Assistant) Reply(ctx context.Context, messages []Message) (string, error) {
	// Only an image the user has just sent makes this an image question. Asking
	// about the whole history instead would pin the conversation to the vision
	// model forever: every later question, on any subject, would be answered
	// from a screenshot sent earlier rather than from the database.
	if hasFreshImages(messages) {
		reply, err := a.provider.Generate(VisionPrompt(), recentTurns(messages, visionHistoryTurns))
		if err != nil && isContextOverflow(err) {
			// Images are the expensive part and the conversation is the
			// disposable part, so shed the history and try once more with just
			// the turn that carries the picture.
			log.Printf("chat AI: vision context overflowed, retrying with the latest turn only")
			return a.provider.Generate(VisionPrompt(), recentTurns(messages, 1))
		}
		return reply, err
	}

	// Past this point the question is textual, so older images are dropped:
	// they would otherwise route the request to the vision model and be paid
	// for in tokens while contributing nothing.
	messages = withoutImages(messages)

	if a.data == nil {
		return a.provider.Generate(SystemPrompt(), withStyleReminder(messages))
	}

	// The rows are attached to the turn that asked for them, which is the newest
	// user turn rather than simply the last element — history can end with an
	// assistant turn, and hanging data off that would leave the question
	// unanswered.
	target := lastUserIndex(messages)
	if target < 0 {
		return a.provider.Generate(SystemPrompt(), withStyleReminder(messages))
	}

	result, query, ok := a.lookup(ctx, planningContext(messages, target))
	if !ok {
		// Either no lookup was warranted or it failed; answer conversationally.
		return a.provider.Generate(SystemPrompt(), withStyleReminder(messages))
	}
	log.Printf("chat AI: answered from data (%d rows) using: %s", len(result.Rows), collapse(query))

	withData := append([]Message(nil), messages...)
	withData[target].Content = fmt.Sprintf(
		"%s\n\n--- Live data from the ERP database for this question ---\n%s",
		withData[target].Content, result.Markdown(),
	)
	return a.provider.Generate(answerPrompt, withData)
}

// maxPlanAttempts bounds how many times the model may try to produce a working
// query. Each attempt is one fast local round trip, and a failure carries a
// specific correction, so a third try is usually the one that lands.
const maxPlanAttempts = 3

// lookup plans a query, runs it, and retries with the database's own error plus
// a concrete correction when an attempt does not parse or run.
func (a *Assistant) lookup(ctx context.Context, question string) (*ResultSet, string, bool) {
	question = strings.TrimSpace(question)
	if question == "" {
		return nil, "", false
	}

	planning := []Message{{Role: "user", Content: question}}
	for attempt := 0; attempt < maxPlanAttempts; attempt++ {
		raw, err := a.provider.Generate(sqlPlannerPrompt, planning)
		if err != nil {
			log.Printf("chat AI: query planning failed: %v", err)
			return nil, "", false
		}
		if strings.Contains(strings.ToUpper(raw), noQuery) {
			return nil, "", false
		}

		result, err := a.data.Query(ctx, raw)
		if err == nil {
			return result, raw, true
		}
		log.Printf("chat AI: generated query rejected (attempt %d): %v — query: %s", attempt+1, err, collapse(raw))

		// Feed the failure back with a specific correction. A bare error message
		// tends to get the identical query back; naming the fix does not.
		planning = []Message{{Role: "user", Content: fmt.Sprintf(
			"%s\n\nYour previous query failed:\n%s\n\nIt was:\n%s\n\n%s\nWrite the corrected query, or reply %s if it cannot be done.",
			question, err.Error(), collapse(raw), repairHint(err), noQuery,
		)}}
	}
	return nil, "", false
}

// visionHistoryTurns bounds how much conversation accompanies an image. Images
// are expensive in tokens, and the whole 20-message scrollback on top of them
// overflows the model's context; the last few turns are all that is needed to
// understand what is being asked about the picture.
const visionHistoryTurns = 6

// hasFreshImages reports whether the user attached an image to what they are
// asking now — meaning anywhere in the unanswered run of user turns, not just
// the final row: sending a screenshot and then typing the question arrives as
// two consecutive user messages, and both belong to the same question.
func hasFreshImages(messages []Message) bool {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "assistant" {
			return false // reached the previous answer; anything older is history
		}
		if len(messages[i].Images) > 0 {
			return true
		}
	}
	return false
}

// withoutImages copies the history with attachments removed, leaving the text
// intact. The originals are not modified.
func withoutImages(messages []Message) []Message {
	out := make([]Message, len(messages))
	for i, m := range messages {
		m.Images = nil
		out[i] = m
	}
	return out
}

// isContextOverflow spots the model server complaining that the prompt is
// longer than its context window. Ollama, llama.cpp and vLLM each word it
// differently, so this matches on the shape rather than an exact string.
func isContextOverflow(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "exceed_context_size") ||
		strings.Contains(msg, "context size") ||
		strings.Contains(msg, "context length") ||
		strings.Contains(msg, "context window") ||
		strings.Contains(msg, "too many tokens")
}

// recentTurns keeps the newest n turns, oldest first.
func recentTurns(messages []Message, n int) []Message {
	if len(messages) <= n {
		return messages
	}
	return messages[len(messages)-n:]
}

// planningContext frames the question for the planner, preceded by the recent
// turns. A follow-up carries no meaning alone — "no i dont want navigate, i
// want the number" is unplannable in isolation but obvious after the question
// before it — so the planner needs the thread, not just the latest line.
func planningContext(messages []Message, target int) string {
	const contextTurns = 4
	start := target - contextTurns
	if start < 0 {
		start = 0
	}
	var b strings.Builder
	if start < target {
		b.WriteString("Conversation so far:\n")
		for _, m := range messages[start:target] {
			b.WriteString(m.Role + ": " + truncate(collapse(m.Content), 300) + "\n")
		}
		b.WriteString("\n")
	}
	b.WriteString("Current question: " + strings.TrimSpace(messages[target].Content))
	return b.String()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// repairHint turns a database or guard error into the specific correction to
// make. These are the mistakes a small model actually repeats: the failures
// come back identical when it is only shown the error text.
func repairHint(err error) string {
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "is ambiguous"):
		return "Fix: that column exists in both joined tables. Drop the join and query the single table that already has every column you need — sales_order_items carries order_at, sales_order_no and customer_name by itself."
	case strings.Contains(msg, "must appear in the group by"):
		return "Fix: you added ORDER BY or a plain column to an aggregate. For one total or count, use only SUM(...) or COUNT(*) with no GROUP BY and no ORDER BY."
	case strings.Contains(msg, "does not exist"):
		return "Fix: that column or table is not in the schema. Use only the names listed, spelled exactly."
	case strings.Contains(msg, "syntax"), strings.Contains(msg, "invalid input syntax"):
		return "Fix: write the simplest query that answers the question — no casts and no arithmetic tricks."
	case errors.Is(err, ErrForbiddenSQL), errors.Is(err, ErrMultiStatment), errors.Is(err, ErrNotAQuery):
		return "Fix: reply with exactly one plain SELECT statement and nothing else — no comments, no semicolon, no explanation."
	default:
		return "Fix: write a simpler query using only the listed tables and columns."
	}
}

// lastUserIndex locates the newest user turn — the question being answered —
// or -1 when there is none.
func lastUserIndex(messages []Message) int {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			return i
		}
	}
	return -1
}

// collapse puts a statement on one line for logging.
func collapse(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
