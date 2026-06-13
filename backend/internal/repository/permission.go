package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/uptrace/bun"

	"auth-backend/internal/models"
)

type PermissionRepository struct {
	db *bun.DB
}

func NewPermissionRepository(db *bun.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

func (r *PermissionRepository) ListAll(ctx context.Context) ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.NewSelect().
		Model(&permissions).
		Order("p.code ASC").
		Scan(ctx)
	return permissions, err
}

func (r *PermissionRepository) FindByCode(ctx context.Context, code string) (*models.Permission, error) {
	permission := new(models.Permission)
	err := r.db.NewSelect().
		Model(permission).
		Where("p.code = ?", code).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return permission, err
}

func (r *PermissionRepository) FindByID(ctx context.Context, id int64) (*models.Permission, error) {
	permission := new(models.Permission)
	err := r.db.NewSelect().
		Model(permission).
		Where("p.id = ?", id).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return permission, err
}

func (r *PermissionRepository) ListCodesByUserID(ctx context.Context, userID int64) ([]string, error) {
	var codes []string
	err := r.db.NewSelect().
		Model((*models.Permission)(nil)).
		Column("p.code").
		Join("JOIN user_permissions AS up ON up.permission_id = p.id").
		Where("up.user_id = ?", userID).
		Scan(ctx, &codes)
	return codes, err
}

func (r *PermissionRepository) Grant(ctx context.Context, userID, permissionID int64) error {
	_, err := r.db.NewInsert().
		Model(&models.UserPermission{UserID: userID, PermissionID: permissionID}).
		On("CONFLICT DO NOTHING").
		Exec(ctx)
	return err
}

func (r *PermissionRepository) Revoke(ctx context.Context, userID, permissionID int64) error {
	_, err := r.db.NewDelete().
		Model((*models.UserPermission)(nil)).
		Where("user_id = ?", userID).
		Where("permission_id = ?", permissionID).
		Exec(ctx)
	return err
}

func (r *PermissionRepository) GrantAll(ctx context.Context, userID int64) error {
	_, err := r.db.NewRaw(
		`INSERT INTO user_permissions (user_id, permission_id)
		 SELECT ?, id FROM permissions
		 ON CONFLICT DO NOTHING`,
		userID,
	).Exec(ctx)
	return err
}

func (r *PermissionRepository) Create(ctx context.Context, permission *models.Permission) error {
	_, err := r.db.NewInsert().Model(permission).Exec(ctx)
	return err
}

func (r *PermissionRepository) UpdateDescription(ctx context.Context, id int64, description string) error {
	res, err := r.db.NewUpdate().
		Model((*models.Permission)(nil)).
		Set("description = ?", description).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PermissionRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.NewDelete().
		Model((*models.Permission)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PermissionRepository) CountAssignments(ctx context.Context, permissionID int64) (int, error) {
	return r.db.NewSelect().
		Model((*models.UserPermission)(nil)).
		Where("permission_id = ?", permissionID).
		Count(ctx)
}

func (r *PermissionRepository) CountActiveUsersWithCode(ctx context.Context, code string) (int, error) {
	return r.db.NewSelect().
		Model((*models.User)(nil)).
		Join("JOIN user_permissions AS up ON up.user_id = u.id").
		Join("JOIN permissions AS p ON p.id = up.permission_id").
		Where("p.code = ?", code).
		Where("u.active = TRUE").
		Count(ctx)
}
