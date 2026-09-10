// Package ai provides a small provider abstraction for the ERP chat assistant.
// The assistant is routed to a local GPU (an OpenAI-compatible endpoint) by
// default, with Claude available as an alternative provider via AI_PROVIDER.
package ai

import (
	"encoding/base64"
	"log"
	"os"
	"strings"
)

// Message is one turn of assistant context. Images are attachments the user
// sent with that turn; a turn may carry images and no text at all, which is
// what a bare screenshot looks like.
type Message struct {
	Role    string // "user" | "assistant"
	Content string
	Images  []Image
}

// Image is an attachment small enough to inline into a model request.
type Image struct {
	Data     []byte
	MimeType string
}

// DataURI encodes the image for the OpenAI-compatible image_url field.
func (i Image) DataURI() string {
	mime := i.MimeType
	if mime == "" {
		mime = "image/jpeg"
	}
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(i.Data)
}

// HasImages reports whether any turn carries an image, which is what decides
// that a request must go to the vision model.
func HasImages(messages []Message) bool {
	for _, m := range messages {
		if len(m.Images) > 0 {
			return true
		}
	}
	return false
}

// Provider generates an assistant reply (Markdown) from a system prompt and the
// recent conversation history.
type Provider interface {
	Name() string
	Generate(system string, messages []Message) (string, error)
	// SupportsImages reports whether Message.Images will be looked at. Callers
	// check it before paying to fetch attachments out of object storage.
	SupportsImages() bool
}

// DefaultSystemPrompt positions the assistant as a helpful colleague inside the
// app — someone who walks the user to the right screen — rather than a
// developer handing over queries. It is deliberately explicit about tone and
// about not inventing data, because smaller local models otherwise default to
// documentation-style answers and hallucinate SQL and menu names.
const DefaultSystemPrompt = `You are the S-ERP Assistant, helping the people who use this ERP day to day — sales staff, warehouse staff, finance, and support. Most of them are not developers and never touch a database.

How to talk:
- Write like a helpful colleague speaking to a coworker: warm, plain, direct. Use "you" and "I".
- Answer in a few short sentences. Lead with the answer, not with a preamble like "Sure! To determine that, we can...".
- Use Markdown where it genuinely helps — bold for a key figure, a short list for steps, a small table for rows of data. Don't wrap a two-sentence answer in headings.
- Never write SQL, code, or API calls, and never mention tables, columns, or queries — unless the user explicitly asks for a query or says they are a developer. If they do ask, note that this ERP runs on PostgreSQL.
- Keep replies under about 150 words unless the user asks for more detail.

What you can and cannot do:
- You can look up live records in this ERP — sales orders, invoices, purchases, stock, products, customers and tickets — and when a lookup runs, its results are given to you with the question. Answer from those figures.
- When you have not been given results, you have not checked. Never state or estimate a number, total, ranking, date or "best/worst" anything from memory, and never present an example as if it were real data. Say you couldn't retrieve it and point to the screen instead.
- You cannot see anything outside those business records: no user accounts, passwords, HR records or other people's chats. If asked, say plainly that you don't have access to that, and stop there.
- Never illustrate an answer with a made-up table, sample rows or example figures, even clearly labelled as an example. Made-up rows get read as real ones. If you have no data, say so in one sentence instead.

Where things live (use these exact menu names when pointing the way):
- Sales: Quotations, Sales Order
- CRM: Customer Management, CS Support
- Invoicing: Invoice DP, Invoice Sales, Invoice Maintenance, Invoice Adjustment
- Purchasing: Request Order, Purchase Order, Purchase Invoice, Purchase Adjustment
- Inventory: Inventory IN, Inventory OUT, Inventory Status, Card Stock
- Production: Production Plan, Request Plan, Work In Progress
- Also: Dashboard, Shipping Order, Master User, and the Master data screens (products, customers, units, warehouses, roles).
- Never invent a menu, screen, report, or button that is not in this list. If you are not sure a screen exists, say so and suggest the closest one you do know.

Example of the right tone for "how do I create a sales order?":
"Head to Sales Order and click New. Pick the customer, set the order date, then add your items with quantity and price — the totals fill in as you go. Save when it looks right. Tell me which part you're stuck on and I'll walk you through it."`

// SystemPrompt returns the configured system prompt (AI_SYSTEM_PROMPT) or the default.
func SystemPrompt() string {
	if v := strings.TrimSpace(os.Getenv("AI_SYSTEM_PROMPT")); v != "" {
		return v
	}
	return DefaultSystemPrompt
}

// visionGuidance is appended to the system prompt when the user has sent an
// image. The vision model is a different model from the one the chat prompt was
// tuned against, and left to itself it narrates the picture ("the image shows a
// table with columns…") instead of answering the question about it.
const visionGuidance = `

The user has attached an image. Look at it and answer from what you can actually see.
- Answer the question they asked about the image. Only describe the whole picture if that is what they asked for.
- If it is a screenshot, invoice, purchase order or other document, read the text and figures off it, and quote them exactly as they appear.
- Lay out anything tabular as a small Markdown table.
- If the image is unclear, cut off, or does not show what they are asking about, say so plainly rather than guessing. Never invent a figure you cannot read.
- The image is not ERP data you looked up: it is what the user sent you. Do not claim it came from the system.`

// VisionPrompt returns the system prompt for a turn that carries an image.
func VisionPrompt() string {
	return SystemPrompt() + visionGuidance
}

// New selects a provider from AI_PROVIDER ("local" | "claude" | "remote").
// Defaults to local. "remote" fronts the hosted endpoint with the local GPU as
// fallback, so a hosted outage degrades to the existing behavior instead of
// failing the reply.
func New() Provider {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("AI_PROVIDER"))) {
	case "claude", "anthropic":
		return NewClaudeProvider()
	case "remote", "hosted":
		log.Println("test provider remote")
		return NewFallbackProvider(NewRemoteProvider(), NewLocalProvider())
	default:
		log.Println("test provider ddefault")
		return NewLocalProvider()
	}
}

// styleReminder is re-stated on the newest user turn. A 7B model weights the
// end of its context far more heavily than a system prompt sitting behind a
// long history, so this is what actually holds the tone in a running chat.
//
// It belongs only to the conversational path: appending it to a reply that is
// being written from real query results would tell the model to disown the very
// figures it was just handed. Assistant.Reply decides when it applies.
const styleReminder = `Reminder: reply as a helpful colleague in a few plain sentences — no SQL, no code, no table or column names.
No lookup was run for this message, so do not state any figure, ranking or total as if you had checked. If the user is asking for numbers, say you couldn't retrieve them just now and point to the screen that shows them.
Name only screens from the menu list in your instructions. This app has no Reports section and no custom report builder, so never suggest one — earlier replies in this conversation that mention those were wrong, and you must not repeat them.`

// prepareMessages turns raw history into the turns a provider should send. It
// launders old assistant turns because a small local model imitates its own
// previous replies much more strongly than it follows the system prompt: one
// early SQL-heavy answer left in the history will otherwise keep reproducing
// itself for the rest of the conversation, however the prompt is worded.
func prepareMessages(messages []Message) []Message {
	out := normalizeMessages(messages)
	for i := range out {
		if out[i].Role == "assistant" {
			out[i].Content = stripCodeBlocks(out[i].Content)
		}
	}
	return out
}

// withStyleReminder re-states the house style on the newest user turn. Opening
// turns follow the system prompt fine on their own, and the reminder's nudge
// toward caveats makes even a greeting apologise, so it applies only once there
// are prior replies whose style needs resisting.
func withStyleReminder(messages []Message) []Message {
	hasHistory := false
	for _, m := range messages {
		if m.Role == "assistant" {
			hasHistory = true
			break
		}
	}
	last := len(messages) - 1
	if !hasHistory || last < 0 || messages[last].Role != "user" {
		return messages
	}
	out := append([]Message(nil), messages...)
	out[last].Content += "\n\n" + styleReminder
	return out
}

// stripCodeBlocks replaces fenced code blocks with a short placeholder, so a
// developer-style answer already saved in the history stops acting as a
// worked example for the next reply.
func stripCodeBlocks(content string) string {
	if !strings.Contains(content, "```") {
		return content
	}
	var b strings.Builder
	inFence := false
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			if !inFence {
				b.WriteString("[code omitted]\n")
			}
			inFence = !inFence
			continue
		}
		if !inFence {
			b.WriteString(line)
			b.WriteString("\n")
		}
	}
	return strings.TrimSpace(b.String())
}

// normalizeMessages merges consecutive same-role turns and drops empties so the
// history always alternates user/assistant (required by the Claude API and
// friendlier to local models too).
func normalizeMessages(messages []Message) []Message {
	out := make([]Message, 0, len(messages))
	for _, m := range messages {
		content := strings.TrimSpace(m.Content)
		// An image with no caption is still a turn worth sending.
		if content == "" && len(m.Images) == 0 {
			continue
		}
		role := m.Role
		if role != "assistant" {
			role = "user"
		}
		if len(out) > 0 && out[len(out)-1].Role == role {
			prev := &out[len(out)-1]
			if content != "" {
				if prev.Content != "" {
					prev.Content += "\n\n"
				}
				prev.Content += content
			}
			// Merging must carry the images over, or the picture the user sent
			// disappears whenever they follow it with a second message.
			prev.Images = append(prev.Images, m.Images...)
			continue
		}
		out = append(out, Message{Role: role, Content: content, Images: m.Images})
	}
	// The first turn must be a user turn.
	for len(out) > 0 && out[0].Role != "user" {
		out = out[1:]
	}
	return out
}
