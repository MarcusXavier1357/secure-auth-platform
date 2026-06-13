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
	if input.Active != nil && !*input.Active && actorID == userID {
		return nil, ErrCannotDeactivateSelf
	}

	if input.Name != nil || input.Email != nil || input.RoleID != nil {
		if err := s.requirePermission(ctx, actorID, "users.update"); err != nil {
			return nil, err
		}
	}
	if input.Password != nil {
		if err := s.requirePermission(ctx, actorID, "users.password.reset"); err != nil {
			return nil, err
		}
	}
	if input.Active != nil {
		if err := s.requirePermission(ctx, actorID, "users.deactivate"); err != nil {
			return nil, err
		}
	}

	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if input.Active != nil && !*input.Active && user.Active {
		hasStar, err := s.userHasPermissionCode(ctx, userID, "*")
		if err != nil {
			return nil, err
		}
		if hasStar {
			count, err := s.permissions.CountActiveUsersWithStar(ctx)
			if err != nil {
				return nil, err
			}
			if count <= 1 {
				return nil, ErrLastAdmin
			}
		}
	}

	oldData := map[string]any{
		"name": user.Name, "email": user.Email,
		"roleId": user.RoleID, "active": user.Active,
	}

	wasActive := user.Active
	deactivated := false
	activated := false
	passwordChanged := false

	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Email != nil {
		if err := validateEmail(*input.Email); err != nil {
			return nil, err
		}
		if *input.Email != user.Email {
			if existing, err := s.users.FindByEmail(ctx, *input.Email); err == nil && existing.ID != userID {
				return nil, ErrEmailTaken
			} else if err != nil && !errors.Is(err, repository.ErrNotFound) {
				return nil, err
			}
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
		passwordChanged = true
	}

	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}

	if deactivated {
		if err := s.sessions.RevokeAllByUser(ctx, userID); err != nil {
			slog.Error("failed to revoke sessions for deactivated user",
				"userId", userID, "error", err)
		}
		s.permissions.InvalidateUserCache(ctx, userID)
	}

	newData := map[string]any{
		"name": user.Name, "email": user.Email,
		"roleId": user.RoleID, "active": user.Active,
	}
	if passwordChanged {
		newData["passwordChanged"] = true
	}

	s.audit.Log(ctx, &actorID, "user.updated", "user", &userID, oldData, newData)

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
