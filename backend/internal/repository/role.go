package repository

import (
	"context"

	"github.com/uptrace/bun"

	"auth-backend/internal/models"
)

type RoleRepository struct {
	db *bun.DB
}

func NewRoleRepository(db *bun.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) List(ctx context.Context) ([]models.Role, error) {
	var roles []models.Role
	err := r.db.NewSelect().
		Model(&roles).
		Order("r.name ASC").
		Scan(ctx)
	return roles, err
}
