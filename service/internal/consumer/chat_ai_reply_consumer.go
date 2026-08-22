package consumer

import (
	"encoding/json"
	"log"

	"github.com/nibroos/s-erp-api/service/internal/ai"
	"github.com/nibroos/s-erp-api/service/internal/chat"
	"github.com/nibroos/s-erp-api/service/internal/config"
	"github.com/nibroos/s-erp-api/service/internal/repository"
	"github.com/nibroos/s-erp-api/service/internal/service"
)

// setupChatAIReplyConsumer runs the AI-reply worker: it pulls "generate a reply"
// jobs off the queue, generates the assistant's answer out-of-process (so AI
// load never competes with the API's request handling), and delivers it to the
// user's socket via the Redis fan-out. Enabled only when CHAT_AI_QUEUE=true.
func (r *ConsumerRouter) setupChatAIReplyConsumer() error {
	// Build the AI generator once. It uses its own hub in publish-only mode: the
	// worker holds no sockets, so replies reach the API instances through Redis.
	minioStorage, err := config.NewMinioStorage()
	if err != nil {
		log.Printf("chat AI worker: MinIO unavailable, image reading disabled: %v", err)
		minioStorage = nil
	}
	assistant := ai.NewAssistant(ai.New(), ai.NewDataSource(r.sqlDB))
	log.Printf("chat AI worker: assistant provider %s", assistant.Name())

	hub := chat.NewHub()
	if config.RedisClient != nil {
		hub.EnableRedis(config.RedisClient, chat.FanoutChannel, false) // publish-only
		log.Printf("chat AI worker: delivering replies via Redis fan-out (%q)", chat.FanoutChannel)
	} else {
		log.Printf("chat AI worker: WARNING — Redis unavailable; generated replies cannot reach sockets")
	}

	chatRepo := repository.NewChatRepository(r.gormDB, r.sqlDB, minioStorage)
	svc := service.NewChatService(chatRepo, hub, minioStorage, assistant)

	ch := r.rabbitmq.Channel
	if err := service.DeclareAIReplyQueue(ch); err != nil {
		return err
	}
	// One unacked job at a time per worker: total AI concurrency then equals the
	// number of consumer replicas, matching the model server's real parallelism.
	if err := ch.Qos(1, 0, false); err != nil {
		return err
	}

	msgs, err := ch.Consume(
		service.ChatAIReplyQueueName, // queue
		"",                           // consumer tag
		false,                        // manual ack
		false, false, false, nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			var job service.AIReplyJob
			if err := json.Unmarshal(d.Body, &job); err != nil {
				log.Printf("chat AI worker: unparseable job, dead-lettering: %v", err)
				_ = d.Reject(false) // → DLQ (poison message)
				continue
			}

			// RunAIReply is idempotent and posts an apology on model failure
			// (returning nil), so a non-nil error here is a retryable infra
			// failure: retry once, then dead-letter.
			if err := svc.RunAIReply(job.ConversationID, job.AIUserID, job.TriggerMessageID); err != nil {
				if d.Redelivered {
					log.Printf("chat AI worker: job failed again, dead-lettering (conv %d): %v", job.ConversationID, err)
					_ = d.Reject(false)
				} else {
					log.Printf("chat AI worker: job failed, retrying once (conv %d): %v", job.ConversationID, err)
					_ = d.Reject(true)
				}
				continue
			}
			_ = d.Ack(false)
		}
	}()

	log.Printf("chat AI worker: consuming %q", service.ChatAIReplyQueueName)
	return nil
}
