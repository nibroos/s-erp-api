package ai

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// stubProvider records what each pass was asked and replays scripted answers.
type stubProvider struct {
	replies  []string
	systems  []string
	messages [][]Message
	vision   bool
}

func (s *stubProvider) Name() string { return "stub" }

func (s *stubProvider) SupportsImages() bool { return s.vision }

func (s *stubProvider) Generate(system string, messages []Message) (string, error) {
	s.systems = append(s.systems, system)
	s.messages = append(s.messages, messages)
	reply := s.replies[0]
	if len(s.replies) > 1 {
		s.replies = s.replies[1:]
	}
	return reply, nil
}

// With no DataSource the assistant must still answer, using the chat prompt.
func TestReplyWithoutDataSourceFallsBackToChat(t *testing.T) {
	p := &stubProvider{replies: []string{"hello there"}}
	got, err := NewAssistant(p, nil).Reply(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello there" {
		t.Fatalf("unexpected reply: %q", got)
	}
	if len(p.systems) != 1 || p.systems[0] != SystemPrompt() {
		t.Fatalf("expected the chat prompt, got %d call(s)", len(p.systems))
	}
}

func TestLastUserIndex(t *testing.T) {
	// History can end with an assistant turn; the question is still the newest
	// user turn, and attaching query results anywhere else loses the question.
	msgs := []Message{
		{Role: "user", Content: "first"},
		{Role: "user", Content: "the question"},
		{Role: "assistant", Content: "a reply"},
	}
	if got := lastUserIndex(msgs); got != 1 {
		t.Fatalf("got %d, want 1", got)
	}
	if got := lastUserIndex([]Message{{Role: "assistant", Content: "x"}}); got != -1 {
		t.Fatalf("got %d, want -1 when there is no user turn", got)
	}
	if got := lastUserIndex(nil); got != -1 {
		t.Fatalf("got %d, want -1 for empty history", got)
	}
}

// NO_QUERY means the question needs no lookup, so the reply must come from the
// chat prompt and must not be a second planning pass.
func TestNoQueryUsesChatPrompt(t *testing.T) {
	p := &stubProvider{replies: []string{noQuery, "a friendly answer"}}
	a := NewAssistant(p, &DataSource{}) // non-nil: the data path is attempted
	got, err := a.Reply(context.Background(), []Message{{Role: "user", Content: "how do i create a sales order?"}})
	if err != nil {
		t.Fatal(err)
	}
	if got != "a friendly answer" {
		t.Fatalf("unexpected reply: %q", got)
	}
	if len(p.systems) != 2 {
		t.Fatalf("expected a planning pass then an answer pass, got %d", len(p.systems))
	}
	if p.systems[0] != sqlPlannerPrompt || p.systems[1] != SystemPrompt() {
		t.Fatalf("wrong prompts used: %q then %q", short(p.systems[0]), short(p.systems[1]))
	}
	// The style reminder must not leak into a planning pass — it would corrupt
	// the SQL the planner is being asked for.
	if strings.Contains(p.messages[0][0].Content, styleReminder) {
		t.Fatal("style reminder leaked into the SQL planning pass")
	}
}

// An image turn must bypass SQL planning entirely: the planner is a separate
// text-only model that cannot see the picture, so it would query for something
// the user never named.
func TestImageTurnSkipsSQLPlanning(t *testing.T) {
	p := &stubProvider{replies: []string{"that invoice totals 1,500"}, vision: true}
	a := NewAssistant(p, &DataSource{})
	got, err := a.Reply(context.Background(), []Message{{
		Role:    "user",
		Content: "what is the total on this?",
		Images:  []Image{{Data: []byte{0x1, 0x2}, MimeType: "image/png"}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if got != "that invoice totals 1,500" {
		t.Fatalf("unexpected reply: %q", got)
	}
	if len(p.systems) != 1 {
		t.Fatalf("expected exactly one pass, got %d", len(p.systems))
	}
	if p.systems[0] != VisionPrompt() {
		t.Fatal("image turn did not use the vision prompt")
	}
}

// The reported bug: after one screenshot, every later question was still being
// answered from that screenshot instead of from the database.
func TestQuestionAfterAnImageGoesBackToData(t *testing.T) {
	p := &stubProvider{replies: []string{noQuery, "a text answer"}, vision: true}
	a := NewAssistant(p, &DataSource{})

	got, err := a.Reply(context.Background(), []Message{
		{Role: "user", Content: "what is this?", Images: []Image{{Data: []byte{1}, MimeType: "image/png"}}},
		{Role: "assistant", Content: "It is a sales order screenshot."},
		{Role: "user", Content: "now a different context: total sales of PT Yubi Technology?"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != "a text answer" {
		t.Fatalf("unexpected reply: %q", got)
	}
	// It must have tried the SQL planner, which the vision path skips entirely.
	if len(p.systems) != 2 || p.systems[0] != sqlPlannerPrompt {
		t.Fatalf("question was not routed to the data path: %d pass(es)", len(p.systems))
	}
	// And the stale image must not have been carried along.
	for _, sent := range p.messages {
		if HasImages(sent) {
			t.Fatal("a stale image was sent with a text question")
		}
	}
}

// Pasting a screenshot and then typing the question arrives as two consecutive
// user messages; both are part of the same question.
func TestImageThenTextIsStillAnImageQuestion(t *testing.T) {
	img := []Message{
		{Role: "user", Images: []Image{{Data: []byte{1}, MimeType: "image/png"}}},
		{Role: "user", Content: "what is the total here?"},
	}
	if !hasFreshImages(img) {
		t.Fatal("image followed by a question was not treated as an image turn")
	}
	// But once answered, the next question stands on its own.
	answered := append(append([]Message{}, img...),
		Message{Role: "assistant", Content: "It says 17,750,000."},
		Message{Role: "user", Content: "and our best selling month?"},
	)
	if hasFreshImages(answered) {
		t.Fatal("a new question after an answer is still stuck on the old image")
	}
}

func TestWithoutImagesLeavesTextIntact(t *testing.T) {
	in := []Message{{Role: "user", Content: "hello", Images: []Image{{Data: []byte{1}}}}}
	out := withoutImages(in)
	if len(out[0].Images) != 0 {
		t.Error("images were not stripped")
	}
	if out[0].Content != "hello" {
		t.Errorf("text was altered: %q", out[0].Content)
	}
	if len(in[0].Images) != 1 {
		t.Error("the caller's slice was mutated")
	}
}

func TestHasImages(t *testing.T) {
	if HasImages([]Message{{Role: "user", Content: "hi"}}) {
		t.Fatal("reported images on a text-only turn")
	}
	if !HasImages([]Message{{Role: "user", Images: []Image{{Data: []byte{1}}}}}) {
		t.Fatal("missed an image turn")
	}
}

func TestImageDataURI(t *testing.T) {
	got := Image{Data: []byte("hello"), MimeType: "image/png"}.DataURI()
	if got != "data:image/png;base64,aGVsbG8=" {
		t.Fatalf("unexpected data URI: %q", got)
	}
	// A missing mime type must still produce a usable URI.
	if got := (Image{Data: []byte("x")}).DataURI(); !strings.HasPrefix(got, "data:image/jpeg;base64,") {
		t.Fatalf("unexpected fallback: %q", got)
	}
}

// normalizeMessages merges consecutive same-role turns; the merge must carry
// images across or a picture vanishes as soon as the user sends a follow-up.
func TestNormalizeKeepsImagesWhenMergingTurns(t *testing.T) {
	out := normalizeMessages([]Message{
		{Role: "user", Content: "look at this", Images: []Image{{Data: []byte{1}, MimeType: "image/png"}}},
		{Role: "user", Content: "what does it say?"},
	})
	if len(out) != 1 {
		t.Fatalf("expected the turns to merge, got %d", len(out))
	}
	if len(out[0].Images) != 1 {
		t.Fatalf("image was dropped by the merge: %+v", out[0])
	}
	if !strings.Contains(out[0].Content, "look at this") || !strings.Contains(out[0].Content, "what does it say?") {
		t.Fatalf("text was lost: %q", out[0].Content)
	}
}

// A caption-less screenshot is a real message and must survive normalization.
func TestNormalizeKeepsImageOnlyTurn(t *testing.T) {
	out := normalizeMessages([]Message{
		{Role: "user", Images: []Image{{Data: []byte{1}, MimeType: "image/png"}}},
	})
	if len(out) != 1 || len(out[0].Images) != 1 {
		t.Fatalf("image-only turn was dropped: %+v", out)
	}
}

// A follow-up like "i want the number" is unplannable alone, so the planner
// must receive the turns before it.
func TestPlanningContextIncludesRecentTurns(t *testing.T) {
	msgs := []Message{
		{Role: "user", Content: "can you total sales of Buyer 2?"},
		{Role: "assistant", Content: "Head to Invoicing and filter by buyer."},
		{Role: "user", Content: "no i dont want navigate, i want the number"},
	}
	got := planningContext(msgs, 2)
	for _, want := range []string{"Conversation so far", "total sales of Buyer 2", "Current question: no i dont want navigate"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

// A first message has no prior turns to quote.
func TestPlanningContextFirstTurnIsBare(t *testing.T) {
	got := planningContext([]Message{{Role: "user", Content: "how many orders?"}}, 0)
	if got != "Current question: how many orders?" {
		t.Fatalf("unexpected framing: %q", got)
	}
}

// Repeating the error verbatim gets the same broken query back, so each failure
// must carry the correction that applies to it.
func TestRepairHintTargetsTheMistake(t *testing.T) {
	cases := map[string]string{
		`column reference "order_at" is ambiguous`:  "Drop the join",
		"column must appear in the GROUP BY clause": "no GROUP BY and no ORDER BY",
		`column "product_id" does not exist`:        "not in the schema",
		"syntax error at or near":                   "simplest query",
		"":                                          "simpler query",
	}
	for errText, want := range cases {
		if got := repairHint(errors.New(errText)); !strings.Contains(got, want) {
			t.Errorf("for %q got %q, want it to mention %q", errText, got, want)
		}
	}
	if got := repairHint(ErrMultiStatment); !strings.Contains(got, "one plain SELECT") {
		t.Errorf("guard rejection hint was %q", got)
	}
}

func short(s string) string {
	if len(s) > 40 {
		return s[:40]
	}
	return s
}
