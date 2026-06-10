package repository

import (
	"context"

	"github.com/uptrace/bun"

	"auth-backend/internal/models"
)

type AuditRepository struct {
	db *bun.DB
}

func NewAuditRepository(db *bun.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Insert(ctx context.Context, log *models.AuditLog) error {
	_, err := r.db.NewInsert().Model(log).Exec(ctx)
	return err
}

func (r *AuditRepository) List(ctx context.Context, limit, offset int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := r.db.NewSelect().
		Model(&logs).
		Order("al.created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(ctx)
	return logs, err
}
