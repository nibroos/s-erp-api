package consumer

import (
	"encoding/json"
	"log"

	"github.com/nibroos/s-erp-api/service/internal/dtos"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/service"
	amqp "github.com/rabbitmq/amqp091-go"
)

func (r *ConsumerRouter) setupBulkSendEmailApprovedInvoiceMaintenanceConsumer() error {
	ch := r.rabbitmq.Channel

	// Setup dead letter exchange and queue
	dlx, err := ch.QueueDeclare(
		"bulk_send_email_approved_invoice_maintenance_dlq", // dead letter queue name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	// Declare the main queue with dead letter configuration
	q, err := ch.QueueDeclare(
		"bulk_send_email_approved_invoice_maintenance_queue", // queue name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		amqp.Table{
			"x-dead-letter-exchange":    "",                               // default exchange
			"x-dead-letter-routing-key": dlx.Name,                         // route to DLQ
			"x-message-ttl":             int32(retryDelay.Milliseconds()), // TTL for retry delay
		},
	)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			// Get retry count from message headers
			var retryCount int
			if xDeath, ok := d.Headers["x-death"].([]interface{}); ok && len(xDeath) > 0 {
				if count, ok := xDeath[0].(amqp.Table)["count"].(int64); ok {
					retryCount = int(count)
				}
			}

			var syncStockData dtos.BulkSendEmailApprovedInvoiceMaintenancesRequest
			if err := json.Unmarshal(d.Body, &syncStockData); err != nil {
				log.Printf("Error parsing send email data: %v", err)
				r.handleMessageFailure(d, retryCount, q.Name, dlx.Name, err)
				continue
			}

			// Process send email
			if err := r.bulkSendEmailApprovedInvoiceMaintenanceService(syncStockData); err != nil {
				log.Printf("Error processing send email: %v", err)
				r.handleMessageFailure(d, retryCount, q.Name, dlx.Name, err)
				continue
			}

			d.Ack(false)
		}
	}()

	return nil
}

func (r *ConsumerRouter) bulkSendEmailApprovedInvoiceMaintenanceService(data dtos.BulkSendEmailApprovedInvoiceMaintenancesRequest) error {
	utilRepo := repository.NewUtilRepository(r.gormDB, r.sqlDB)
	invoiceMaintenanceRepo := repository.NewInvoiceMaintenanceRepository(r.gormDB, r.sqlDB, utilRepo, r.rabbitmq, r.tracer)
	invoiceMaintenanceService := service.NewInvoiceMaintenanceService(invoiceMaintenanceRepo, utilRepo, r.rabbitmq, r.tracer)

	log.Printf("Processing bulkSendEmailApprovedInvoiceMaintenanceService data: %+v", data)
	if err := invoiceMaintenanceService.ConsumeBulkSendEmailApprovedInvoiceMaintenance(data); err != nil {
		return err
	}

	return nil
}
