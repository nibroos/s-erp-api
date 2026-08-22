package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/nibroos/s-erp-api/service/internal/config"
	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/models"
	"github.com/nibroos/s-erp-api/service/internal/utils"
	"gorm.io/gorm"
)

type ChatRepository struct {
	db    *gorm.DB
	sqlDB *sqlx.DB
	minio *config.MinioStorage
}

func NewChatRepository(db *gorm.DB, sqlDB *sqlx.DB, minio *config.MinioStorage) *ChatRepository {
	return &ChatRepository{db: db, sqlDB: sqlDB, minio: minio}
}

func (r *ChatRepository) BeginTransaction() *gorm.DB {
	return r.db.Begin()
}

// GetOrCreateDirectConversation returns the id of the 1-1 conversation between
// userA and userB, creating it (and its two participant rows) if it does not
// exist yet. A deterministic direct_key guarded by a unique index guarantees a
// single conversation per pair.
func (r *ChatRepository) GetOrCreateDirectConversation(ctx *fiber.Ctx, userA, userB uint) (uint, error) {
	low, high := userA, userB
	if low > high {
		low, high = high, low
	}
	key := fmt.Sprintf("%d:%d", low, high)

	var convID uint
	err := r.db.WithContext(ctx.Context()).Transaction(func(tx *gorm.DB) error {
		var existing models.Conversation
		res := tx.Where("direct_key = ? AND deleted_at IS NULL", key).First(&existing)
		if res.Error == nil {
			convID = existing.ID
			return nil
		}
		if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return res.Error
		}

		convType := "direct"
		conv := models.Conversation{
			Type:        convType,
			DirectKey:   &key,
			CreatedByID: &userA,
		}
		if err := tx.Create(&conv).Error; err != nil {
			return err
		}
		convID = conv.ID

		parts := []models.ConversationParticipant{
			{ConversationID: conv.ID, UserID: low},
			{ConversationID: conv.ID, UserID: high},
		}
		if err := tx.Create(&parts).Error; err != nil {
			return err
		}
		return nil
	})

	return convID, err
}

// IsParticipant reports whether userID belongs to the given conversation.
// It is the authorization gate for every read/write on a conversation.
func (r *ChatRepository) IsParticipant(ctx *fiber.Ctx, conversationID, userID uint) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM conversation_participants
		WHERE conversation_id = $1 AND user_id = $2 AND deleted_at IS NULL`
	if err := r.sqlDB.GetContext(ctx.Context(), &count, query, conversationID, userID); err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetContactUserIDs returns the distinct user ids that share at least one
// conversation with userID (i.e. the people whose presence userID should see
// and be seen by). It takes a plain context so it can be called from the
// websocket handler, which has no *fiber.Ctx.
func (r *ChatRepository) GetContactUserIDs(ctx context.Context, userID uint) ([]uint, error) {
	ids := []uint{}
	query := `SELECT DISTINCT cp2.user_id
		FROM conversation_participants cp1
		JOIN conversation_participants cp2 ON cp2.conversation_id = cp1.conversation_id
		WHERE cp1.user_id = $1 AND cp2.user_id <> $1
			AND cp1.deleted_at IS NULL AND cp2.deleted_at IS NULL`
	if err := r.sqlDB.SelectContext(ctx, &ids, query, userID); err != nil {
		return nil, err
	}
	return ids, nil
}

// GetParticipantUserIDs returns every user id in a conversation (used to fan a
// new message out to the online participants).
func (r *ChatRepository) GetParticipantUserIDs(ctx *fiber.Ctx, conversationID uint) ([]uint, error) {
	ids := []uint{}
	query := `SELECT user_id FROM conversation_participants
		WHERE conversation_id = $1 AND deleted_at IS NULL`
	if err := r.sqlDB.SelectContext(ctx.Context(), &ids, query, conversationID); err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *ChatRepository) ListConversations(ctx *fiber.Ctx, userID uint, filters map[string]string) ([]dtos.ConversationListDTO, int, error) {
	conversations := []dtos.ConversationListDTO{}
	var total int

	base := `FROM conversation_participants cp
		JOIN conversations c ON c.id = cp.conversation_id AND c.deleted_at IS NULL
		LEFT JOIN messages lm ON lm.id = c.last_message_id
		LEFT JOIN LATERAL (
			SELECT cp2.user_id, cp2.last_read_message_id, cp2.last_read_at FROM conversation_participants cp2
			WHERE cp2.conversation_id = c.id AND cp2.user_id <> $1
				AND cp2.deleted_at IS NULL AND c.type = 'direct'
			LIMIT 1
		) ou ON true
		LEFT JOIN users u ON u.id = ou.user_id
		WHERE cp.user_id = $1 AND cp.deleted_at IS NULL AND c.parent_id IS NULL`

	args := []interface{}{userID}
	i := 2

	if global, ok := filters["global"]; ok && global != "" {
		// Match on the counterpart (name/email), the group title, or the text
		// of ANY message in the conversation.
		base += fmt.Sprintf(
			" AND (u.name ILIKE $%d OR u.email ILIKE $%d OR c.title ILIKE $%d"+
				" OR EXISTS (SELECT 1 FROM messages mm WHERE mm.conversation_id = c.id"+
				" AND mm.deleted_at IS NULL AND mm.content ILIKE $%d))",
			i, i, i, i,
		)
		args = append(args, "%"+global+"%")
		i++
	}

	selectCols := `SELECT c.id, c.type, c.title, c.description,
		c.last_message_at::text AS last_message_at,
		CASE WHEN lm.deleted_for_all THEN 'message deleted' ELSE lm.content END AS last_message_text,
		(SELECT COUNT(*) FROM messages m2
			WHERE m2.conversation_id = c.id
			AND m2.deleted_at IS NULL
			AND m2.sender_id <> $1
			AND (cp.last_read_message_id IS NULL OR m2.id > cp.last_read_message_id)
		) AS unread_count,
		cp.role AS my_role,
		c.parent_id,
		(SELECT COUNT(*) FROM conversation_participants cpc
			WHERE cpc.conversation_id = c.id AND cpc.deleted_at IS NULL
		) AS member_count,
		ou.user_id AS other_user_id,
		u.name AS other_user_name,
		u.email AS other_user_email,
		u.profile_image_url AS other_user_image,
			COALESCE(u.is_ai, false) AS other_user_is_ai,
		ou.last_read_message_id AS other_user_last_read,
		ou.last_read_at::text AS other_user_last_read_at `

	countQuery := `SELECT COUNT(*) ` + base
	countArgs := append([]interface{}{}, args...)
	if err := r.sqlDB.GetContext(ctx.Context(), &total, countQuery, countArgs...); err != nil {
		return nil, 0, err
	}

	query := selectCols + base
	query += " ORDER BY c.last_message_at DESC NULLS LAST, c.id DESC"

	perPage := utils.GetIntOrDefault(filters["per_page"], 20)
	page := utils.GetIntOrDefault(filters["page"], 1)
	offset := (page - 1) * perPage
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", i, i+1)
	args = append(args, perPage, offset)

	if err := r.sqlDB.SelectContext(ctx.Context(), &conversations, query, args...); err != nil {
		return nil, 0, err
	}

	return conversations, total, nil
}

// ListMessages returns messages for a conversation in ascending id order,
// using keyset pagination (before_id) for efficient infinite scroll. Messages
// the user hid ("delete for me") are excluded; messages deleted for everyone
// are returned with blanked content and the deleted_for_all flag set.
func (r *ChatRepository) ListMessages(ctx *fiber.Ctx, conversationID, meID uint, beforeID uint, limit int) ([]dtos.MessageDTO, error) {
	messages := []dtos.MessageDTO{}

	query := `SELECT m.id, m.conversation_id, m.sender_id, u.name AS sender_name, u.is_ai AS sender_is_ai,
			CASE WHEN m.deleted_for_all THEN '' ELSE m.content END AS content,
			m.type, m.deleted_for_all, m.edited_at::text AS edited_at, m.reply_to_id,
			m.created_at::text AS created_at
		FROM messages m
		JOIN users u ON u.id = m.sender_id
		WHERE m.conversation_id = $1 AND m.deleted_at IS NULL
			AND NOT EXISTS (
				SELECT 1 FROM message_hides mh
				WHERE mh.message_id = m.id AND mh.user_id = $2
			)`
	args := []interface{}{conversationID, meID}
	i := 3
	if beforeID > 0 {
		query += fmt.Sprintf(" AND m.id < $%d", i)
		args = append(args, beforeID)
		i++
	}
	query += fmt.Sprintf(" ORDER BY m.id DESC LIMIT $%d", i)
	args = append(args, limit)

	if err := r.sqlDB.SelectContext(ctx.Context(), &messages, query, args...); err != nil {
		return nil, err
	}

	// Reverse into ascending order for natural display.
	for l, rgt := 0, len(messages)-1; l < rgt; l, rgt = l+1, rgt-1 {
		messages[l], messages[rgt] = messages[rgt], messages[l]
	}

	if err := r.enrichMessages(ctx, meID, messages); err != nil {
		return nil, err
	}

	return messages, nil
}

// GetMessageByID returns one fully-enriched message for the given user.
func (r *ChatRepository) GetMessageByID(ctx *fiber.Ctx, meID, messageID uint) (*dtos.MessageDTO, error) {
	result := dtos.MessageDTO{}
	query := `SELECT m.id, m.conversation_id, m.sender_id, u.name AS sender_name, u.is_ai AS sender_is_ai,
			CASE WHEN m.deleted_for_all THEN '' ELSE m.content END AS content,
			m.type, m.deleted_for_all, m.edited_at::text AS edited_at, m.reply_to_id,
			m.created_at::text AS created_at
		FROM messages m JOIN users u ON u.id = m.sender_id
		WHERE m.id = $1`
	if err := r.sqlDB.GetContext(ctx.Context(), &result, query, messageID); err != nil {
		return nil, err
	}
	batch := []dtos.MessageDTO{result}
	if err := r.enrichMessages(ctx, meID, batch); err != nil {
		return nil, err
	}
	return &batch[0], nil
}

// enrichMessages populates reply previews, mentions, reactions and thread info
// for a batch of messages using a handful of set-based queries.
func (r *ChatRepository) enrichMessages(ctx *fiber.Ctx, meID uint, messages []dtos.MessageDTO) error {
	if len(messages) == 0 {
		return nil
	}
	c := ctx.Context()

	ids := make([]uint, 0, len(messages))
	replyIDs := make([]uint, 0)
	index := make(map[uint]int, len(messages))
	for idx := range messages {
		messages[idx].Mentions = []dtos.MentionDTO{}
		messages[idx].Reactions = []dtos.ReactionGroupDTO{}
		ids = append(ids, messages[idx].ID)
		index[messages[idx].ID] = idx
		if messages[idx].ReplyToID != nil {
			replyIDs = append(replyIDs, *messages[idx].ReplyToID)
		}
	}

	// Reactions grouped by message + emoji.
	reactionRows := []struct {
		MessageID uint          `db:"message_id"`
		Emoji     string        `db:"emoji"`
		UserIDs   pq.Int64Array `db:"user_ids"`
	}{}
	if err := r.sqlDB.SelectContext(c, &reactionRows,
		`SELECT message_id, emoji, array_agg(user_id ORDER BY id) AS user_ids
		 FROM message_reactions WHERE message_id = ANY($1)
		 GROUP BY message_id, emoji ORDER BY message_id, emoji`,
		pq.Array(ids)); err != nil {
		return err
	}
	for _, row := range reactionRows {
		idx, ok := index[row.MessageID]
		if !ok {
			continue
		}
		userIDs := make([]uint, 0, len(row.UserIDs))
		reacted := false
		for _, u := range row.UserIDs {
			userIDs = append(userIDs, uint(u))
			if uint(u) == meID {
				reacted = true
			}
		}
		messages[idx].Reactions = append(messages[idx].Reactions, dtos.ReactionGroupDTO{
			Emoji:   row.Emoji,
			Count:   len(userIDs),
			UserIDs: userIDs,
			Reacted: reacted,
		})
	}

	// Mentions.
	mentionRows := []struct {
		MessageID uint   `db:"message_id"`
		UserID    uint   `db:"user_id"`
		Name      string `db:"name"`
	}{}
	if err := r.sqlDB.SelectContext(c, &mentionRows,
		`SELECT mm.message_id, mm.user_id, u.name
		 FROM message_mentions mm JOIN users u ON u.id = mm.user_id
		 WHERE mm.message_id = ANY($1)`, pq.Array(ids)); err != nil {
		return err
	}
	for _, row := range mentionRows {
		if idx, ok := index[row.MessageID]; ok {
			messages[idx].Mentions = append(messages[idx].Mentions, dtos.MentionDTO{
				UserID: row.UserID, Name: row.Name,
			})
		}
	}

	// Reply previews.
	if len(replyIDs) > 0 {
		replyRows := []dtos.ReplyPreviewDTO{}
		if err := r.sqlDB.SelectContext(c, &replyRows,
			`SELECT m.id, u.name AS sender_name,
				CASE WHEN m.deleted_for_all THEN '' ELSE m.content END AS content,
				m.deleted_for_all
			 FROM messages m JOIN users u ON u.id = m.sender_id
			 WHERE m.id = ANY($1)`, pq.Array(replyIDs)); err != nil {
			return err
		}
		byID := make(map[uint]dtos.ReplyPreviewDTO, len(replyRows))
		for _, rw := range replyRows {
			byID[rw.ID] = rw
		}
		for idx := range messages {
			if messages[idx].ReplyToID != nil {
				if p, ok := byID[*messages[idx].ReplyToID]; ok {
					pc := p
					messages[idx].ReplyTo = &pc
				}
			}
		}
	}

	// Threads branched from these messages.
	threadRows := []struct {
		RootMessageID uint    `db:"root_message_id"`
		ConvID        uint    `db:"conv_id"`
		Title         *string `db:"title"`
		ReplyCount    int     `db:"reply_count"`
	}{}
	if err := r.sqlDB.SelectContext(c, &threadRows,
		`SELECT c.root_message_id, c.id AS conv_id, c.title,
			(SELECT COUNT(*) FROM messages tm
			 WHERE tm.conversation_id = c.id AND tm.deleted_at IS NULL AND tm.type <> 'system') AS reply_count
		 FROM conversations c
		 WHERE c.type = 'thread' AND c.deleted_at IS NULL AND c.root_message_id = ANY($1)`,
		pq.Array(ids)); err != nil {
		return err
	}
	for _, row := range threadRows {
		if idx, ok := index[row.RootMessageID]; ok {
			cid := row.ConvID
			messages[idx].ThreadConversationID = &cid
			messages[idx].ThreadReplyCount = row.ReplyCount
			messages[idx].ThreadTitle = row.Title
		}
	}

	// Attachments (with a presigned URL per object).
	for idx := range messages {
		messages[idx].Attachments = []dtos.AttachmentDTO{}
	}
	attachments := []dtos.AttachmentDTO{}
	if err := r.sqlDB.SelectContext(c, &attachments,
		`SELECT id, message_id, kind, object_key, file_name, mime_type, size_bytes,
			duration_seconds, width, height, description
		 FROM message_attachments WHERE message_id = ANY($1) ORDER BY id`,
		pq.Array(ids)); err != nil {
		return err
	}
	for i := range attachments {
		idx, ok := index[attachments[i].MessageID]
		if !ok || messages[idx].DeletedForAll {
			continue // never surface attachments for a deleted message
		}
		if r.minio != nil {
			download := attachments[i].Kind == models.AttachmentDocument
			fileName := ""
			if attachments[i].FileName != nil {
				fileName = *attachments[i].FileName
			}
			if u, err := r.minio.PresignedURL(c, attachments[i].ObjectKey, fileName, download); err == nil {
				attachments[i].URL = u
			}
		}
		messages[idx].Attachments = append(messages[idx].Attachments, attachments[i])
	}

	return nil
}

// CreateMessage persists a message (with optional reply + mentions), bumps the
// conversation's last-message pointer and marks the sender as having read it,
// all in one transaction. It returns the fully-enriched stored message.
func (r *ChatRepository) CreateMessage(ctx *fiber.Ctx, in dtos.CreateMessageInput) (*dtos.MessageDTO, error) {
	msgType := in.Type
	if msgType == "" {
		msgType = "text"
	}
	msg := models.Message{
		ConversationID: in.ConversationID,
		SenderID:       in.SenderID,
		Content:        in.Content,
		Type:           msgType,
		ReplyToID:      in.ReplyToID,
	}

	err := r.db.WithContext(ctx.Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&msg).Error; err != nil {
			return err
		}
		for _, uid := range in.MentionIDs {
			if uid == 0 {
				continue
			}
			if err := tx.Exec(
				`INSERT INTO message_mentions (message_id, user_id) VALUES (?, ?)
				 ON CONFLICT (message_id, user_id) DO NOTHING`, msg.ID, uid).Error; err != nil {
				return err
			}
		}
		for _, a := range in.Attachments {
			if a.ObjectKey == "" {
				continue
			}
			att := models.MessageAttachment{
				MessageID:       msg.ID,
				Kind:            a.Kind,
				ObjectKey:       a.ObjectKey,
				FileName:        a.FileName,
				MimeType:        a.MimeType,
				SizeBytes:       a.SizeBytes,
				DurationSeconds: a.DurationSeconds,
				Width:           a.Width,
				Height:          a.Height,
				Description:     a.Description,
			}
			if err := tx.Create(&att).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&models.Conversation{}).
			Where("id = ?", in.ConversationID).
			Updates(map[string]interface{}{
				"last_message_id": msg.ID,
				"last_message_at": gorm.Expr("now()"),
				"updated_at":      gorm.Expr("now()"),
			}).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.ConversationParticipant{}).
			Where("conversation_id = ? AND user_id = ?", in.ConversationID, in.SenderID).
			Update("last_read_message_id", msg.ID).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return r.GetMessageByID(ctx, in.SenderID, msg.ID)
}

// EditMessage updates a message's content and stamps edited_at.
func (r *ChatRepository) EditMessage(ctx *fiber.Ctx, meID, messageID uint, content string) (*dtos.MessageDTO, error) {
	if _, err := r.sqlDB.ExecContext(ctx.Context(),
		`UPDATE messages SET content = $2, edited_at = now(), updated_at = now()
		 WHERE id = $1 AND deleted_at IS NULL`, messageID, content); err != nil {
		return nil, err
	}
	return r.GetMessageByID(ctx, meID, messageID)
}

// ---- Reactions ----

func (r *ChatRepository) AddReaction(ctx *fiber.Ctx, messageID, userID uint, emoji string) error {
	_, err := r.sqlDB.ExecContext(ctx.Context(),
		`INSERT INTO message_reactions (message_id, user_id, emoji) VALUES ($1, $2, $3)
		 ON CONFLICT (message_id, user_id, emoji) DO NOTHING`, messageID, userID, emoji)
	return err
}

func (r *ChatRepository) RemoveReaction(ctx *fiber.Ctx, messageID, userID uint, emoji string) error {
	_, err := r.sqlDB.ExecContext(ctx.Context(),
		`DELETE FROM message_reactions WHERE message_id = $1 AND user_id = $2 AND emoji = $3`,
		messageID, userID, emoji)
	return err
}

// GetReactionGroups returns the aggregated reactions for a message (UserIDs let
// each client derive its own "reacted" flag).
func (r *ChatRepository) GetReactionGroups(ctx *fiber.Ctx, messageID uint) ([]dtos.ReactionGroupDTO, error) {
	rows := []struct {
		Emoji   string        `db:"emoji"`
		UserIDs pq.Int64Array `db:"user_ids"`
	}{}
	if err := r.sqlDB.SelectContext(ctx.Context(), &rows,
		`SELECT emoji, array_agg(user_id ORDER BY id) AS user_ids
		 FROM message_reactions WHERE message_id = $1 GROUP BY emoji ORDER BY emoji`,
		messageID); err != nil {
		return nil, err
	}
	groups := make([]dtos.ReactionGroupDTO, 0, len(rows))
	for _, row := range rows {
		ids := make([]uint, 0, len(row.UserIDs))
		for _, u := range row.UserIDs {
			ids = append(ids, uint(u))
		}
		groups = append(groups, dtos.ReactionGroupDTO{
			Emoji: row.Emoji, Count: len(ids), UserIDs: ids,
		})
	}
	return groups, nil
}

// ---- Threads ----

// CreateThread creates a thread conversation branched from a message and copies
// the parent's participants into it (creator becomes owner).
func (r *ChatRepository) CreateThread(ctx *fiber.Ctx, meID, parentID, rootMessageID uint, title string) (uint, error) {
	var threadID uint
	err := r.db.WithContext(ctx.Context()).Transaction(func(tx *gorm.DB) error {
		conv := models.Conversation{
			Type:          "thread",
			Title:         &title,
			ParentID:      &parentID,
			RootMessageID: &rootMessageID,
			CreatedByID:   &meID,
		}
		if err := tx.Create(&conv).Error; err != nil {
			return err
		}
		threadID = conv.ID
		return tx.Exec(
			`INSERT INTO conversation_participants (conversation_id, user_id, role, created_at, updated_at)
			 SELECT ?, user_id, CASE WHEN user_id = ? THEN 'owner' ELSE 'member' END, now(), now()
			 FROM conversation_participants WHERE conversation_id = ? AND deleted_at IS NULL`,
			conv.ID, meID, parentID).Error
	})
	return threadID, err
}

// ListThreads returns the threads under a parent conversation.
func (r *ChatRepository) ListThreads(ctx *fiber.Ctx, parentID uint) ([]dtos.ThreadListDTO, error) {
	threads := []dtos.ThreadListDTO{}
	query := `SELECT c.id AS conversation_id, c.title, c.root_message_id,
			CASE WHEN rm.deleted_for_all THEN 'message deleted' ELSE left(rm.content, 80) END AS root_snippet,
			(SELECT COUNT(*) FROM messages tm
			 WHERE tm.conversation_id = c.id AND tm.deleted_at IS NULL AND tm.type <> 'system') AS reply_count,
			c.last_message_at::text AS last_message_at
		FROM conversations c
		LEFT JOIN messages rm ON rm.id = c.root_message_id
		WHERE c.parent_id = $1 AND c.type = 'thread' AND c.deleted_at IS NULL
		ORDER BY c.last_message_at DESC NULLS LAST, c.id DESC`
	if err := r.sqlDB.SelectContext(ctx.Context(), &threads, query, parentID); err != nil {
		return nil, err
	}
	return threads, nil
}

// GetThreadByRoot returns the existing thread for a root message, if any.
func (r *ChatRepository) GetThreadByRoot(ctx *fiber.Ctx, rootMessageID uint) (uint, bool, error) {
	var id uint
	err := r.sqlDB.GetContext(ctx.Context(), &id,
		`SELECT id FROM conversations
		 WHERE root_message_id = $1 AND type = 'thread' AND deleted_at IS NULL LIMIT 1`,
		rootMessageID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return id, true, nil
}

// GetConversationForUser returns one conversation's meta for a participant
// (works for threads too, which are excluded from the main list).
func (r *ChatRepository) GetConversationForUser(ctx *fiber.Ctx, meID, conversationID uint) (*dtos.ConversationListDTO, error) {
	conv := dtos.ConversationListDTO{}
	query := `SELECT c.id, c.type, c.title, c.description,
			c.last_message_at::text AS last_message_at,
			CASE WHEN lm.deleted_for_all THEN 'message deleted' ELSE lm.content END AS last_message_text,
			0 AS unread_count,
			cp.role AS my_role,
			c.parent_id,
			(SELECT COUNT(*) FROM conversation_participants cpc
				WHERE cpc.conversation_id = c.id AND cpc.deleted_at IS NULL) AS member_count,
			ou.user_id AS other_user_id,
			u.name AS other_user_name,
			u.email AS other_user_email,
			u.profile_image_url AS other_user_image,
			COALESCE(u.is_ai, false) AS other_user_is_ai,
			ou.last_read_message_id AS other_user_last_read,
			ou.last_read_at::text AS other_user_last_read_at
		FROM conversation_participants cp
		JOIN conversations c ON c.id = cp.conversation_id AND c.deleted_at IS NULL
		LEFT JOIN messages lm ON lm.id = c.last_message_id
		LEFT JOIN LATERAL (
			SELECT cp2.user_id, cp2.last_read_message_id, cp2.last_read_at FROM conversation_participants cp2
			WHERE cp2.conversation_id = c.id AND cp2.user_id <> $1
				AND cp2.deleted_at IS NULL AND c.type = 'direct'
			LIMIT 1
		) ou ON true
		LEFT JOIN users u ON u.id = ou.user_id
		WHERE cp.user_id = $1 AND cp.deleted_at IS NULL AND c.id = $2`
	if err := r.sqlDB.GetContext(ctx.Context(), &conv, query, meID, conversationID); err != nil {
		return nil, err
	}
	return &conv, nil
}

// ---- Groups & members ----

// CreateGroup creates a group conversation with the creator as owner and the
// given users as members.
func (r *ChatRepository) CreateGroup(ctx *fiber.Ctx, ownerID uint, title string, memberIDs []uint) (uint, error) {
	var convID uint
	err := r.db.WithContext(ctx.Context()).Transaction(func(tx *gorm.DB) error {
		conv := models.Conversation{Type: "group", Title: &title, CreatedByID: &ownerID}
		if err := tx.Create(&conv).Error; err != nil {
			return err
		}
		convID = conv.ID

		parts := []models.ConversationParticipant{
			{ConversationID: conv.ID, UserID: ownerID, Role: models.RoleOwner},
		}
		seen := map[uint]bool{ownerID: true}
		for _, m := range memberIDs {
			if m == 0 || seen[m] {
				continue
			}
			seen[m] = true
			parts = append(parts, models.ConversationParticipant{
				ConversationID: conv.ID, UserID: m, Role: models.RoleMember,
			})
		}
		return tx.Create(&parts).Error
	})
	return convID, err
}

// ListMembers returns all active members of a conversation, owner/admin first.
func (r *ChatRepository) ListMembers(ctx *fiber.Ctx, conversationID uint) ([]dtos.MemberDTO, error) {
	members := []dtos.MemberDTO{}
	query := `SELECT cp.user_id, u.name, u.email, u.phone_number, u.profile_image_url,
			cp.role, u.is_ai, cp.last_read_message_id, cp.last_read_at::text AS last_read_at
		FROM conversation_participants cp
		JOIN users u ON u.id = cp.user_id
		WHERE cp.conversation_id = $1 AND cp.deleted_at IS NULL
		ORDER BY CASE cp.role WHEN 'owner' THEN 0 WHEN 'admin' THEN 1 ELSE 2 END, u.name`
	if err := r.sqlDB.SelectContext(ctx.Context(), &members, query, conversationID); err != nil {
		return nil, err
	}
	return members, nil
}

// GetParticipantRole returns a user's role in a conversation and whether they
// are an active participant.
func (r *ChatRepository) GetParticipantRole(ctx *fiber.Ctx, conversationID, userID uint) (string, bool, error) {
	var role string
	query := `SELECT role FROM conversation_participants
		WHERE conversation_id = $1 AND user_id = $2 AND deleted_at IS NULL`
	err := r.sqlDB.GetContext(ctx.Context(), &role, query, conversationID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	return role, true, nil
}

// AddMembers inserts members, restoring anyone previously removed.
func (r *ChatRepository) AddMembers(ctx *fiber.Ctx, conversationID uint, memberIDs []uint) error {
	return r.db.WithContext(ctx.Context()).Transaction(func(tx *gorm.DB) error {
		for _, m := range memberIDs {
			if m == 0 {
				continue
			}
			if err := tx.Exec(`INSERT INTO conversation_participants (conversation_id, user_id, role, created_at, updated_at)
				VALUES (?, ?, 'member', now(), now())
				ON CONFLICT (conversation_id, user_id)
				DO UPDATE SET deleted_at = NULL, role = 'member', updated_at = now()`,
				conversationID, m).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// RemoveMember soft-deletes a participant (kick or leave).
func (r *ChatRepository) RemoveMember(ctx *fiber.Ctx, conversationID, userID uint) error {
	_, err := r.sqlDB.ExecContext(ctx.Context(),
		`UPDATE conversation_participants SET deleted_at = now(), updated_at = now()
		 WHERE conversation_id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		conversationID, userID)
	return err
}

// SetMemberRole updates a participant's role.
func (r *ChatRepository) SetMemberRole(ctx *fiber.Ctx, conversationID, userID uint, role string) error {
	_, err := r.sqlDB.ExecContext(ctx.Context(),
		`UPDATE conversation_participants SET role = $3, updated_at = now()
		 WHERE conversation_id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		conversationID, userID, role)
	return err
}

// CountRole counts active participants with the given role.
func (r *ChatRepository) CountRole(ctx *fiber.Ctx, conversationID uint, role string) (int, error) {
	var n int
	err := r.sqlDB.GetContext(ctx.Context(), &n,
		`SELECT COUNT(*) FROM conversation_participants
		 WHERE conversation_id = $1 AND role = $2 AND deleted_at IS NULL`,
		conversationID, role)
	return n, err
}

// GetOldestAdmin returns the earliest-joined admin, if any.
func (r *ChatRepository) GetOldestAdmin(ctx *fiber.Ctx, conversationID uint) (uint, bool, error) {
	var uid uint
	err := r.sqlDB.GetContext(ctx.Context(), &uid,
		`SELECT user_id FROM conversation_participants
		 WHERE conversation_id = $1 AND role = 'admin' AND deleted_at IS NULL
		 ORDER BY id ASC LIMIT 1`, conversationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return uid, true, nil
}

// ---- Profiles ----

// GetUserProfile returns a user's profile card plus their application roles.
func (r *ChatRepository) GetUserProfile(ctx *fiber.Ctx, userID uint) (*dtos.UserProfileDTO, error) {
	profile := dtos.UserProfileDTO{}
	if err := r.sqlDB.GetContext(ctx.Context(), &profile,
		`SELECT id, name, email, status, phone_number, address, profile_image_url, is_ai
		 FROM users WHERE id = $1`, userID); err != nil {
		return nil, err
	}
	// Roles are modelled through the pools <-> mix_values pivot (same query the
	// login flow uses), not a dedicated roles table.
	roles := []string{}
	if err := r.sqlDB.SelectContext(ctx.Context(), &roles,
		`SELECT mv.name
		 FROM pools p
		 JOIN mix_values mv ON p.mv2_id = mv.id
		 JOIN groups g1 ON p.group1_id = g1.id
		 JOIN groups g2 ON p.group2_id = g2.id
		 WHERE p.deleted_at IS NULL AND g1.name = 'users' AND g2.name = 'roles'
		 AND p.mv1_id = $1`, userID); err == nil {
		profile.Roles = roles
	}
	return &profile, nil
}

// ---- Message deletion ----

// GetMessageMeta returns a message's conversation and sender for authorization.
func (r *ChatRepository) GetMessageMeta(ctx *fiber.Ctx, messageID uint) (conversationID, senderID uint, err error) {
	row := struct {
		ConversationID uint `db:"conversation_id"`
		SenderID       uint `db:"sender_id"`
	}{}
	err = r.sqlDB.GetContext(ctx.Context(), &row,
		`SELECT conversation_id, sender_id FROM messages WHERE id = $1 AND deleted_at IS NULL`, messageID)
	if err != nil {
		return 0, 0, err
	}
	return row.ConversationID, row.SenderID, nil
}

// DeleteMessageForEveryone marks a message deleted for all, blanks its content
// and removes its attachment rows. It returns the object keys of any attachments
// so the caller can purge them from object storage.
func (r *ChatRepository) DeleteMessageForEveryone(ctx *fiber.Ctx, messageID, byUserID uint) ([]string, error) {
	objectKeys := []string{}
	if err := r.sqlDB.SelectContext(ctx.Context(), &objectKeys,
		`SELECT object_key FROM message_attachments WHERE message_id = $1`, messageID); err != nil {
		return nil, err
	}
	if _, err := r.sqlDB.ExecContext(ctx.Context(),
		`DELETE FROM message_attachments WHERE message_id = $1`, messageID); err != nil {
		return nil, err
	}
	if _, err := r.sqlDB.ExecContext(ctx.Context(),
		`UPDATE messages SET deleted_for_all = true, deleted_by_id = $2, content = '', updated_at = now()
		 WHERE id = $1`, messageID, byUserID); err != nil {
		return nil, err
	}
	return objectKeys, nil
}

// HideMessageForUser hides a message for a single user ("delete for me").
func (r *ChatRepository) HideMessageForUser(ctx *fiber.Ctx, messageID, userID uint) error {
	_, err := r.sqlDB.ExecContext(ctx.Context(),
		`INSERT INTO message_hides (message_id, user_id) VALUES ($1, $2)
		 ON CONFLICT (message_id, user_id) DO NOTHING`, messageID, userID)
	return err
}

// ---- Context-based helpers (websocket handler, no *fiber.Ctx) ----

// IsParticipantCtx is IsParticipant for callers holding a plain context.
func (r *ChatRepository) IsParticipantCtx(ctx context.Context, conversationID, userID uint) (bool, error) {
	var count int
	if err := r.sqlDB.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM conversation_participants
		 WHERE conversation_id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		conversationID, userID); err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetParticipantUserIDsCtx is GetParticipantUserIDs for a plain context.
func (r *ChatRepository) GetParticipantUserIDsCtx(ctx context.Context, conversationID uint) ([]uint, error) {
	ids := []uint{}
	if err := r.sqlDB.SelectContext(ctx, &ids,
		`SELECT user_id FROM conversation_participants
		 WHERE conversation_id = $1 AND deleted_at IS NULL`, conversationID); err != nil {
		return nil, err
	}
	return ids, nil
}

// ---- AI assistant ----

// GetAIUserID returns the id of the built-in AI assistant user.
func (r *ChatRepository) GetAIUserID(ctx *fiber.Ctx) (uint, bool, error) {
	var id uint
	err := r.sqlDB.GetContext(ctx.Context(), &id,
		`SELECT id FROM users WHERE is_ai = true AND deleted_at IS NULL ORDER BY id LIMIT 1`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return id, true, nil
}

// GetDirectAIParticipant returns the AI user's id when the given direct
// conversation's other participant is an AI assistant.
func (r *ChatRepository) GetDirectAIParticipant(ctx context.Context, conversationID, meID uint) (uint, bool, error) {
	var id uint
	err := r.sqlDB.GetContext(ctx, &id,
		`SELECT cp.user_id
		 FROM conversation_participants cp
		 JOIN conversations c ON c.id = cp.conversation_id
		 JOIN users u ON u.id = cp.user_id
		 WHERE cp.conversation_id = $1 AND cp.user_id <> $2 AND cp.deleted_at IS NULL
		   AND c.type = 'direct' AND u.is_ai = true
		 LIMIT 1`, conversationID, meID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return id, true, nil
}

// GetRecentAIContext returns the most recent non-deleted, non-system messages
// (oldest first) for building the assistant prompt.
// A message with no text but an image attached still has to reach the model —
// "what does this say?" is often sent as a bare screenshot — so rows are kept
// when they carry either content or an image.
func (r *ChatRepository) GetRecentAIContext(ctx context.Context, conversationID uint, limit int) ([]dtos.AIMessage, error) {
	rows := []dtos.AIMessage{}
	if err := r.sqlDB.SelectContext(ctx, &rows,
		`SELECT id, sender_id, content FROM (
			SELECT m.id, m.sender_id, m.content FROM messages m
			WHERE m.conversation_id = $1 AND m.deleted_at IS NULL
			  AND m.deleted_for_all = false AND m.type <> 'system'
			  AND (m.content <> '' OR EXISTS (
				SELECT 1 FROM message_attachments a
				WHERE a.message_id = m.id AND a.kind = 'image'
			  ))
			ORDER BY m.id DESC LIMIT $2
		) t ORDER BY id ASC`, conversationID, limit); err != nil {
		return nil, err
	}
	return rows, nil
}

// GetImageAttachments returns the image attachments for the given messages,
// oldest first, so the assistant can look at what was actually sent.
func (r *ChatRepository) GetImageAttachments(ctx context.Context, messageIDs []uint) ([]dtos.AIAttachment, error) {
	if len(messageIDs) == 0 {
		return nil, nil
	}
	query, args, err := sqlx.In(
		`SELECT message_id, object_key, mime_type, size_bytes
		 FROM message_attachments
		 WHERE message_id IN (?) AND kind = 'image'
		 ORDER BY message_id ASC, id ASC`, messageIDs)
	if err != nil {
		return nil, err
	}
	rows := []dtos.AIAttachment{}
	if err := r.sqlDB.SelectContext(ctx, &rows, r.sqlDB.Rebind(query), args...); err != nil {
		return nil, err
	}
	return rows, nil
}

// CreateAIMessage persists an assistant reply and returns the enriched message
// for broadcasting. Runs on a plain context (called from a background goroutine).
func (r *ChatRepository) CreateAIMessage(ctx context.Context, conversationID, aiUserID uint, content string) (*dtos.MessageDTO, error) {
	msg := models.Message{ConversationID: conversationID, SenderID: aiUserID, Content: content, Type: "text"}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&msg).Error; err != nil {
			return err
		}
		return tx.Model(&models.Conversation{}).
			Where("id = ?", conversationID).
			Updates(map[string]interface{}{
				"last_message_id": msg.ID,
				"last_message_at": gorm.Expr("now()"),
				"updated_at":      gorm.Expr("now()"),
			}).Error
	})
	if err != nil {
		return nil, err
	}

	result := dtos.MessageDTO{}
	if err := r.sqlDB.GetContext(ctx, &result,
		`SELECT m.id, m.conversation_id, m.sender_id, u.name AS sender_name, u.is_ai AS sender_is_ai,
			m.content, m.type, m.deleted_for_all, m.edited_at::text AS edited_at, m.reply_to_id,
			m.created_at::text AS created_at
		 FROM messages m JOIN users u ON u.id = m.sender_id WHERE m.id = $1`, msg.ID); err != nil {
		return nil, err
	}
	result.Mentions = []dtos.MentionDTO{}
	result.Reactions = []dtos.ReactionGroupDTO{}
	result.Attachments = []dtos.AttachmentDTO{}
	return &result, nil
}

// MarkRead advances the participant's read pointer to the latest message and
// stamps the read time. It returns that message id (0 if none) and the read
// timestamp.
func (r *ChatRepository) MarkRead(ctx *fiber.Ctx, conversationID, userID uint) (uint, string, error) {
	query := `UPDATE conversation_participants
		SET last_read_message_id = (
			SELECT MAX(id) FROM messages WHERE conversation_id = $1 AND deleted_at IS NULL
		), last_read_at = now(), updated_at = now()
		WHERE conversation_id = $1 AND user_id = $2 AND deleted_at IS NULL
		RETURNING COALESCE(last_read_message_id, 0) AS id, last_read_at::text AS at`
	row := struct {
		ID uint   `db:"id"`
		At string `db:"at"`
	}{}
	if err := r.sqlDB.GetContext(ctx.Context(), &row, query, conversationID, userID); err != nil {
		return 0, "", err
	}
	return row.ID, row.At, nil
}

// UpdateConversation updates a conversation's title and/or description.
func (r *ChatRepository) UpdateConversation(ctx *fiber.Ctx, conversationID uint, title, description *string) error {
	sets := map[string]interface{}{"updated_at": gorm.Expr("now()")}
	if title != nil {
		sets["title"] = *title
	}
	if description != nil {
		sets["description"] = *description
	}
	return r.db.WithContext(ctx.Context()).Model(&models.Conversation{}).
		Where("id = ?", conversationID).Updates(sets).Error
}
