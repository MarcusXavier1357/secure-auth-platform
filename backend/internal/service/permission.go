package service

import (
	"context"
	"log/slog"
	"auth-backend/cache"
	"auth-backend/internal/models"
	"auth-backend/internal/repository"
)

type PermissionService struct {
	permRepo  *repository.PermissionRepository
	userRepo  *repository.UserRepository
	roleRepo  *repository.RoleRepository
	permCache *cache.PermissionCache
	audit     *AuditService
}

func NewPermissionService(
	permRepo *repository.PermissionRepository,
	userRepo *repository.UserRepository,
	roleRepo *repository.RoleRepository,
	permCache *cache.PermissionCache,
	audit *AuditService,
) *PermissionService {
	return &PermissionService{permRepo: permRepo, userRepo: userRepo, roleRepo: roleRepo, permCache: permCache, audit: audit}
}

// HasPermission verifica via cache-aside: Redis primeiro, PostgreSQL no miss.
// Se o Redis estiver indisponível, consulta o PostgreSQL diretamente
// (degraded mode) e registra alerta — auth continua funcionando.
func (s *PermissionService) HasPermission(ctx context.Context, userID int64, code string) (bool, error) {
	codes, found, err := s.permCache.Get(ctx, userID)
	if err != nil {
		slog.Error("redis unavailable for permission check, falling back to postgres", "error", err)
		codes, err = s.permRepo.ListCodesByUserID(ctx, userID)
		if err != nil {
			return false, err
		}
		return matchPermission(codes, code), nil
	}

	if !found {
		codes, err = s.permRepo.ListCodesByUserID(ctx, userID)
		if err != nil {
			return false, err
		}
		if cacheErr := s.permCache.Set(ctx, userID, codes); cacheErr != nil {
			slog.Error("failed to populate permission cache", "userId", userID, "error", cacheErr)
		}
	}

	return matchPermission(codes, code), nil
}

func (s *PermissionService) ListAll(ctx context.Context) ([]models.Permission, error) {
	return s.permRepo.ListAll(ctx)
}

func (s *PermissionService) ListCodesByUser(ctx context.Context, userID int64) ([]string, error) {
	return s.permRepo.ListCodesByUserID(ctx, userID)
}

func (s *PermissionService) Grant(ctx context.Context, actorID, userID, permissionID int64) error {
	if _, err := s.userRepo.FindByID(ctx, userID); err != nil {
		return err
	}
	perm, err := s.permRepo.FindByID(ctx, permissionID)
	if err != nil {
		return err
	}

	allowed, err := s.actorCanGrant(ctx, actorID, perm.Code)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}

	if err := s.permRepo.Grant(ctx, userID, permissionID); err != nil {
		return err
	}
	s.invalidateCache(ctx, userID)
	s.audit.Log(ctx, &actorID, "permission.granted", "user_permissions", &userID,
		nil, map[string]any{
			"userId": userID, "permissionId": permissionID, "permissionCode": perm.Code,
		})
	return nil
}

func (s *PermissionService) Revoke(ctx context.Context, actorID, userID, permissionID int64) error {
	if _, err := s.userRepo.FindByID(ctx, userID); err != nil {
		return err
	}
	perm, err := s.permRepo.FindByID(ctx, permissionID)
	if err != nil {
		return err
	}
	if err := s.permRepo.Revoke(ctx, userID, permissionID); err != nil {
		return err
	}
	s.invalidateCache(ctx, userID)
	s.audit.Log(ctx, &actorID, "permission.revoked", "user_permissions", &userID,
		map[string]any{
			"userId": userID, "permissionId": permissionID, "permissionCode": perm.Code,
		}, nil)
	return nil
}

// InvalidateUserCache deve ser chamado ao desativar usuário ou revogar
// todas as suas sessões.
func (s *PermissionService) InvalidateUserCache(ctx context.Context, userID int64) {
	s.invalidateCache(ctx, userID)
}

func (s *PermissionService) CountActiveUsersWithStar(ctx context.Context) (int, error) {
	return s.permRepo.CountActiveUsersWithCode(ctx, "*")
}

func (s *PermissionService) invalidateCache(ctx context.Context, userID int64) {
	if err := s.permCache.Invalidate(ctx, userID); err != nil {
		slog.Error("failed to invalidate permission cache", "userId", userID, "error", err)
	}
}
