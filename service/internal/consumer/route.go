package consumer

import (
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/nibroos/s-erp-api/service/internal/config"
	"github.com/opentracing/opentracing-go"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

const (
	maxRetries = 3
	retryDelay = 5 * time.Second
)

type ConsumerRouter struct {
	rabbitmq *config.RabbitMQ
	gormDB   *gorm.DB
	sqlDB    *sqlx.DB
	tracer   opentracing.Tracer
}

func NewConsumerRouter(rabbitmq *config.RabbitMQ, gormDB *gorm.DB, sqlDB *sqlx.DB, tracer opentracing.Tracer) *ConsumerRouter {
	return &ConsumerRouter{
		rabbitmq: rabbitmq,
		gormDB:   gormDB,
		sqlDB:    sqlDB,
		tracer:   tracer,
	}
}

func (r *ConsumerRouter) SetupConsumers() error {
	// Email consumer
	if err := r.setupSendEmailSolutionConsumer(); err != nil {
		return err
	}

	// Notification consumer
	if err := r.setupSyncStockConsumer(); err != nil {
		return err
	}

	// Notification consumer
	if err := r.setupBulkSendEmailApprovedInvoiceMaintenanceConsumer(); err != nil {
		return err
	}

	// AI-reply worker — only when the queued path is enabled.
	if os.Getenv("CHAT_AI_QUEUE") == "true" {
		if err := r.setupChatAIReplyConsumer(); err != nil {
			return err
		}
	}

	// Other consumers...

	return nil
}

func (r *ConsumerRouter) handleMessageFailure(d amqp.Delivery, retryCount int, queueName string, deadLetterQueueName string, err error) {
	if retryCount >= maxRetries {
		log.Printf("[MOVE] %s message failed after %d retries, moving to %s DLQ. Error: %v", queueName, maxRetries, deadLetterQueueName, err)
		// Reject and don't requeue - message will go to DLQ
		d.Reject(false)
		return
	}

	log.Printf("[RETRY] Retrying %s message, attempt %d of %d. Error: %v", queueName, retryCount+1, maxRetries, err)
	// Reject and requeue
	d.Reject(true)
}
