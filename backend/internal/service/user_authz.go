package service

import (
	"context"
)

func (s *UserService) requirePermission(ctx context.Context, actorID int64, code string) error {
	ok, err := s.permissions.HasPermission(ctx, actorID, code)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}

func (s *UserService) userHasPermissionCode(ctx context.Context, userID int64, code string) (bool, error) {
	codes, err := s.permissions.ListCodesByUser(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, c := range codes {
		if c == code {
			return true, nil
		}
	}
	return false, nil
}
