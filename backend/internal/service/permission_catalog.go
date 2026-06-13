package service

import (
	"context"
	"errors"
	"strings"

	"auth-backend/internal/models"
	"auth-backend/internal/repository"
)

type CreatePermissionInput struct {
	Code        string
	Description string
}

func (s *PermissionService) Create(ctx context.Context, actorID int64, input CreatePermissionInput) (*models.Permission, error) {
	code := strings.TrimSpace(input.Code)
	if err := validatePermissionCode(code); err != nil {
		return nil, err
	}

	if _, err := s.permRepo.FindByCode(ctx, code); err == nil {
		return nil, ErrPermissionCodeTaken
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	perm := &models.Permission{
		Code:        code,
		Description: strings.TrimSpace(input.Description),
	}
	if err := s.permRepo.Create(ctx, perm); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, &actorID, "permission.created", "permission", &perm.ID, nil,
		map[string]any{"code": perm.Code, "description": perm.Description})
	return perm, nil
}

func (s *PermissionService) UpdateDescription(ctx context.Context, actorID, id int64, description string) (*models.Permission, error) {
	perm, err := s.permRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldData := map[string]any{"description": perm.Description}
	perm.Description = strings.TrimSpace(description)

	if err := s.permRepo.UpdateDescription(ctx, id, perm.Description); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, &actorID, "permission.updated", "permission", &id, oldData,
		map[string]any{"code": perm.Code, "description": perm.Description})
	return perm, nil
}

func (s *PermissionService) Delete(ctx context.Context, actorID, id int64) error {
	perm, err := s.permRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if isProtectedPermissionCode(perm.Code) {
		return ErrProtectedPermission
	}

	count, err := s.permRepo.CountAssignments(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrPermissionInUse
	}

	if err := s.permRepo.Delete(ctx, id); err != nil {
		return err
	}

	s.audit.Log(ctx, &actorID, "permission.deleted", "permission", &id,
		map[string]any{"code": perm.Code, "description": perm.Description}, nil)
	return nil
}

func (s *PermissionService) actorCanGrant(ctx context.Context, actorID int64, permCode string) (bool, error) {
	codes, err := s.permRepo.ListCodesByUserID(ctx, actorID)
	if err != nil {
		return false, err
	}
	return matchPermission(codes, permCode), nil
}
