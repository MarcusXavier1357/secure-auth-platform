package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/uptrace/bun"

	"auth-backend/internal/models"
)

var ErrNotFound = errors.New("record not found")

type UserRepository struct {
	db *bun.DB
}

func NewUserRepository(db *bun.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	user := new(models.User)
	err := r.db.NewSelect().
		Model(user).
		Relation("Role").
		Where("u.email = ?", email).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return user, err
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*models.User, error) {
	user := new(models.User)
	err := r.db.NewSelect().
		Model(user).
		Relation("Role").
		Relation("Permissions").
		Where("u.id = ?", id).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return user, err
}

func (r *UserRepository) List(ctx context.Context) ([]models.User, error) {
	var users []models.User
	err := r.db.NewSelect().
		Model(&users).
		Relation("Role").
		Relation("Permissions").
		Order("u.id ASC").
		Scan(ctx)
	return users, err
}

func (r *UserRepository) Count(ctx context.Context) (int, error) {
	return r.db.NewSelect().Model((*models.User)(nil)).Count(ctx)
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	_, err := r.db.NewInsert().Model(user).Exec(ctx)
	return err
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()
	res, err := r.db.NewUpdate().
		Model(user).
		Column("role_id", "name", "email", "password_hash", "active", "updated_at").
		WherePK().
		Exec(ctx)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return ErrNotFound
	}
	return nil
}
