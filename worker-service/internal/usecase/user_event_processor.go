package usecase

import (
	"context"
	"log/slog"
	"time"
)

type UserProcessor struct {
	logger *slog.Logger
}

func NewUserProcessor(logger *slog.Logger) *UserProcessor {
	return &UserProcessor{logger: logger}
}

// ProcessUserRegistration implementa domain.EventProcessor
func (p *UserProcessor) ProcessUserRegistration(ctx context.Context, email string) error {
	p.logger.Info("Iniciando processamento para novo usuário", "email", email)

	// Simulando uma tarefa demorada (ex: envio de e-mail via SendGrid/SES)
	time.Sleep(2 * time.Second)

	p.logger.Info("E-mail de boas-vindas enviado com sucesso!", "email", email)

	// Retornar nil significa que a mensagem foi processada com sucesso
	// e a infraestrutura (SQS) pode deletar a mensagem da fila.
	return nil
}
