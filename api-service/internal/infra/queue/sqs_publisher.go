package queue

import (
	"context"
	"fmt"

	"github.com/DaniloNogueira-Silva/go-microservices/api-service/internal/domain"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// SQSPublisher implementa a interface domain.EventPublisher
type SQSPublisher struct {
	client   *sqs.Client
	queueURL string
}

func NewSQSPublisher(client *sqs.Client, queueURL string) domain.EventPublisher {
	return &SQSPublisher{
		client:   client,
		queueURL: queueURL,
	}
}

func (p *SQSPublisher) PublishUserRegistered(ctx context.Context, email string) error {
	msgBody := fmt.Sprintf(`{"action": "USER_REGISTERED", "email": "%s"}`, email)

	_, err := p.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(p.queueURL),
		MessageBody: aws.String(msgBody),
	})

	if err != nil {
		return fmt.Errorf("falha ao enviar mensagem para o SQS: %w", err)
	}

	return nil
}
