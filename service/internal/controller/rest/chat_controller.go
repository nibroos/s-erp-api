package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/nibroos/s-erp-api/service/internal/auth"
	"github.com/nibroos/s-erp-api/service/internal/chat"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/middleware"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/service"
	"github.com/nibroos/s-erp-api/service/internal/utils"
)

const (
	wsWriteWait      = 10 * time.Second
	wsPongWait       = 60 * time.Second
	wsPingPeriod     = (wsPongWait * 9) / 10
	wsMaxMessageSize = 8192
	wsSendBuffer     = 64
)

type ChatController struct {
	service *service.ChatService
	repo    *repository.ChatRepository
	hub     *chat.Hub
}

func NewChatController(service *service.ChatService, repo *repository.ChatRepository, hub *chat.Hub) *ChatController {
	return &ChatController{service: service, repo: repo, hub: hub}
}

// currentUserID extracts the authenticated user id from the JWT claims.
func currentUserID(ctx *fiber.Ctx) (uint, error) {
	claims, err := auth.GetAuthUser(ctx)
	if err != nil {
		return 0, err
	}
	uid, ok := claims["user_id"].(float64)
	if !ok || uid <= 0 {
		return 0, errors.New("invalid user id in token")
	}
	return uint(uid), nil
}

// mapServiceError maps known chat errors to HTTP status codes.
func mapServiceError(err error) int16 {
	switch {
	case errors.Is(err, service.ErrForbiddenConversation),
		errors.Is(err, service.ErrForbiddenAction),
		errors.Is(err, service.ErrCannotKickOwner):
		return http.StatusForbidden
	case errors.Is(err, service.ErrEmptyMessage),
		errors.Is(err, service.ErrCannotChatSelf),
		errors.Is(err, service.ErrOwnerMustPromote),
		errors.Is(err, service.ErrInvalidRole),
		errors.Is(err, service.ErrEmptyGroup),
		errors.Is(err, service.ErrInvalidReaction),
		errors.Is(err, service.ErrFileTooLarge):
		return http.StatusBadRequest
	case errors.Is(err, service.ErrUploadUnavailable):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// StartConversation opens/creates a direct conversation with another user.
func (c *ChatController) StartConversation(ctx *fiber.Ctx) error {
	meID, err := currentUserID(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}

	var req dtos.StartConversationRequest
	if err := ctx.BodyParser(&req); err != nil || req.UserID == nil {
		return utils.GetResponse(ctx, nil, nil, "user_id is required", http.StatusBadRequest, "user_id is required", nil)
	}

	convID, err := c.service.StartConversation(ctx, meID, *req.UserID)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to start conversation", mapServiceError(err), err.Error(), nil)
	}

	return utils.GetResponse(ctx, fiber.Map{"conversation_id": convID}, nil, "Conversation ready", http.StatusOK, nil, nil)
}

// StartAIChat opens/creates the direct conversation with the AI assistant.
func (c *ChatController) StartAIChat(ctx *fiber.Ctx) error {
	meID, err := currentUserID(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	convID, err := c.service.StartAIChat(ctx, meID)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "AI assistant unavailable", mapServiceError(err), err.Error(), nil)
	}
	return utils.GetResponse(ctx, fiber.Map{"conversation_id": convID}, nil, "AI conversation ready", http.StatusOK, nil, nil)
}

// GetConversations lists the current user's conversations with unread counts.
func (c *ChatController) GetConversations(ctx *fiber.Ctx) error {
	meID, err := currentUserID(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}

	filters, _ := ctx.Locals("filters").(map[string]string)
	if filters == nil {
		filters = map[string]string{}
	}

	conversations, total, err := c.service.ListConversations(ctx, meID, filters)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to fetch conversations", http.StatusInternalServerError, err.Error(), nil)
	}

	meta := utils.CreatePaginationMeta(filters, total)
	return utils.GetResponse(ctx, conversations, meta, "Conversations fetched", http.StatusOK, nil, nil)
}

// GetMessages returns a page of messages for a conversation.
func (c *ChatController) GetMessages(ctx *fiber.Ctx) error {
	meID, err := currentUserID(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}

	var req dtos.ListMessagesRequest
	if err := ctx.BodyParser(&req); err != nil || req.ConversationID == nil {
		return utils.GetResponse(ctx, nil, nil, "conversation_id is required", http.StatusBadRequest, "conversation_id is required", nil)
	}

	beforeID := uint(0)
	if req.BeforeID != nil {
		beforeID = *req.BeforeID
	}
	limit := 30
	if req.PerPage != nil {
		limit = *req.PerPage
	}

	messages, err := c.service.ListMessages(ctx, meID, *req.ConversationID, beforeID, limit)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to fetch messages", mapServiceError(err), err.Error(), nil)
	}

	return utils.GetResponse(ctx, messages, nil, "Messages fetched", http.StatusOK, nil, nil)
}

// SendMessage persists a message and broadcasts it to online participants.
func (c *ChatController) SendMessage(ctx *fiber.Ctx) error {
	meID, err := currentUserID(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}

	var req dtos.SendMessageRequest
	if err := ctx.BodyParser(&req); err != nil || req.ConversationID == nil {
		return utils.GetResponse(ctx, nil, nil, "conversation_id is required", http.StatusBadRequest, "conversation_id is required", nil)
	}
	content := ""
	if req.Content != nil {
		content = *req.Content
	}

	msg, err := c.service.SendMessage(ctx, meID, *req.ConversationID, content, req.ReplyToID, req.MentionIDs, req.Attachments)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to send message", mapServiceError(err), err.Error(), nil)
	}

	return utils.GetResponse(ctx, msg, nil, "Message sent", http.StatusOK, nil, nil)
}

// EditMessage edits the caller's own message.
func (c *ChatController) EditMessage(ctx *fiber.Ctx) error {
	meID, err := currentUserID(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	var req dtos.EditMessageRequest
	if err := ctx.BodyParser(&req); err != nil || req.MessageID == nil || req.Content == nil {
		return utils.GetResponse(ctx, nil, nil, "message_id and content are required", http.StatusBadRequest, "message_id and content are required", nil)
	}
	msg, err := c.service.EditMessage(ctx, meID, *req.MessageID, *req.Content)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to edit message", mapServiceError(err), err.Error(), nil)
	}
	return utils.GetResponse(ctx, msg, nil, "Message edited", http.StatusOK, nil, nil)
}

// ReactMessage adds an emoji reaction.
func (c *ChatController) ReactMessage(ctx *fiber.Ctx) error {
	return c.react(ctx, true)
}

// UnreactMessage removes an emoji reaction.
func (c *ChatController) UnreactMessage(ctx *fiber.Ctx) error {
	return c.react(ctx, false)
}

func (c *ChatController) react(ctx *fiber.Ctx, add bool) error {
	meID, err := currentUserID(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	var req dtos.ReactMessageRequest
	if err := ctx.BodyParser(&req); err != nil || req.MessageID == nil || req.Emoji == nil {
		return utils.GetResponse(ctx, nil, nil, "message_id and emoji are required", http.StatusBadRequest, "message_id and emoji are required", nil)
	}
	groups, err := c.service.React(ctx, meID, *req.MessageID, *req.Emoji, add)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to react", mapServiceError(err), err.Error(), nil)
	}
	return utils.GetResponse(ctx, groups, nil, "Reaction updated", http.StatusOK, nil, nil)
}

// CreateThread branches a thread off a message.
func (c *ChatController) CreateThread(ctx *fiber.Ctx) error {
	meID, err := currentUserID(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	var req dtos.CreateThreadRequest
	if err := ctx.BodyParser(&req); err != nil || req.ConversationID == nil || req.RootMessageID == nil {
		return utils.GetResponse(ctx, nil, nil, "conversation_id and root_message_id are required", http.StatusBadRequest, "conversation_id and root_message_id are required", nil)
	}
	title := ""
	if req.Title != nil {
		title = *req.Title
	}
	threadID, err := c.service.CreateThread(ctx, meID, *req.ConversationID, *req.RootMessageID, title)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to create thread", mapServiceError(err), err.Error(), nil)
	}
	return utils.GetResponse(ctx, fiber.Map{"conversation_id": threadID}, nil, "Thread ready", http.StatusOK, nil, nil)
}

// GetThreads lists the threads under a conversation.
func (c *ChatController) GetThreads(ctx *fiber.Ctx) error {
	meID, err := currentUserID(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	var req dtos.ListThreadsRequest
	if err := ctx.BodyParser(&req); err != nil || req.ConversationID == nil {
		return utils.GetResponse(ctx, nil, nil, "conversation_id is required", http.StatusBadRequest, "conversation_id is required", nil)
	}
	threads, err := c.service.ListThreads(ctx, meID, *req.ConversationID)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to fetch threads", mapServiceError(err), err.Error(), nil)
	}
	return utils.GetResponse(ctx, threads, nil, "Threads fetched", http.StatusOK, nil, nil)
}

// UpdateConversation renames / re-describes a group or thread.
func (c *ChatController) UpdateConversation(ctx *fiber.Ctx) error {
	meID, err := currentUserID(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	var req dtos.UpdateConversationRequest
	if err := ctx.BodyParser(&req); err != nil || req.ConversationID == nil {
		return utils.GetResponse(ctx, nil, nil, "conversation_id is required", http.StatusBadRequest, "conversation_id is required", nil)
	}
	if err := c.service.UpdateConversation(ctx, meID, *req.ConversationID, req.Title, req.Description); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to update conversation", mapServiceError(err), err.Error(), nil)
	}
	return utils.GetResponse(ctx, fiber.Map{"conversation_id": *req.ConversationID}, nil, "Conversation updated", http.StatusOK, nil, nil)
}

// GetConversation returns one conversation's meta (hydrates thread headers).
func (c *ChatController) GetConversation(ctx *fiber.Ctx) error {
	meID, err := currentUserID(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	var req dtos.ShowConversationRequest
	if err := ctx.BodyParser(&req); err != nil || req.ConversationID == nil {
		return utils.GetResponse(ctx, nil, nil, "conversation_id is required", http.StatusBadRequest, "conversation_id is required", nil)
	}
	conv, err := c.service.GetConversation(ctx, meID, *req.ConversationID)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to fetch conversation", mapServiceError(err), err.Error(), nil)
	}
	return utils.GetResponse(ctx, conv, nil, "Conversation fetched", http.StatusOK, nil, nil)
}

// ReadConversation marks a conversation as read up to its latest message.
func (c *ChatController) ReadConversation(ctx *fiber.Ctx) error {
	meID, err := currentUserID(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}

	var req dtos.ReadConversationRequest
	if err := ctx.BodyParser(&req); err != nil || req.ConversationID == nil {
		return utils.GetResponse(ctx, nil, nil, "conversation_id is required", http.StatusBadRequest, "conversation_id is required", nil)
	}

	if err := c.service.MarkRead(ctx, meID, *req.ConversationID); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to mark as read", mapServiceError(err), err.Error(), nil)
	}

	return utils.GetResponse(ctx, fiber.Map{"conversation_id": *req.ConversationID}, nil, "Marked as read", http.StatusOK, nil, nil)
}

// Upload stores a file in object storage and returns its metadata + a
// presigned URL. The client then references it in a send-message call.
func (c *ChatController) Upload(ctx *fiber.Ctx) error {
	if _, err := currentUserID(ctx); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "file is required", http.StatusBadRequest, "file is required", nil)
	}
	res, err := c.service.UploadFile(ctx, fileHeader)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to upload", mapServiceError(err), err.Error(), nil)
	}
	return utils.GetResponse(ctx, res, nil, "Uploaded", http.StatusOK, nil, nil)
}

// ---- Groups, members, profile, delete ----

// CreateGroup creates a group conversation owned by the caller.
func (c *ChatController) CreateGroup(ctx *fiber.Ctx) error {
	meID, err := currentUserID(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	var req dtos.CreateGroupRequest
	if err := ctx.BodyParser(&req); err != nil || req.Title == nil {
		return utils.GetResponse(ctx, nil, nil, "title and member_ids are required", http.StatusBadRequest, "title and member_ids are required", nil)
	}
	convID, err := c.service.CreateGroup(ctx, meID, *req.Title, req.MemberIDs)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to create group", mapServiceError(err), err.Error(), nil)
	}
	return utils.GetResponse(ctx, fiber.Map{"conversation_id": convID}, nil, "Group created", http.StatusOK, nil, nil)
}

// GetMembers lists the members of a conversation.
func (c *ChatController) GetMembers(ctx *fiber.Ctx) error {
	meID, err := currentUserID(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	var req dtos.ListMembersRequest
	if err := ctx.BodyParser(&req); err != nil || req.ConversationID == nil {
		return utils.GetResponse(ctx, nil, nil, "conversation_id is required", http.StatusBadRequest, "conversation_id is required", nil)
	}
	members, err := c.service.ListMembers(ctx, meID, *req.ConversationID)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to fetch members", mapServiceError(err), err.Error(), nil)
	}
	return utils.GetResponse(ctx, members, nil, "Members fetched", http.StatusOK, nil, nil)
}

// AddMembers invites users to a group.
func (c *ChatController) AddMembers(ctx *fiber.Ctx) error {
	meID, err := currentUserID(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	var req dtos.AddMembersRequest
	if err := ctx.BodyParser(&req); err != nil || req.ConversationID == nil {
		return utils.GetResponse(ctx, nil, nil, "conversation_id and member_ids are required", http.StatusBadRequest, "conversation_id and member_ids are required", nil)
	}
	if err := c.service.AddMembers(ctx, meID, *req.ConversationID, req.MemberIDs); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to add members", mapServiceError(err), err.Error(), nil)
	}
	return utils.GetResponse(ctx, fiber.Map{"conversation_id": *req.ConversationID}, nil, "Members added", http.StatusOK, nil, nil)
}

// RemoveMember kicks a member from a group.
func (c *ChatController) RemoveMember(ctx *fiber.Ctx) error {
	meID, err := currentUserID(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	var req dtos.RemoveMemberRequest
	if err := ctx.BodyParser(&req); err != nil || req.ConversationID == nil || req.UserID == nil {
		return utils.GetResponse(ctx, nil, nil, "conversation_id and user_id are required", http.StatusBadRequest, "conversation_id and user_id are required", nil)
	}
	if err := c.service.RemoveMember(ctx, meID, *req.ConversationID, *req.UserID); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to remove member", mapServiceError(err), err.Error(), nil)
	}
	return utils.GetResponse(ctx, fiber.Map{"conversation_id": *req.ConversationID}, nil, "Member removed", http.StatusOK, nil, nil)
}

// SetMemberRole appoints/changes a member's role (owner only).
func (c *ChatController) SetMemberRole(ctx *fiber.Ctx) error {
	meID, err := currentUserID(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	var req dtos.SetMemberRoleRequest
	if err := ctx.BodyParser(&req); err != nil || req.ConversationID == nil || req.UserID == nil || req.Role == nil {
		return utils.GetResponse(ctx, nil, nil, "conversation_id, user_id and role are required", http.StatusBadRequest, "conversation_id, user_id and role are required", nil)
	}
	if err := c.service.SetMemberRole(ctx, meID, *req.ConversationID, *req.UserID, *req.Role); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to update role", mapServiceError(err), err.Error(), nil)
	}
	return utils.GetResponse(ctx, fiber.Map{"conversation_id": *req.ConversationID}, nil, "Role updated", http.StatusOK, nil, nil)
}

// LeaveConversation removes the caller from a group.
func (c *ChatController) LeaveConversation(ctx *fiber.Ctx) error {
	meID, err := currentUserID(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	var req dtos.LeaveConversationRequest
	if err := ctx.BodyParser(&req); err != nil || req.ConversationID == nil {
		return utils.GetResponse(ctx, nil, nil, "conversation_id is required", http.StatusBadRequest, "conversation_id is required", nil)
	}
	if err := c.service.Leave(ctx, meID, *req.ConversationID); err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to leave", mapServiceError(err), err.Error(), nil)
	}
	return utils.GetResponse(ctx, fiber.Map{"conversation_id": *req.ConversationID}, nil, "Left conversation", http.StatusOK, nil, nil)
}

// GetProfile returns a contact/member profile card.
func (c *ChatController) GetProfile(ctx *fiber.Ctx) error {
	meID, err := currentUserID(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	var req dtos.MemberProfileRequest
	if err := ctx.BodyParser(&req); err != nil || req.UserID == nil {
		return utils.GetResponse(ctx, nil, nil, "user_id is required", http.StatusBadRequest, "user_id is required", nil)
	}
	profile, err := c.service.GetProfile(ctx, meID, req.ConversationID, *req.UserID)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to fetch profile", mapServiceError(err), err.Error(), nil)
	}
	return utils.GetResponse(ctx, profile, nil, "Profile fetched", http.StatusOK, nil, nil)
}

// DeleteMessage deletes a message for the caller or for everyone.
func (c *ChatController) DeleteMessage(ctx *fiber.Ctx) error {
	meID, err := currentUserID(ctx)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Unauthorized", http.StatusUnauthorized, err.Error(), nil)
	}
	var req dtos.DeleteMessageRequest
	if err := ctx.BodyParser(&req); err != nil || req.MessageID == nil {
		return utils.GetResponse(ctx, nil, nil, "message_id is required", http.StatusBadRequest, "message_id is required", nil)
	}
	scope := "me"
	if req.Scope != nil && *req.Scope == "everyone" {
		scope = "everyone"
	}
	convID, err := c.service.DeleteMessage(ctx, meID, *req.MessageID, scope)
	if err != nil {
		return utils.GetResponse(ctx, nil, nil, "Failed to delete message", mapServiceError(err), err.Error(), nil)
	}
	return utils.GetResponse(ctx, fiber.Map{"message_id": *req.MessageID, "conversation_id": convID, "scope": scope}, nil, "Message deleted", http.StatusOK, nil, nil)
}

// ---- WebSocket ----

// WSUpgradeGuard authenticates the websocket handshake. Browsers cannot set
// custom headers on a websocket request, so the JWT is accepted via the
// `token` query parameter (falling back to the Authorization header for
// non-browser clients) and validated before the connection is upgraded.
func (c *ChatController) WSUpgradeGuard(ctx *fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(ctx) {
		return fiber.ErrUpgradeRequired
	}

	token := ctx.Query("token")
	if token == "" {
		token = strings.TrimPrefix(ctx.Get("Authorization"), "Bearer ")
	}

	claims, err := middleware.VerifyJWT(token)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid or expired token"})
	}
	uid, ok := claims["user_id"].(float64)
	if !ok || uid <= 0 {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid token"})
	}

	ctx.Locals("user_id", uint(uid))
	return ctx.Next()
}

// WSHandler serves an authenticated websocket connection. It is receive-first:
// the server pushes new messages / read receipts to the client, while message
// creation stays on the REST endpoints (validated, traced, transactional).
// The read loop keeps the connection alive and detects disconnects.
func (c *ChatController) WSHandler(conn *websocket.Conn) {
	userID, _ := conn.Locals("user_id").(uint)
	if userID == 0 {
		_ = conn.Close()
		return
	}

	client := &chat.Client{UserID: userID, Send: make(chan []byte, wsSendBuffer)}
	becameOnline := c.hub.Add(client)

	done := make(chan struct{})

	// Write pump: drain the send channel and keep the socket warm with pings.
	go func() {
		ticker := time.NewTicker(wsPingPeriod)
		defer ticker.Stop()
		for {
			select {
			case msg, ok := <-client.Send:
				_ = conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
				if !ok {
					_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					return
				}
			case <-ticker.C:
				_ = conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			case <-done:
				return
			}
		}
	}()

	// Announce presence to contacts and send this client a snapshot of which
	// of its contacts are currently online.
	c.emitPresenceOnConnect(userID, becameOnline)

	// Read pump: enforce limits, refresh deadlines on pong, detect close.
	conn.SetReadLimit(wsMaxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(wsPongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(wsPongWait))
	})

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			break
		}
		_ = conn.SetReadDeadline(time.Now().Add(wsPongWait))
		c.handleInbound(userID, raw)
	}

	close(done)
	becameOffline := c.hub.Remove(client)
	if becameOffline {
		c.broadcastPresence(userID, false)
	}
	_ = conn.Close()
}

// handleInbound processes a client-sent websocket frame. Currently only typing
// signals are accepted; everything else is treated as a keepalive.
func (c *ChatController) handleInbound(userID uint, raw []byte) {
	if len(raw) == 0 {
		return
	}
	var in dtos.WSInbound
	if err := json.Unmarshal(raw, &in); err != nil {
		return
	}
	switch in.Type {
	case "typing":
		if in.ConversationID != nil {
			c.service.HandleTyping(context.Background(), userID, *in.ConversationID)
		}
	}
}

// emitPresenceOnConnect announces the user as online to their contacts (only on
// the first connection) and pushes the connecting client a snapshot of which of
// its contacts are already online.
func (c *ChatController) emitPresenceOnConnect(userID uint, becameOnline bool) {
	contacts, err := c.repo.GetContactUserIDs(context.Background(), userID)
	if err != nil {
		contacts = nil
	}
	if becameOnline {
		c.broadcastPresenceTo(contacts, userID, true)
	}

	online := c.hub.FilterOnline(contacts)
	env := dtos.WSEnvelope{
		Type:    "presence_snapshot",
		Payload: fiber.Map{"online_user_ids": online},
	}
	if b, err := json.Marshal(env); err == nil {
		c.hub.SendToUsers([]uint{userID}, b)
	}
}

// broadcastPresence tells all of the user's contacts about their online state.
func (c *ChatController) broadcastPresence(userID uint, online bool) {
	contacts, err := c.repo.GetContactUserIDs(context.Background(), userID)
	if err != nil {
		return
	}
	c.broadcastPresenceTo(contacts, userID, online)
}

func (c *ChatController) broadcastPresenceTo(targets []uint, userID uint, online bool) {
	if len(targets) == 0 {
		return
	}
	env := dtos.WSEnvelope{
		Type:    "presence",
		Payload: fiber.Map{"user_id": userID, "online": online},
	}
	if b, err := json.Marshal(env); err == nil {
		c.hub.SendToUsers(targets, b)
	}
}
