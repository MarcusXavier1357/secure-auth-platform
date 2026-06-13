package service

import (
	"context"
	"errors"
	"log/slog"

	"auth-backend/internal/models"
	"auth-backend/internal/password"
	"auth-backend/internal/repository"
)

var (
	ErrEmailTaken           = errors.New("email already in use")
	ErrCannotDeactivateSelf = errors.New("cannot deactivate own account")
)

type CreateUserInput struct {
	Name     string
	Email    string
	Password string
	RoleID   *int64
}

type UpdateUserInput struct {
	Name     *string
	Email    *string
	Password *string
	RoleID   *int64
	Active   *bool
}

type UserService struct {
	users       *repository.UserRepository
	sessions    *repository.SessionRepository
	permissions *PermissionService
	audit       *AuditService
}

func NewUserService(
	users *repository.UserRepository,
	sessions *repository.SessionRepository,
	permissions *PermissionService,
	audit *AuditService,
) *UserService {
	return &UserService{users: users, sessions: sessions, permissions: permissions, audit: audit}
}

func (s *UserService) List(ctx context.Context) ([]models.User, error) {
	return s.users.List(ctx)
}

func (s *UserService) FindByID(ctx context.Context, id int64) (*models.User, error) {
	return s.users.FindByID(ctx, id)
}

func (s *UserService) Create(ctx context.Context, actorID int64, input CreateUserInput) (*models.User, error) {
	if err := validateEmail(input.Email); err != nil {
		return nil, err
	}
	if err := validatePassword(input.Password); err != nil {
		return nil, err
	}

	if _, err := s.users.FindByEmail(ctx, input.Email); err == nil {
		return nil, ErrEmailTaken
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	hash, err := password.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: hash,
		RoleID:       input.RoleID,
		Active:       true,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, &actorID, "user.created", "user", &user.ID, nil,
		map[string]any{"name": user.Name, "email": user.Email, "roleId": user.RoleID})

	return user, nil
}

func (s *UserService) Update(ctx context.Context, actorID, userID int64, input UpdateUserInput) (*models.User, error) {
	// Auto-desativação causaria lockout (ex.: último admin do sistema).
	if input.Active != nil && !*input.Active && actorID == userID {
		return nil, ErrCannotDeactivateSelf
	}

	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	oldData := map[string]any{
		"name": user.Name, "email": user.Email,
		"roleId": user.RoleID, "active": user.Active,
	}

	wasActive := user.Active
	deactivated := false
	activated := false
	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Email != nil {
		if err := validateEmail(*input.Email); err != nil {
			return nil, err
		}
		user.Email = *input.Email
	}
	if input.RoleID != nil {
		user.RoleID = input.RoleID
	}
	if input.Active != nil {
		deactivated = user.Active && !*input.Active
		activated = !user.Active && *input.Active
		user.Active = *input.Active
	}
	if input.Password != nil {
		if err := validatePassword(*input.Password); err != nil {
			return nil, err
		}
		hash, err := password.Hash(*input.Password)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = hash
	}

	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}

	if deactivated {
		// Usuário desativado: revogar sessões e invalidar cache de permissões.
		if err := s.sessions.RevokeAllByUser(ctx, userID); err != nil {
			slog.Error("failed to revoke sessions for deactivated user",
				"userId", userID, "error", err)
		}
		s.permissions.InvalidateUserCache(ctx, userID)
	}

	s.audit.Log(ctx, &actorID, "user.updated", "user", &userID, oldData,
		map[string]any{
			"name": user.Name, "email": user.Email,
			"roleId": user.RoleID, "active": user.Active,
		})

	if deactivated && wasActive {
		s.audit.Log(ctx, &actorID, "user.deactivated", "user", &userID, oldData,
			map[string]any{"active": false})
	}
	if activated {
		s.audit.Log(ctx, &actorID, "user.activated", "user", &userID, oldData,
			map[string]any{"active": true})
	}

	return user, nil
}
