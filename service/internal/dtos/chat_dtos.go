package dtos

// ---- Requests ----

type StartConversationRequest struct {
	// UserID is the account to start / open a direct conversation with.
	UserID *uint `json:"user_id"`
}

type ListConversationsRequest struct {
	Global         *string `json:"global"`
	PerPage        *string `json:"per_page" default:"20"`
	Page           *string `json:"page" default:"1"`
	OrderColumn    *string `json:"order_column" default:"last_message_at"`
	OrderDirection *string `json:"order_direction" default:"desc"`
}

type ListMessagesRequest struct {
	ConversationID *uint `json:"conversation_id"`
	BeforeID       *uint `json:"before_id"` // keyset pagination: fetch messages older than this id
	PerPage        *int  `json:"per_page"`
}

type SendMessageRequest struct {
	ConversationID *uint             `json:"conversation_id"`
	Content        *string           `json:"content"`
	ReplyToID      *uint             `json:"reply_to_id"`
	MentionIDs     []uint            `json:"mention_ids"`
	Attachments    []AttachmentInput `json:"attachments"`
}

// AttachmentInput is a stored-object reference the client attaches to a message
// (after uploading via /chats/upload).
type AttachmentInput struct {
	ObjectKey       string  `json:"object_key"`
	Kind            string  `json:"kind"`
	FileName        *string `json:"file_name"`
	MimeType        *string `json:"mime_type"`
	SizeBytes       *int64  `json:"size_bytes"`
	DurationSeconds *int    `json:"duration_seconds"`
	Width           *int    `json:"width"`
	Height          *int    `json:"height"`
	Description     *string `json:"description"`
}

type EditMessageRequest struct {
	MessageID *uint   `json:"message_id"`
	Content   *string `json:"content"`
}

type ReactMessageRequest struct {
	MessageID *uint   `json:"message_id"`
	Emoji     *string `json:"emoji"`
}

type CreateThreadRequest struct {
	ConversationID *uint   `json:"conversation_id"`
	RootMessageID  *uint   `json:"root_message_id"`
	Title          *string `json:"title"`
}

type ListThreadsRequest struct {
	ConversationID *uint `json:"conversation_id"`
}

type ShowConversationRequest struct {
	ConversationID *uint `json:"conversation_id"`
}

type UpdateConversationRequest struct {
	ConversationID *uint   `json:"conversation_id"`
	Title          *string `json:"title"`
	Description    *string `json:"description"`
}

type ReadConversationRequest struct {
	ConversationID *uint `json:"conversation_id"`
}

type CreateGroupRequest struct {
	Title     *string `json:"title"`
	MemberIDs []uint  `json:"member_ids"`
}

type AddMembersRequest struct {
	ConversationID *uint  `json:"conversation_id"`
	MemberIDs      []uint `json:"member_ids"`
}

type RemoveMemberRequest struct {
	ConversationID *uint `json:"conversation_id"`
	UserID         *uint `json:"user_id"`
}

type SetMemberRoleRequest struct {
	ConversationID *uint   `json:"conversation_id"`
	UserID         *uint   `json:"user_id"`
	Role           *string `json:"role"` // "admin" | "member" | "owner" (owner = transfer)
}

type LeaveConversationRequest struct {
	ConversationID *uint `json:"conversation_id"`
}

type ListMembersRequest struct {
	ConversationID *uint `json:"conversation_id"`
}

type MemberProfileRequest struct {
	ConversationID *uint `json:"conversation_id"`
	UserID         *uint `json:"user_id"`
}

type DeleteMessageRequest struct {
	MessageID *uint   `json:"message_id"`
	Scope     *string `json:"scope"` // "everyone" | "me"
}

// CreateMessageInput carries everything needed to persist a message.
type CreateMessageInput struct {
	ConversationID uint
	SenderID       uint
	Content        string
	Type           string // "text" | "system"
	ReplyToID      *uint
	MentionIDs     []uint
	Attachments    []AttachmentInput
}

// UploadResponse is returned by /chats/upload; the client echoes these fields
// back inside a SendMessageRequest.Attachments entry.
type UploadResponse struct {
	ObjectKey string `json:"object_key"`
	Kind      string `json:"kind"`
	FileName  string `json:"file_name"`
	MimeType  string `json:"mime_type"`
	SizeBytes int64  `json:"size_bytes"`
	URL       string `json:"url"`
}

// AttachmentDTO is an attachment returned with a message (URL is presigned).
type AttachmentDTO struct {
	ID              uint    `json:"id" db:"id"`
	Kind            string  `json:"kind" db:"kind"`
	ObjectKey       string  `json:"-" db:"object_key"`
	FileName        *string `json:"file_name" db:"file_name"`
	MimeType        *string `json:"mime_type" db:"mime_type"`
	SizeBytes       *int64  `json:"size_bytes" db:"size_bytes"`
	DurationSeconds *int    `json:"duration_seconds" db:"duration_seconds"`
	Width           *int    `json:"width" db:"width"`
	Height          *int    `json:"height" db:"height"`
	Description     *string `json:"description" db:"description"`
	MessageID       uint    `json:"-" db:"message_id"`
	URL             string  `json:"url" db:"-"`
}

// AIMessage is one message row used to build assistant context.
type AIMessage struct {
	ID       uint   `db:"id"`
	SenderID uint   `db:"sender_id"`
	Content  string `db:"content"`
}

// AIAttachment is an image attached to a message in the assistant's context.
type AIAttachment struct {
	MessageID uint    `db:"message_id"`
	ObjectKey string  `db:"object_key"`
	MimeType  *string `db:"mime_type"`
	SizeBytes *int64  `db:"size_bytes"`
}

// ---- Responses ----

// ConversationListDTO is one row in the current user's conversation list.
type ConversationListDTO struct {
	ID              uint    `json:"id" db:"id"`
	Type            string  `json:"type" db:"type"`
	Title           *string `json:"title" db:"title"`
	Description     *string `json:"description" db:"description"`
	LastMessageAt   *string `json:"last_message_at" db:"last_message_at"`
	LastMessageText *string `json:"last_message_text" db:"last_message_text"`
	UnreadCount     int     `json:"unread_count" db:"unread_count"`
	// Counterpart info for direct conversations (null for groups).
	OtherUserID         *uint   `json:"other_user_id" db:"other_user_id"`
	OtherUserName       *string `json:"other_user_name" db:"other_user_name"`
	OtherUserEmail      *string `json:"other_user_email" db:"other_user_email"`
	OtherUserImage      *string `json:"other_user_image" db:"other_user_image"`
	OtherUserIsAI       bool    `json:"other_user_is_ai" db:"other_user_is_ai"`
	OtherUserLastRead   *uint   `json:"other_user_last_read" db:"other_user_last_read"`
	OtherUserLastReadAt *string `json:"other_user_last_read_at" db:"other_user_last_read_at"`
	// Group metadata.
	MyRole      string `json:"my_role" db:"my_role"`
	MemberCount int    `json:"member_count" db:"member_count"`
	// Thread metadata (parent_id is set for threads).
	ParentID *uint `json:"parent_id" db:"parent_id"`
	// Online is derived from the live websocket hub (not persisted).
	Online bool `json:"online" db:"-"`
}

// ReplyPreviewDTO is the small quoted snippet of the message being replied to.
type ReplyPreviewDTO struct {
	ID            uint    `json:"id" db:"id"`
	SenderName    *string `json:"sender_name" db:"sender_name"`
	Content       string  `json:"content" db:"content"`
	DeletedForAll bool    `json:"deleted_for_all" db:"deleted_for_all"`
}

// MentionDTO is a user tagged in a message.
type MentionDTO struct {
	UserID uint   `json:"user_id" db:"user_id"`
	Name   string `json:"name" db:"name"`
}

// ReactionGroupDTO aggregates one emoji's reactions on a message.
type ReactionGroupDTO struct {
	Emoji   string `json:"emoji"`
	Count   int    `json:"count"`
	UserIDs []uint `json:"user_ids"`
	Reacted bool   `json:"reacted"` // did the requesting user react with this emoji
}

// MessageDTO is a single chat message enriched with sender info + relations.
type MessageDTO struct {
	ID             uint    `json:"id" db:"id"`
	ConversationID uint    `json:"conversation_id" db:"conversation_id"`
	SenderID       uint    `json:"sender_id" db:"sender_id"`
	SenderName     *string `json:"sender_name" db:"sender_name"`
	SenderIsAI     bool    `json:"sender_is_ai" db:"sender_is_ai"`
	Content        string  `json:"content" db:"content"`
	Type           string  `json:"type" db:"type"`
	DeletedForAll  bool    `json:"deleted_for_all" db:"deleted_for_all"`
	EditedAt       *string `json:"edited_at" db:"edited_at"`
	ReplyToID      *uint   `json:"reply_to_id" db:"reply_to_id"`
	CreatedAt      *string `json:"created_at" db:"created_at"`

	// Enriched relations (populated in a second pass, not scanned from SQL).
	ReplyTo     *ReplyPreviewDTO   `json:"reply_to" db:"-"`
	Mentions    []MentionDTO       `json:"mentions" db:"-"`
	Reactions   []ReactionGroupDTO `json:"reactions" db:"-"`
	Attachments []AttachmentDTO    `json:"attachments" db:"-"`
	// Thread branched from this message, if any.
	ThreadConversationID *uint   `json:"thread_conversation_id" db:"-"`
	ThreadReplyCount     int     `json:"thread_reply_count" db:"-"`
	ThreadTitle          *string `json:"thread_title" db:"-"`
}

// ThreadListDTO is one thread in the "Threads" panel.
type ThreadListDTO struct {
	ConversationID uint    `json:"conversation_id" db:"conversation_id"`
	Title          *string `json:"title" db:"title"`
	RootMessageID  *uint   `json:"root_message_id" db:"root_message_id"`
	RootSnippet    *string `json:"root_snippet" db:"root_snippet"`
	ReplyCount     int     `json:"reply_count" db:"reply_count"`
	LastMessageAt  *string `json:"last_message_at" db:"last_message_at"`
}

// MemberDTO describes one member of a conversation.
type MemberDTO struct {
	UserID            uint    `json:"user_id" db:"user_id"`
	Name              string  `json:"name" db:"name"`
	Email             string  `json:"email" db:"email"`
	PhoneNumber       *string `json:"phone_number" db:"phone_number"`
	ProfileImage      *string `json:"profile_image_url" db:"profile_image_url"`
	Role              string  `json:"role" db:"role"`
	IsAI              bool    `json:"is_ai" db:"is_ai"`
	LastReadMessageID *uint   `json:"last_read_message_id" db:"last_read_message_id"`
	LastReadAt        *string `json:"last_read_at" db:"last_read_at"`
	Online            bool    `json:"online" db:"-"`
}

// UserProfileDTO is the contact/member profile card.
type UserProfileDTO struct {
	ID           uint     `json:"id" db:"id"`
	Name         string   `json:"name" db:"name"`
	Email        string   `json:"email" db:"email"`
	Status       *int     `json:"status" db:"status"`
	PhoneNumber  *string  `json:"phone_number" db:"phone_number"`
	Address      *string  `json:"address" db:"address"`
	ProfileImage *string  `json:"profile_image_url" db:"profile_image_url"`
	IsAI         bool     `json:"is_ai" db:"is_ai"`
	Roles        []string `json:"roles" db:"-"`
	// ConversationRole is this user's role in the conversation (if any).
	ConversationRole *string `json:"conversation_role" db:"-"`
	Online           bool    `json:"online" db:"-"`
}

// WSEnvelope is the message envelope exchanged over the websocket connection.
type WSEnvelope struct {
	Type    string      `json:"type"` // "message" | "read" | "typing" | "error" | "ping"
	Payload interface{} `json:"payload,omitempty"`
}

// WSInbound is what the client sends over the socket.
type WSInbound struct {
	Type           string `json:"type"`
	ConversationID *uint  `json:"conversation_id"`
	Content        string `json:"content"`
}
