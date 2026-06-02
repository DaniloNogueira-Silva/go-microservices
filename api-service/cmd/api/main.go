package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/DaniloNogueira-Silva/go-microservices/api-service/internal/domain"
	"github.com/DaniloNogueira-Silva/go-microservices/api-service/internal/infra/database"
	"github.com/DaniloNogueira-Silva/go-microservices/api-service/internal/infra/queue"
	"github.com/DaniloNogueira-Silva/go-microservices/api-service/internal/usecase"
	"github.com/DaniloNogueira-Silva/go-microservices/api-service/internal/web/handlers"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. Setup de Logs Estruturados
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 2. Setup Infra: Banco de Dados
	dsn := "host=localhost user=user password=password dbname=microservices_db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Error("Falha ao conectar ao banco", "erro", err)
		os.Exit(1)
	}
	db.AutoMigrate(&domain.User{}) // Migrate da Entidade Limpa

	// 3. Setup Infra: SQS (LocalStack)
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
		logger.Error("Erro ao carregar config da AWS", "erro", err)
		os.Exit(1)
	}
	sqsClient := sqs.NewFromConfig(cfg)
	queueURL := "http://localhost:4566/000000000000/user-events-queue"

	// --- O CORAÇÃO DA INJEÇÃO DE DEPENDÊNCIA --- //

	// A. Instancia as implementações de Infraestrutura
	userRepo := database.NewPostgresUserRepository(db)
	eventPublisher := queue.NewSQSPublisher(sqsClient, queueURL)

	// B. Injeta a Infraestrutura na Regra de Negócio (Use Case)
	authService := usecase.NewAuthService(userRepo, eventPublisher, logger)

	// C. Injeta a Regra de Negócio na Camada de Apresentação (Web)
	authHandler := handlers.NewAuthHandler(authService)

	// ------------------------------------------ //

	// 4. Setup e Execução do Servidor Web (Gin)
	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.POST("/register", authHandler.Register)
	}

	logger.Info("Iniciando API Service", "porta", 8080)
	if err := r.Run(":8080"); err != nil {
		logger.Error("Erro ao rodar servidor", "erro", err)
	}
}
