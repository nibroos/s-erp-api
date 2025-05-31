package consumer

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

type ConsumerHealth struct {
	DB       *gorm.DB
	RabbitMQ *amqp.Connection
}

func (h *ConsumerHealth) Handler(c *fiber.Ctx) error {
	// Check DB
	sqlDB, err := h.DB.DB()
	if err != nil || sqlDB.Ping() != nil {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"status": "db unhealthy"})
	}
	// Check RabbitMQ
	if h.RabbitMQ == nil || h.RabbitMQ.IsClosed() {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"status": "rabbitmq unhealthy"})
	}
	return c.JSON(fiber.Map{"status": "ok"})
}
