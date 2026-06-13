package service

import (
	"context"

	"auth-backend/internal/models"
)

func (s *PermissionService) ListRoles(ctx context.Context) ([]models.Role, error) {
	return s.roleRepo.List(ctx)
}
