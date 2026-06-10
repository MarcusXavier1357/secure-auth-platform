package service

import (
	"context"
	"log/slog"

	"auth-backend/internal/models"
	"auth-backend/internal/repository"
)

type AuditService struct {
	repo *repository.AuditRepository
}

func NewAuditService(repo *repository.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

// Log registra a ação. Falha de auditoria não interrompe a operação principal,
// mas é registrada no log da aplicação para alerta.
func (s *AuditService) Log(ctx context.Context, userID *int64, action, entity string, entityID *int64, oldData, newData map[string]any) {
	entry := &models.AuditLog{
		UserID:   userID,
		Action:   action,
		Entity:   entity,
		EntityID: entityID,
		OldData:  oldData,
		NewData:  newData,
	}
	// WithoutCancel: a trilha de auditoria deve ser gravada mesmo que o
	// cliente desconecte e o contexto da request seja cancelado.
	if err := s.repo.Insert(context.WithoutCancel(ctx), entry); err != nil {
		slog.Error("failed to write audit log",
			"action", action, "entity", entity, "error", err)
	}
}

func (s *AuditService) List(ctx context.Context, limit, offset int) ([]models.AuditLog, error) {
	return s.repo.List(ctx, limit, offset)
}
