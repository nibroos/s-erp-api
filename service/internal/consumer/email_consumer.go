package consumer

import (
	"encoding/json"
	"log"
)

func (r *ConsumerRouter) setupEmailConsumer() error {
	ch := r.rabbitmq.Channel

	q, err := ch.QueueDeclare(
		"email_queue", // queue name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
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
			var emailData map[string]interface{}
			if err := json.Unmarshal(d.Body, &emailData); err != nil {
				log.Printf("Error parsing email data: %v", err)
				d.Nack(false, true)
				continue
			}

			// Process email
			if err := r.processEmail(emailData); err != nil {
				log.Printf("Error processing email: %v", err)
				d.Nack(false, true)
				continue
			}

			d.Ack(false)
		}
	}()

	return nil
}

func (r *ConsumerRouter) processEmail(data map[string]interface{}) error {
	// Implement email sending logic
	return nil
}
