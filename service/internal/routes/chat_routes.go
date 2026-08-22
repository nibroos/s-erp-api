package routes

import (
	"log"
	"os"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/ai"
	"github.com/nibroos/s-erp-api/service/internal/chat"
	"github.com/nibroos/s-erp-api/service/internal/config"
	"github.com/nibroos/s-erp-api/service/internal/controller/rest"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/service"
	"gorm.io/gorm"
)

// newChatController wires the chat dependency graph once so both the public
// websocket endpoint and the protected REST endpoints share the same hub.
func newChatController(gormDB *gorm.DB, sqlDB *sqlx.DB, rabbitmq *config.RabbitMQ) *rest.ChatController {
	// Object storage for attachments; if it can't be reached, chat still works
	// (uploads return 503) rather than blocking startup.
	minioStorage, err := config.NewMinioStorage()
	if err != nil {
		log.Printf("Warning: MinIO storage unavailable, attachments disabled: %v", err)
		minioStorage = nil
	}
	// The assistant answers from live ERP data through the curated read-only
	// `ai` schema (migration 20260713000009). Without that schema it still
	// chats, it just cannot look anything up.
	assistant := ai.NewAssistant(ai.New(), ai.NewDataSource(sqlDB))
	log.Printf("Chat AI assistant provider: %s", assistant.Name())
	chatRepo := repository.NewChatRepository(gormDB, sqlDB, minioStorage)
	chatService := service.NewChatService(chatRepo, chat.GlobalHub, minioStorage, assistant)

	// When the queued path is on, dispatch AI replies to the consumer-service via
	// RabbitMQ instead of generating them in-process. Delivery back to sockets
	// then rides the Redis fan-out enabled on the hub in main.
	if os.Getenv("CHAT_AI_QUEUE") == "true" {
		chatService.EnableAIQueue(rabbitmq)
	}
	return rest.NewChatController(chatService, chatRepo, chat.GlobalHub)
}

// SetupChatWSRoute registers the websocket endpoint. It must be mounted BEFORE
// the header-based JWT middleware because browsers cannot attach an
// Authorization header to a websocket handshake; the guard authenticates via a
// `token` query parameter instead.
func SetupChatWSRoute(version fiber.Router, controller *rest.ChatController) {
	version.Use("/chats/ws", controller.WSUpgradeGuard)
	version.Get("/chats/ws", websocket.New(controller.WSHandler))
}

// SetupChatRoutes registers the protected REST endpoints for chat.
func SetupChatRoutes(chats fiber.Router, controller *rest.ChatController) {
	chats.Post("/start-conversation", controller.StartConversation)
	chats.Post("/start-ai", controller.StartAIChat)
	chats.Post("/index-conversation", controller.GetConversations)
	chats.Post("/index-message", controller.GetMessages)
	chats.Post("/send-message", controller.SendMessage)
	chats.Post("/read-conversation", controller.ReadConversation)
	chats.Post("/upload", controller.Upload)

	// Groups & members
	chats.Post("/create-group", controller.CreateGroup)
	chats.Post("/index-member", controller.GetMembers)
	chats.Post("/add-member", controller.AddMembers)
	chats.Post("/remove-member", controller.RemoveMember)
	chats.Post("/set-member-role", controller.SetMemberRole)
	chats.Post("/leave-conversation", controller.LeaveConversation)

	// Profile & message deletion
	chats.Post("/show-profile", controller.GetProfile)
	chats.Post("/delete-message", controller.DeleteMessage)

	// Editing, reactions, threads
	chats.Post("/edit-message", controller.EditMessage)
	chats.Post("/react-message", controller.ReactMessage)
	chats.Post("/unreact-message", controller.UnreactMessage)
	chats.Post("/create-thread", controller.CreateThread)
	chats.Post("/index-thread", controller.GetThreads)
	chats.Post("/show-conversation", controller.GetConversation)
	chats.Post("/update-conversation", controller.UpdateConversation)
}
