package handlers

import (
	"errors"
	"net/http"

	"github.com/DaniloNogueira-Silva/go-microservices/api-service/internal/usecase"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *usecase.AuthService
}

func NewAuthHandler(service *usecase.AuthService) *AuthHandler {
	return &AuthHandler{authService: service}
}

// Register lida com a rota POST /register
func (h *AuthHandler) Register(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	// Validação básica de entrada
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos: verifique email e senha (mín. 6 caracteres)"})
		return
	}

	// Chama a regra de negócio
	err := h.authService.Register(c.Request.Context(), input.Email, input.Password)
	if err != nil {
		// Tratamento de erro amigável baseado nos erros conhecidos do UseCase
		if errors.Is(err, usecase.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "Este email já está em uso"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno ao processar registro"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Usuário registrado com sucesso"})
}
