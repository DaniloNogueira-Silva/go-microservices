package main

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

const queueURL = "http://localhost:4566/000000000000/user-events-queue"

func main() {
	// 1. Conectar ao LocalStack SQS
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: "http://localhost:4566"}, nil
			})),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{AccessKeyID: "test", SecretAccessKey: "test"}, nil
		})),
	)
	if err != nil {
		log.Fatal("Erro ao carregar config da AWS:", err)
	}
	sqsClient := sqs.NewFromConfig(cfg)

	log.Println("Worker iniciado. Aguardando mensagens na fila...")

	// 2. Loop infinito (Polling)
	for {
		// Puxar mensagens
		result, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(queueURL),
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     5, // Long polling
		})

		if err != nil {
			log.Println("Erro ao receber mensagem:", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if len(result.Messages) == 0 {
			continue // Nenhuma mensagem na fila
		}

		// 3. Processar mensagens
		for _, msg := range result.Messages {
			log.Printf("Processando evento: %s\n", *msg.Body)

			// Aqui você enviaria um email, geraria um relatório, etc.
			// ...

			// 4. Deletar a mensagem da fila após sucesso
			_, err := sqsClient.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(queueURL),
				ReceiptHandle: msg.ReceiptHandle,
			})
			if err != nil {
				log.Println("Erro ao deletar mensagem:", err)
			} else {
				log.Println("Mensagem processada e deletada com sucesso.")
			}
		}
	}
}
