package queue

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/DaniloNogueira-Silva/go-microservices/worker-service/internal/domain"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSConsumer struct {
	client    *sqs.Client
	queueURL  string
	processor domain.EventProcessor
	logger    *slog.Logger
}

func NewSQSConsumer(client *sqs.Client, queueURL string, processor domain.EventProcessor, logger *slog.Logger) *SQSConsumer {
	return &SQSConsumer{
		client:    client,
		queueURL:  queueURL,
		processor: processor,
		logger:    logger,
	}
}

func (c *SQSConsumer) Start(ctx context.Context) {
	c.logger.Info("Worker iniciado. Aguardando mensagens...", "fila", c.queueURL)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Encerrando worker gracefulmente...")
			return
		default:
			// 1. Puxar mensagens
			result, err := c.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
				QueueUrl:            aws.String(c.queueURL),
				MaxNumberOfMessages: 10,
				WaitTimeSeconds:     5, // Long polling para economizar CPU
			})

			if err != nil {
				c.logger.Error("Erro ao receber mensagem", "erro", err)
				time.Sleep(2 * time.Second)
				continue
			}

			// 2. Processar mensagens recebidas
			for _, msg := range result.Messages {
				var event domain.UserEvent
				if err := json.Unmarshal([]byte(*msg.Body), &event); err != nil {
					c.logger.Error("Falha ao fazer parse do JSON", "erro", err, "body", *msg.Body)
					continue
				}

				if event.Action == "USER_REGISTERED" {
					// Chama a regra de negócio
					err := c.processor.ProcessUserRegistration(ctx, event.Email)
					if err == nil {
						// 3. Deletar da fila apenas se o processamento for bem-sucedido
						c.deleteMessage(ctx, msg.ReceiptHandle)
					} else {
						c.logger.Error("Falha ao processar evento de usuário", "erro", err)
					}
				}
			}
		}
	}
}

func (c *SQSConsumer) deleteMessage(ctx context.Context, receiptHandle *string) {
	_, err := c.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.queueURL),
		ReceiptHandle: receiptHandle,
	})
	if err != nil {
		c.logger.Error("Erro ao deletar mensagem da fila", "erro", err)
	} else {
		c.logger.Info("Mensagem processada e deletada com sucesso")
	}
}
