package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/nibroos/s-erp-api/service/internal/ai"
	"github.com/nibroos/s-erp-api/service/internal/chat"
	"github.com/nibroos/s-erp-api/service/internal/config"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/repository"
)

const maxMessageLength = 4000

var (
	ErrForbiddenConversation = errors.New("you are not a participant of this conversation")
	ErrEmptyMessage          = errors.New("message content cannot be empty")
	ErrCannotChatSelf        = errors.New("cannot start a conversation with yourself")
	ErrForbiddenAction       = errors.New("you do not have permission for this action")
	ErrCannotKickOwner       = errors.New("the group owner cannot be removed")
	ErrOwnerMustPromote      = errors.New("promote a member to admin before leaving as owner")
	ErrInvalidRole           = errors.New("invalid role")
	ErrEmptyGroup            = errors.New("a group needs a title and at least one member")
	ErrInvalidReaction       = errors.New("invalid reaction")
	ErrUploadUnavailable     = errors.New("file storage is not available")
	ErrFileTooLarge          = errors.New("file is empty or exceeds the size limit")
)

type ChatService struct {
	repo  *repository.ChatRepository
	hub   *chat.Hub
	minio *config.MinioStorage
	ai    *ai.Assistant

	rabbitmq       *config.RabbitMQ // AI-reply work-queue publisher (nil = disabled)
	aiQueueEnabled bool             // publish AI replies to the queue vs generate in-process
	aiSem          chan struct{}    // bounds concurrent in-process AI generations (fallback path)
}

func NewChatService(repo *repository.ChatRepository, hub *chat.Hub, minio *config.MinioStorage, assistant *ai.Assistant) *ChatService {
	return &ChatService{
		repo: repo, hub: hub, minio: minio, ai: assistant,
		aiSem: make(chan struct{}, aiInProcessConcurrency()),
	}
}

const (
	// chatAIReplyQueue is the work queue that carries "generate an AI reply for
	// this message" jobs from the API to the consumer-service.
	chatAIReplyQueue = "chat_ai_reply_queue"
	chatAIReplyDLQ   = "chat_ai_reply_dlq"
	// ChatAIReplyQueueName is exported for the consumer that reads the queue.
	ChatAIReplyQueueName = chatAIReplyQueue
)

// AIReplyJob is the payload published to chatAIReplyQueue.
type AIReplyJob struct {
	ConversationID   uint `json:"conversation_id"`
	AIUserID         uint `json:"ai_user_id"`
	TriggerMessageID uint `json:"trigger_message_id"`
}

// DeclareAIReplyQueue declares the AI-reply work queue and its dead-letter queue
// with the exact arguments the publisher (API) and consumer (worker) must agree
// on. Both call this at startup so a message is never published to a missing
// queue (the default exchange would silently drop it). Idempotent.
func DeclareAIReplyQueue(ch *amqp.Channel) error {
	if _, err := ch.QueueDeclare(chatAIReplyDLQ, true, false, false, false, nil); err != nil {
		return err
	}
	_, err := ch.QueueDeclare(chatAIReplyQueue, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": chatAIReplyDLQ,
	})
	return err
}

// EnableAIQueue routes AI replies through the RabbitMQ work queue (consumed by
// the consumer-service) instead of generating them in-process. It declares the
// queue up front; if that fails it leaves the queue disabled so sends fall back
// to the bounded in-process path.
func (s *ChatService) EnableAIQueue(rabbitmq *config.RabbitMQ) {
	if rabbitmq == nil || rabbitmq.Channel == nil {
		return
	}
	if err := DeclareAIReplyQueue(rabbitmq.Channel); err != nil {
		log.Printf("chat AI: could not declare reply queue, staying in-process: %v", err)
		return
	}
	s.rabbitmq = rabbitmq
	s.aiQueueEnabled = true
	log.Printf("chat AI: replies dispatched via RabbitMQ queue %q", chatAIReplyQueue)
}

// aiInProcessConcurrency bounds how many AI replies the in-process (fallback)
// path may generate at once — set to the model server's real parallelism so a
// burst of users can't spawn unbounded goroutines all blocked on inference.
func aiInProcessConcurrency() int {
	n := 2
	if v := os.Getenv("AI_MAX_CONCURRENCY"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			n = parsed
		}
	}
	return n
}

const maxUploadBytes = 50 << 20 // 50 MiB

var attachmentKinds = map[string]bool{
	"image": true, "video": true, "voice": true, "document": true,
}

// kindForMime maps a content type to an attachment kind.
func kindForMime(mime string) string {
	switch {
	case strings.HasPrefix(mime, "image/"):
		return "image"
	case strings.HasPrefix(mime, "video/"):
		return "video"
	case strings.HasPrefix(mime, "audio/"):
		return "voice"
	default:
		return "document"
	}
}

// StartConversation opens (or creates) the direct conversation between the
// authenticated user and otherID and returns its id.
func (s *ChatService) StartConversation(ctx *fiber.Ctx, meID, otherID uint) (uint, error) {
	if meID == otherID {
		return 0, ErrCannotChatSelf
	}
	return s.repo.GetOrCreateDirectConversation(ctx, meID, otherID)
}

// StartAIChat opens (or creates) the direct conversation with the AI assistant.
func (s *ChatService) StartAIChat(ctx *fiber.Ctx, meID uint) (uint, error) {
	aiID, ok, err := s.repo.GetAIUserID(ctx)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, ErrForbiddenAction
	}
	return s.repo.GetOrCreateDirectConversation(ctx, meID, aiID)
}

func (s *ChatService) ListConversations(ctx *fiber.Ctx, meID uint, filters map[string]string) ([]dtos.ConversationListDTO, int, error) {
	conversations, total, err := s.repo.ListConversations(ctx, meID, filters)
	if err != nil {
		return nil, 0, err
	}
	// Decorate each direct conversation with the counterpart's live status.
	for i := range conversations {
		if conversations[i].OtherUserID != nil {
			conversations[i].Online = s.hub.IsOnline(*conversations[i].OtherUserID)
		}
	}
	return conversations, total, nil
}

// ListMessages returns a page of history after verifying membership.
func (s *ChatService) ListMessages(ctx *fiber.Ctx, meID, conversationID, beforeID uint, limit int) ([]dtos.MessageDTO, error) {
	ok, err := s.repo.IsParticipant(ctx, conversationID, meID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbiddenConversation
	}
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	return s.repo.ListMessages(ctx, conversationID, meID, beforeID, limit)
}

// SendMessage validates, persists and fans out a message to online participants.
// Content may be empty when attachments are present.
func (s *ChatService) SendMessage(ctx *fiber.Ctx, meID, conversationID uint, content string, replyToID *uint, mentionIDs []uint, attachments []dtos.AttachmentInput) (*dtos.MessageDTO, error) {
	content = strings.TrimSpace(content)

	// Keep only well-formed attachments.
	clean := attachments[:0]
	for _, a := range attachments {
		if a.ObjectKey == "" {
			continue
		}
		if !attachmentKinds[a.Kind] {
			a.Kind = "document"
		}
		clean = append(clean, a)
	}

	if content == "" && len(clean) == 0 {
		return nil, ErrEmptyMessage
	}
	if len(content) > maxMessageLength {
		content = content[:maxMessageLength]
	}

	ok, err := s.repo.IsParticipant(ctx, conversationID, meID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbiddenConversation
	}

	// A reply must point at a message in the same conversation.
	if replyToID != nil {
		rc, _, err := s.repo.GetMessageMeta(ctx, *replyToID)
		if err != nil || rc != conversationID {
			replyToID = nil
		}
	}

	msg, err := s.repo.CreateMessage(ctx, dtos.CreateMessageInput{
		ConversationID: conversationID,
		SenderID:       meID,
		Content:        content,
		Type:           "text",
		ReplyToID:      replyToID,
		MentionIDs:     mentionIDs,
		Attachments:    clean,
	})
	if err != nil {
		return nil, err
	}

	s.broadcast(ctx, conversationID, dtos.WSEnvelope{Type: "message", Payload: msg})

	// If this is a direct chat with the AI assistant, produce a reply out of the
	// request path (the send stays fast; the reply arrives over the socket).
	if s.ai != nil {
		if aiID, isAI, _ := s.repo.GetDirectAIParticipant(ctx.Context(), conversationID, meID); isAI && meID != aiID {
			s.dispatchAIReply(conversationID, aiID, msg.ID)
		}
	}
	return msg, nil
}

// dispatchAIReply routes a reply job: to the RabbitMQ work queue when enabled
// (the consumer-service generates it), otherwise to the bounded in-process path.
// A publish failure falls back to in-process so the assistant never goes silent
// just because the broker is unavailable.
func (s *ChatService) dispatchAIReply(conversationID, aiUserID, triggerMessageID uint) {
	if s.aiQueueEnabled && s.rabbitmq != nil {
		if err := s.publishAIReplyJob(AIReplyJob{ConversationID: conversationID, AIUserID: aiUserID, TriggerMessageID: triggerMessageID}); err == nil {
			return
		} else {
			log.Printf("chat AI: publish job failed, falling back to in-process: %v", err)
		}
	}

	// Bounded in-process fallback: cap concurrent generations so a burst of
	// users can't spawn unbounded goroutines all blocked on the model.
	select {
	case s.aiSem <- struct{}{}:
		go func() {
			defer func() { <-s.aiSem }()
			if err := s.RunAIReply(conversationID, aiUserID, triggerMessageID); err != nil {
				log.Printf("chat AI reply (in-process) failed for conversation %d: %v", conversationID, err)
			}
		}()
	default:
		// Pool full — tell the user rather than leaving them hanging.
		if saved, err := s.repo.CreateAIMessage(context.Background(), conversationID, aiUserID,
			"⚠️ The assistant is busy right now. Please try again in a moment."); err == nil {
			s.broadcastCtx(context.Background(), conversationID, dtos.WSEnvelope{Type: "message", Payload: saved})
		}
	}
}

// publishAIReplyJob enqueues an AI-reply job on the work queue.
func (s *ChatService) publishAIReplyJob(job AIReplyJob) error {
	body, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return s.rabbitmq.Channel.PublishWithContext(context.Background(),
		"",               // default exchange
		chatAIReplyQueue, // routing key = queue name
		false, false,
		amqp.Publishing{DeliveryMode: amqp.Persistent, ContentType: "application/json", Body: body},
	)
}

// RunAIReply builds context from recent messages, calls the configured
// provider, and posts the assistant's Markdown reply, keeping a "typing"
// indicator alive while the model works. It is called both by the in-process
// fallback and by the consumer-service worker, so it is idempotent (a
// redelivered job whose reply already exists is a no-op) and returns an error
// only for retryable infrastructure failures — a model failure posts an apology
// and returns nil, so the job is acked rather than retried.
func (s *ChatService) RunAIReply(conversationID, aiUserID, triggerMessageID uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
	defer cancel()

	history, err := s.repo.GetRecentAIContext(ctx, conversationID, 20)
	if err != nil {
		return err // transient — let the caller retry
	}

	// Idempotency: if the assistant already answered a message at/after the one
	// that triggered this job, a duplicate (redelivered) job must do nothing.
	if triggerMessageID > 0 {
		for _, h := range history {
			if h.SenderID == aiUserID && h.ID > triggerMessageID {
				return nil
			}
		}
	}

	// Keep the "AI is typing…" indicator alive until the reply is ready.
	stop := make(chan struct{})
	go func() {
		s.emitTyping(ctx, conversationID, aiUserID)
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.emitTyping(ctx, conversationID, aiUserID)
			case <-stop:
				return
			}
		}
	}()

	images := s.loadContextImages(ctx, history, aiUserID)
	msgs := make([]ai.Message, 0, len(history))
	for _, h := range history {
		role := "user"
		if h.SenderID == aiUserID {
			role = "assistant"
		}
		msgs = append(msgs, ai.Message{Role: role, Content: h.Content, Images: images[h.ID]})
	}

	reply, genErr := s.ai.Reply(ctx, msgs)
	close(stop)

	if genErr != nil || strings.TrimSpace(reply) == "" {
		// The user only sees a generic apology, so keep the cause in the log.
		log.Printf("chat AI reply failed (conversation %d, provider %s): %v", conversationID, s.ai.Name(), genErr)
		reply = "⚠️ Sorry, I couldn't reach the AI service right now. Please try again in a moment."
	}

	saved, err := s.repo.CreateAIMessage(ctx, conversationID, aiUserID, reply)
	if err != nil {
		return err // transient — the reply was generated but not saved; retry
	}
	s.broadcastCtx(ctx, conversationID, dtos.WSEnvelope{Type: "message", Payload: saved})
	return nil
}

const (
	// A vision model re-encodes every image it is given, so a long scrollback of
	// screenshots would make each reply progressively slower for no benefit.
	// Only the newest few are worth sending.
	maxContextImages = 3
	// Base64 inflates by a third, and a photo straight off a phone can be
	// tens of megabytes; anything larger is skipped rather than stalling the reply.
	maxContextImageBytes = 6 << 20 // 6 MiB
)

// loadContextImages fetches the image attachments for a run of history, keyed
// by message id. Newest first and capped, so an old screenshot never crowds out
// the one the user just sent. Failures are skipped rather than fatal: a missing
// object should cost the picture, not the whole reply.
func (s *ChatService) loadContextImages(ctx context.Context, history []dtos.AIMessage, aiUserID uint) map[uint][]ai.Image {
	if s.minio == nil || !s.ai.SupportsImages() || len(history) == 0 {
		return nil
	}

	// Only the messages the assistant has not answered yet can carry an image
	// belonging to the current question. Fetching older ones would re-download
	// the same screenshot from object storage on every later message, and the
	// assistant discards them anyway.
	ids := unansweredMessageIDs(history, aiUserID)
	if len(ids) == 0 {
		return nil
	}
	attachments, err := s.repo.GetImageAttachments(ctx, ids)
	if err != nil {
		log.Printf("chat AI: could not list image attachments: %v", err)
		return nil
	}
	if len(attachments) == 0 {
		return nil
	}

	out := make(map[uint][]ai.Image)
	loaded := 0
	// Walk newest first so the cap keeps the most recent images.
	for i := len(attachments) - 1; i >= 0 && loaded < maxContextImages; i-- {
		a := attachments[i]
		if a.SizeBytes != nil && *a.SizeBytes > maxContextImageBytes {
			log.Printf("chat AI: skipping oversized image %s (%d bytes)", a.ObjectKey, *a.SizeBytes)
			continue
		}
		data, err := s.minio.Download(ctx, a.ObjectKey, maxContextImageBytes)
		if err != nil {
			log.Printf("chat AI: could not read image %s: %v", a.ObjectKey, err)
			continue
		}
		mime := ""
		if a.MimeType != nil {
			mime = *a.MimeType
		}
		// Prepend: the loop runs backwards, so this restores upload order.
		out[a.MessageID] = append([]ai.Image{{Data: data, MimeType: mime}}, out[a.MessageID]...)
		loaded++
	}
	if loaded > 0 {
		log.Printf("chat AI: attached %d image(s) to the assistant context", loaded)
	}
	return out
}

// unansweredMessageIDs returns the trailing run of messages sent since the
// assistant last replied — the ones that make up the question being answered.
func unansweredMessageIDs(history []dtos.AIMessage, aiUserID uint) []uint {
	var ids []uint
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].SenderID == aiUserID {
			break
		}
		ids = append(ids, history[i].ID)
	}
	return ids
}

// emitTyping broadcasts a typing signal from a specific user (the AI).
func (s *ChatService) emitTyping(ctx context.Context, conversationID, userID uint) {
	s.broadcastCtx(ctx, conversationID, dtos.WSEnvelope{
		Type:    "typing",
		Payload: map[string]interface{}{"conversation_id": conversationID, "user_id": userID},
	})
}

// broadcastCtx is broadcast for callers holding a plain context (goroutines).
func (s *ChatService) broadcastCtx(ctx context.Context, conversationID uint, env dtos.WSEnvelope) {
	userIDs, err := s.repo.GetParticipantUserIDsCtx(ctx, conversationID)
	if err != nil {
		return
	}
	payload, err := json.Marshal(env)
	if err != nil {
		return
	}
	s.hub.SendToUsers(userIDs, payload)
}

// UploadFile stores an uploaded file in object storage and returns its metadata
// (including a presigned URL for an immediate client-side preview).
func (s *ChatService) UploadFile(ctx *fiber.Ctx, fileHeader *multipart.FileHeader) (*dtos.UploadResponse, error) {
	if s.minio == nil {
		return nil, ErrUploadUnavailable
	}
	if fileHeader.Size <= 0 || fileHeader.Size > maxUploadBytes {
		return nil, ErrFileTooLarge
	}

	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	kind := kindForMime(mimeType)

	src, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	ext := filepath.Ext(fileHeader.Filename)
	objectKey := fmt.Sprintf("chat/%s/%s%s", time.Now().Format("2006/01"), uuid.NewString(), ext)

	if err := s.minio.Upload(ctx.Context(), objectKey, src, fileHeader.Size, mimeType); err != nil {
		return nil, err
	}

	download := kind == "document"
	url, err := s.minio.PresignedURL(ctx.Context(), objectKey, fileHeader.Filename, download)
	if err != nil {
		return nil, err
	}

	return &dtos.UploadResponse{
		ObjectKey: objectKey,
		Kind:      kind,
		FileName:  fileHeader.Filename,
		MimeType:  mimeType,
		SizeBytes: fileHeader.Size,
		URL:       url,
	}, nil
}

// EditMessage lets the sender edit their own (non-deleted) message.
func (s *ChatService) EditMessage(ctx *fiber.Ctx, meID, messageID uint, content string) (*dtos.MessageDTO, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, ErrEmptyMessage
	}
	if len(content) > maxMessageLength {
		content = content[:maxMessageLength]
	}
	existing, err := s.repo.GetMessageByID(ctx, meID, messageID)
	if err != nil {
		return nil, err
	}
	if existing.SenderID != meID || existing.DeletedForAll {
		return nil, ErrForbiddenAction
	}
	updated, err := s.repo.EditMessage(ctx, meID, messageID, content)
	if err != nil {
		return nil, err
	}
	s.broadcast(ctx, existing.ConversationID, dtos.WSEnvelope{Type: "message_edited", Payload: updated})
	return updated, nil
}

// React adds or removes an emoji reaction and broadcasts the new totals.
func (s *ChatService) React(ctx *fiber.Ctx, meID, messageID uint, emoji string, add bool) ([]dtos.ReactionGroupDTO, error) {
	emoji = strings.TrimSpace(emoji)
	if emoji == "" || len(emoji) > 16 {
		return nil, ErrInvalidReaction
	}
	conversationID, _, err := s.repo.GetMessageMeta(ctx, messageID)
	if err != nil {
		return nil, err
	}
	ok, err := s.repo.IsParticipant(ctx, conversationID, meID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbiddenConversation
	}
	if add {
		err = s.repo.AddReaction(ctx, messageID, meID, emoji)
	} else {
		err = s.repo.RemoveReaction(ctx, messageID, meID, emoji)
	}
	if err != nil {
		return nil, err
	}
	groups, err := s.repo.GetReactionGroups(ctx, messageID)
	if err != nil {
		return nil, err
	}
	s.broadcast(ctx, conversationID, dtos.WSEnvelope{
		Type: "reaction_updated",
		Payload: map[string]interface{}{
			"message_id":      messageID,
			"conversation_id": conversationID,
			"reactions":       groups,
		},
	})
	return groups, nil
}

// ---- Threads ----

// CreateThread branches a thread off a message (idempotent per root message),
// posts a system notice in the parent, and returns the thread conversation id.
func (s *ChatService) CreateThread(ctx *fiber.Ctx, meID, parentConvID, rootMessageID uint, title string) (uint, error) {
	ok, err := s.repo.IsParticipant(ctx, parentConvID, meID)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, ErrForbiddenConversation
	}
	// The root message must belong to the parent conversation.
	rc, _, err := s.repo.GetMessageMeta(ctx, rootMessageID)
	if err != nil || rc != parentConvID {
		return 0, ErrForbiddenAction
	}
	// One thread per message.
	if existing, found, err := s.repo.GetThreadByRoot(ctx, rootMessageID); err != nil {
		return 0, err
	} else if found {
		return existing, nil
	}

	title = strings.TrimSpace(title)
	if title == "" {
		title = "Thread"
	}
	threadID, err := s.repo.CreateThread(ctx, meID, parentConvID, rootMessageID, title)
	if err != nil {
		return 0, err
	}

	// Post a system message in the parent announcing the thread.
	sysMsg, err := s.repo.CreateMessage(ctx, dtos.CreateMessageInput{
		ConversationID: parentConvID,
		SenderID:       meID,
		Content:        "started a thread: " + title,
		Type:           "system",
	})
	if err == nil {
		s.broadcast(ctx, parentConvID, dtos.WSEnvelope{Type: "message", Payload: sysMsg})
	}

	s.broadcast(ctx, parentConvID, dtos.WSEnvelope{
		Type: "thread_created",
		Payload: map[string]interface{}{
			"parent_id":              parentConvID,
			"thread_conversation_id": threadID,
			"root_message_id":        rootMessageID,
		},
	})
	return threadID, nil
}

// ListThreads returns the threads under a conversation the caller belongs to.
func (s *ChatService) ListThreads(ctx *fiber.Ctx, meID, parentConvID uint) ([]dtos.ThreadListDTO, error) {
	ok, err := s.repo.IsParticipant(ctx, parentConvID, meID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbiddenConversation
	}
	return s.repo.ListThreads(ctx, parentConvID)
}

// GetConversation returns a single conversation's meta (used to hydrate thread
// headers, which are not present in the main list).
func (s *ChatService) GetConversation(ctx *fiber.Ctx, meID, conversationID uint) (*dtos.ConversationListDTO, error) {
	ok, err := s.repo.IsParticipant(ctx, conversationID, meID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbiddenConversation
	}
	conv, err := s.repo.GetConversationForUser(ctx, meID, conversationID)
	if err != nil {
		return nil, err
	}
	if conv.OtherUserID != nil {
		conv.Online = s.hub.IsOnline(*conv.OtherUserID)
	}
	return conv, nil
}

// MarkRead advances the user's read pointer and notifies participants.
func (s *ChatService) MarkRead(ctx *fiber.Ctx, meID, conversationID uint) error {
	ok, err := s.repo.IsParticipant(ctx, conversationID, meID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbiddenConversation
	}
	lastRead, lastReadAt, err := s.repo.MarkRead(ctx, conversationID, meID)
	if err != nil {
		return err
	}
	s.broadcast(ctx, conversationID, dtos.WSEnvelope{
		Type: "read",
		Payload: map[string]interface{}{
			"conversation_id":      conversationID,
			"user_id":              meID,
			"last_read_message_id": lastRead,
			"last_read_at":         lastReadAt,
		},
	})
	return nil
}

// UpdateConversation renames / re-describes a group or thread (owner/admin only).
func (s *ChatService) UpdateConversation(ctx *fiber.Ctx, meID, conversationID uint, title, description *string) error {
	role, ok, err := s.repo.GetParticipantRole(ctx, conversationID, meID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbiddenConversation
	}
	if role != models.RoleOwner && role != models.RoleAdmin {
		return ErrForbiddenAction
	}
	if title != nil {
		t := strings.TrimSpace(*title)
		title = &t
	}
	if err := s.repo.UpdateConversation(ctx, conversationID, title, description); err != nil {
		return err
	}
	s.broadcastConversationUpdated(ctx, conversationID)
	return nil
}

// ---- Groups & members ----

// CreateGroup creates a group with the caller as owner.
func (s *ChatService) CreateGroup(ctx *fiber.Ctx, meID uint, title string, memberIDs []uint) (uint, error) {
	title = strings.TrimSpace(title)
	if title == "" || len(memberIDs) == 0 {
		return 0, ErrEmptyGroup
	}
	convID, err := s.repo.CreateGroup(ctx, meID, title, memberIDs)
	if err != nil {
		return 0, err
	}
	s.broadcastConversationUpdated(ctx, convID)
	return convID, nil
}

// ListMembers returns the members of a conversation the caller belongs to.
func (s *ChatService) ListMembers(ctx *fiber.Ctx, meID, conversationID uint) ([]dtos.MemberDTO, error) {
	ok, err := s.repo.IsParticipant(ctx, conversationID, meID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbiddenConversation
	}
	members, err := s.repo.ListMembers(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	for i := range members {
		members[i].Online = s.hub.IsOnline(members[i].UserID)
	}
	return members, nil
}

// GetProfile returns a user's profile card, optionally scoped to a conversation
// (which the caller must belong to) to include the member's role there.
func (s *ChatService) GetProfile(ctx *fiber.Ctx, meID uint, conversationID *uint, userID uint) (*dtos.UserProfileDTO, error) {
	if conversationID != nil {
		ok, err := s.repo.IsParticipant(ctx, *conversationID, meID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrForbiddenConversation
		}
	}
	profile, err := s.repo.GetUserProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	profile.Online = s.hub.IsOnline(userID)
	if conversationID != nil {
		if role, ok, _ := s.repo.GetParticipantRole(ctx, *conversationID, userID); ok {
			profile.ConversationRole = &role
		}
	}
	return profile, nil
}

// AddMembers adds members to a group (owner/admin only).
func (s *ChatService) AddMembers(ctx *fiber.Ctx, meID, conversationID uint, memberIDs []uint) error {
	role, ok, err := s.repo.GetParticipantRole(ctx, conversationID, meID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbiddenConversation
	}
	if role != models.RoleOwner && role != models.RoleAdmin {
		return ErrForbiddenAction
	}
	if len(memberIDs) == 0 {
		return ErrEmptyGroup
	}
	if err := s.repo.AddMembers(ctx, conversationID, memberIDs); err != nil {
		return err
	}
	s.broadcastConversationUpdated(ctx, conversationID)
	return nil
}

// RemoveMember kicks a member. Owner can kick admins/members; admins can kick
// members only; the owner can never be kicked.
func (s *ChatService) RemoveMember(ctx *fiber.Ctx, meID, conversationID, targetID uint) error {
	myRole, ok, err := s.repo.GetParticipantRole(ctx, conversationID, meID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbiddenConversation
	}
	targetRole, tok, err := s.repo.GetParticipantRole(ctx, conversationID, targetID)
	if err != nil {
		return err
	}
	if !tok {
		return ErrForbiddenAction
	}
	if targetRole == models.RoleOwner {
		return ErrCannotKickOwner
	}
	switch myRole {
	case models.RoleOwner:
		// may remove admins and members
	case models.RoleAdmin:
		if targetRole != models.RoleMember {
			return ErrForbiddenAction
		}
	default:
		return ErrForbiddenAction
	}
	if err := s.repo.RemoveMember(ctx, conversationID, targetID); err != nil {
		return err
	}
	s.broadcastConversationUpdated(ctx, conversationID)
	// Notify the removed user directly (they are no longer a participant).
	s.notifyUsers([]uint{targetID}, dtos.WSEnvelope{
		Type:    "conversation_updated",
		Payload: map[string]interface{}{"conversation_id": conversationID, "removed": true},
	})
	return nil
}

// SetMemberRole changes a member's role. Only the owner may do this; setting a
// member to 'owner' transfers ownership and demotes the current owner to admin.
func (s *ChatService) SetMemberRole(ctx *fiber.Ctx, meID, conversationID, targetID uint, role string) error {
	if role != models.RoleAdmin && role != models.RoleMember && role != models.RoleOwner {
		return ErrInvalidRole
	}
	myRole, ok, err := s.repo.GetParticipantRole(ctx, conversationID, meID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbiddenConversation
	}
	if myRole != models.RoleOwner {
		return ErrForbiddenAction
	}
	if targetID == meID {
		return ErrForbiddenAction
	}
	if _, tok, err := s.repo.GetParticipantRole(ctx, conversationID, targetID); err != nil {
		return err
	} else if !tok {
		return ErrForbiddenAction
	}

	if role == models.RoleOwner {
		if err := s.repo.SetMemberRole(ctx, conversationID, targetID, models.RoleOwner); err != nil {
			return err
		}
		if err := s.repo.SetMemberRole(ctx, conversationID, meID, models.RoleAdmin); err != nil {
			return err
		}
	} else if err := s.repo.SetMemberRole(ctx, conversationID, targetID, role); err != nil {
		return err
	}
	s.broadcastConversationUpdated(ctx, conversationID)
	return nil
}

// Leave removes the caller from a conversation. An owner may only leave if an
// admin exists, in which case the earliest-joined admin is promoted to owner.
func (s *ChatService) Leave(ctx *fiber.Ctx, meID, conversationID uint) error {
	role, ok, err := s.repo.GetParticipantRole(ctx, conversationID, meID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbiddenConversation
	}
	if role == models.RoleOwner {
		adminID, hasAdmin, err := s.repo.GetOldestAdmin(ctx, conversationID)
		if err != nil {
			return err
		}
		if !hasAdmin {
			return ErrOwnerMustPromote
		}
		if err := s.repo.SetMemberRole(ctx, conversationID, adminID, models.RoleOwner); err != nil {
			return err
		}
	}
	if err := s.repo.RemoveMember(ctx, conversationID, meID); err != nil {
		return err
	}
	s.broadcastConversationUpdated(ctx, conversationID)
	return nil
}

// ---- Message deletion ----

// DeleteMessage deletes a message either just for the caller ("me") or for
// everyone. "everyone" is allowed for the sender, or a group owner/admin.
func (s *ChatService) DeleteMessage(ctx *fiber.Ctx, meID, messageID uint, scope string) (uint, error) {
	conversationID, senderID, err := s.repo.GetMessageMeta(ctx, messageID)
	if err != nil {
		return 0, err
	}
	ok, err := s.repo.IsParticipant(ctx, conversationID, meID)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, ErrForbiddenConversation
	}

	if scope == "me" {
		if err := s.repo.HideMessageForUser(ctx, messageID, meID); err != nil {
			return 0, err
		}
		return conversationID, nil
	}

	// scope == everyone
	if senderID != meID {
		role, _, _ := s.repo.GetParticipantRole(ctx, conversationID, meID)
		if role != models.RoleOwner && role != models.RoleAdmin {
			return 0, ErrForbiddenAction
		}
	}
	objectKeys, err := s.repo.DeleteMessageForEveryone(ctx, messageID, meID)
	if err != nil {
		return 0, err
	}
	// Purge attachment objects from storage (best-effort).
	if s.minio != nil {
		for _, key := range objectKeys {
			_ = s.minio.Remove(ctx.Context(), key)
		}
	}
	s.broadcast(ctx, conversationID, dtos.WSEnvelope{
		Type:    "message_deleted",
		Payload: map[string]interface{}{"conversation_id": conversationID, "message_id": messageID},
	})
	return conversationID, nil
}

// HandleTyping relays a typing signal to the other online participants. Called
// from the websocket read loop, so it takes a plain context and silently
// ignores unauthorized/expired requests.
func (s *ChatService) HandleTyping(stdctx context.Context, meID, conversationID uint) {
	ok, err := s.repo.IsParticipantCtx(stdctx, conversationID, meID)
	if err != nil || !ok {
		return
	}
	ids, err := s.repo.GetParticipantUserIDsCtx(stdctx, conversationID)
	if err != nil {
		return
	}
	targets := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id != meID {
			targets = append(targets, id)
		}
	}
	env := dtos.WSEnvelope{
		Type:    "typing",
		Payload: map[string]interface{}{"conversation_id": conversationID, "user_id": meID},
	}
	if b, err := json.Marshal(env); err == nil {
		s.hub.SendToUsers(targets, b)
	}
}

// broadcastConversationUpdated tells current participants their conversation
// (membership/roles/title) changed so they can refresh.
func (s *ChatService) broadcastConversationUpdated(ctx *fiber.Ctx, conversationID uint) {
	s.broadcast(ctx, conversationID, dtos.WSEnvelope{
		Type:    "conversation_updated",
		Payload: map[string]interface{}{"conversation_id": conversationID},
	})
}

// notifyUsers pushes an envelope to a specific set of users.
func (s *ChatService) notifyUsers(targets []uint, env dtos.WSEnvelope) {
	if b, err := json.Marshal(env); err == nil {
		s.hub.SendToUsers(targets, b)
	}
}

// broadcast serializes an envelope and delivers it to every participant that
// currently has an open socket.
func (s *ChatService) broadcast(ctx *fiber.Ctx, conversationID uint, env dtos.WSEnvelope) {
	userIDs, err := s.repo.GetParticipantUserIDs(ctx, conversationID)
	if err != nil {
		return
	}
	payload, err := json.Marshal(env)
	if err != nil {
		return
	}
	s.hub.SendToUsers(userIDs, payload)
}
