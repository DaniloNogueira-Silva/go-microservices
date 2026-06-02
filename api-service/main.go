package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Configurações
var jwtKey = []byte("super_secret_key")

const queueURL = "http://localhost:4566/000000000000/user-events-queue"

// Modelo do Banco de Dados
type User struct {
	ID       uint   `gorm:"primaryKey"`
	Email    string `gorm:"unique"`
	Password string
}

func main() {
	// 1. Conectar ao PostgreSQL
	dsn := "host=localhost user=user password=password dbname=microservices_db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Falha ao conectar ao banco:", err)
	}
	db.AutoMigrate(&User{}) // Cria a tabela se não existir

	// 2. Conectar ao LocalStack SQS
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

	// 3. Configurar Rotas Gin
	r := gin.Default()

	// Rota de Registro
	r.POST("/register", func(c *gin.Context) {
		var input struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.BindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Input inválido"})
			return
		}

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), 14)
		user := User{Email: input.Email, Password: string(hashedPassword)}

		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar usuário"})
			return
		}

		// Enviar mensagem para o SQS (Evento de novo usuário)
		msgBody := fmt.Sprintf(`{"action": "USER_REGISTERED", "email": "%s"}`, user.Email)
		_, err = sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
			QueueUrl:    aws.String(queueURL),
			MessageBody: aws.String(msgBody),
		})
		if err != nil {
			log.Println("Erro ao enviar para o SQS:", err)
		}

		c.JSON(http.StatusOK, gin.H{"message": "Usuário registrado com sucesso"})
	})

	// Rota de Login (Gera JWT)
	r.POST("/login", func(c *gin.Context) {
		var input User
		c.BindJSON(&input)

		var user User
		db.Where("email = ?", input.Email).First(&user)
		if user.ID == 0 || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)) != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Credenciais inválidas"})
			return
		}

		// Gera Token JWT
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": user.ID,
			"exp":     time.Now().Add(time.Hour * 24).Unix(),
		})
		tokenString, _ := token.SignedString(jwtKey)

		c.JSON(http.StatusOK, gin.H{"token": tokenString})
	})

	// Rota Protegida com Middleware JWT
	r.GET("/protected", AuthMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("user_id")
		c.JSON(http.StatusOK, gin.H{"message": "Bem-vindo à rota protegida!", "user_id": userID})
	})

	log.Println("API rodando na porta 8080")
	r.Run(":8080")
}

// Middleware de Autenticação JWT
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token não fornecido"})
			c.Abort()
			return
		}

		// Remove o prefixo "Bearer " se existir
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
			c.Abort()
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		c.Set("user_id", claims["user_id"])
		c.Next()
	}
}
