package models

import (
	"gorm.io/gorm"
)

// Conversation is a chat thread between two (direct) or more (group) users.
type Conversation struct {
	ID            uint           `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Type          string         `json:"type" gorm:"column:type"`
	Title         *string        `json:"title" gorm:"column:title"`
	Description   *string        `json:"description" gorm:"column:description"`
	DirectKey     *string        `json:"-" gorm:"column:direct_key"`
	ParentID      *uint          `json:"parent_id" gorm:"column:parent_id"`
	RootMessageID *uint          `json:"root_message_id" gorm:"column:root_message_id"`
	LastMessageID *uint          `json:"last_message_id" gorm:"column:last_message_id"`
	LastMessageAt *string        `json:"last_message_at" gorm:"column:last_message_at"`
	CreatedByID   *uint          `json:"created_by_id" gorm:"column:created_by_id"`
	UpdatedByID   *uint          `json:"updated_by_id" gorm:"column:updated_by_id"`
	DeletedByID   *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (Conversation) TableName() string { return "conversations" }

// Participant role constants.
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
)

// ConversationParticipant links a user to a conversation.
type ConversationParticipant struct {
	ID                uint           `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ConversationID    uint           `json:"conversation_id" gorm:"column:conversation_id"`
	UserID            uint           `json:"user_id" gorm:"column:user_id"`
	Role              string         `json:"role" gorm:"column:role"`
	LastReadMessageID *uint          `json:"last_read_message_id" gorm:"column:last_read_message_id"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (ConversationParticipant) TableName() string { return "conversation_participants" }

// Message is a single chat message.
// created_at / updated_at are intentionally omitted so GORM does not insert
// NULL over the database's CURRENT_TIMESTAMP defaults; they are read back via
// the DTO queries instead.
type Message struct {
	ID             uint           `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	ConversationID uint           `json:"conversation_id" gorm:"column:conversation_id"`
	SenderID       uint           `json:"sender_id" gorm:"column:sender_id"`
	Content        string         `json:"content" gorm:"column:content"`
	Type           string         `json:"type" gorm:"column:type"`
	ReplyToID      *uint          `json:"reply_to_id" gorm:"column:reply_to_id"`
	DeletedForAll  bool           `json:"deleted_for_all" gorm:"column:deleted_for_all"`
	DeletedByID    *uint          `json:"deleted_by_id" gorm:"column:deleted_by_id"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (Message) TableName() string { return "messages" }

// MessageHide records a per-user "delete for me".
type MessageHide struct {
	ID        uint `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	MessageID uint `json:"message_id" gorm:"column:message_id"`
	UserID    uint `json:"user_id" gorm:"column:user_id"`
}

func (MessageHide) TableName() string { return "message_hides" }

// MessageReaction is a single emoji reaction by a user on a message.
type MessageReaction struct {
	ID        uint   `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	MessageID uint   `json:"message_id" gorm:"column:message_id"`
	UserID    uint   `json:"user_id" gorm:"column:user_id"`
	Emoji     string `json:"emoji" gorm:"column:emoji"`
}

func (MessageReaction) TableName() string { return "message_reactions" }

// MessageMention links a message to a mentioned user.
type MessageMention struct {
	ID        uint `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	MessageID uint `json:"message_id" gorm:"column:message_id"`
	UserID    uint `json:"user_id" gorm:"column:user_id"`
}

func (MessageMention) TableName() string { return "message_mentions" }

// Attachment kind constants.
const (
	AttachmentImage    = "image"
	AttachmentVideo    = "video"
	AttachmentVoice    = "voice"
	AttachmentDocument = "document"
)

// MessageAttachment is a file stored in object storage, linked to a message.
type MessageAttachment struct {
	ID              uint    `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	MessageID       uint    `json:"message_id" gorm:"column:message_id"`
	Kind            string  `json:"kind" gorm:"column:kind"`
	ObjectKey       string  `json:"object_key" gorm:"column:object_key"`
	FileName        *string `json:"file_name" gorm:"column:file_name"`
	MimeType        *string `json:"mime_type" gorm:"column:mime_type"`
	SizeBytes       *int64  `json:"size_bytes" gorm:"column:size_bytes"`
	DurationSeconds *int    `json:"duration_seconds" gorm:"column:duration_seconds"`
	Width           *int    `json:"width" gorm:"column:width"`
	Height          *int    `json:"height" gorm:"column:height"`
	Description     *string `json:"description" gorm:"column:description"`
}

func (MessageAttachment) TableName() string { return "message_attachments" }
