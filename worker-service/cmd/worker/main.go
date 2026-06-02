package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/DaniloNogueira-Silva/go-microservices/worker-service/internal/infra/queue"
	"github.com/DaniloNogueira-Silva/go-microservices/worker-service/internal/usecase"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func main() {
	// 1. Setup de Logs
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Contexto com cancelamento para Graceful Shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2. Setup AWS SQS (LocalStack)
	cfg, err := config.LoadDefaultConfig(ctx,
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
		logger.Error("Erro ao carregar config da AWS", "erro", err)
		os.Exit(1)
	}
	sqsClient := sqs.NewFromConfig(cfg)
	queueURL := "http://localhost:4566/000000000000/user-events-queue"

	// 3. Injeção de Dependências
	processor := usecase.NewUserProcessor(logger)
	consumer := queue.NewSQSConsumer(sqsClient, queueURL, processor, logger)

	// 4. Lidar com encerramento seguro (Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Sinal de parada recebido. Desligando...")
		cancel() // Cancela o contexto, fazendo o loop do consumidor parar
	}()

	// 5. Iniciar o Consumidor (Trava a execução aqui até o context ser cancelado)
	consumer.Start(ctx)

	logger.Info("Worker encerrado com segurança.")
}
