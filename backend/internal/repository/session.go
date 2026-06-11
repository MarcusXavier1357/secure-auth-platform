package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/uptrace/bun"

	"auth-backend/internal/models"
)

type SessionRepository struct {
	db *bun.DB
}

func NewSessionRepository(db *bun.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, session *models.Session) error {
	_, err := r.db.NewInsert().Model(session).Exec(ctx)
	return err
}

// FindActiveByTokenHash busca uma sessão válida (não revogada e não expirada)
// pelo hash SHA-256 do refresh token.
func (r *SessionRepository) FindActiveByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	session := new(models.Session)
	err := r.db.NewSelect().
		Model(session).
		Where("s.refresh_token_hash = ?", tokenHash).
		Where("s.revoked = ?", false).
		Where("s.expires_at > ?", time.Now()).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return session, err
}

// FindByTokenHash busca qualquer sessão pelo hash (inclui revogadas/expiradas).
func (r *SessionRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	session := new(models.Session)
	err := r.db.NewSelect().
		Model(session).
		Where("s.refresh_token_hash = ?", tokenHash).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return session, err
}

func (r *SessionRepository) FindByID(ctx context.Context, id int64) (*models.Session, error) {
	session := new(models.Session)
	err := r.db.NewSelect().
		Model(session).
		Where("s.id = ?", id).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return session, err
}

func (r *SessionRepository) RevokeWithTimestamp(ctx context.Context, sessionID int64) error {
	now := time.Now()
	_, err := r.db.NewUpdate().
		Model((*models.Session)(nil)).
		Set("revoked = ?", true).
		Set("revoked_at = ?", now).
		Where("id = ?", sessionID).
		Where("revoked = ?", false).
		Exec(ctx)
	return err
}

func (r *SessionRepository) TouchActivity(ctx context.Context, sessionID int64) error {
	_, err := r.db.NewUpdate().
		Model((*models.Session)(nil)).
		Set("last_activity_at = ?", time.Now()).
		Where("id = ?", sessionID).
		Where("revoked = ?", false).
		Exec(ctx)
	return err
}

func (r *SessionRepository) CountByUserID(ctx context.Context, userID int64) (int, error) {
	count, err := r.db.NewSelect().
		Model((*models.Session)(nil)).
		Where("user_id = ?", userID).
		Count(ctx)
	return count, err
}

func (r *SessionRepository) CountActiveByUserID(ctx context.Context, userID int64) (int, error) {
	count, err := r.db.NewSelect().
		Model((*models.Session)(nil)).
		Where("user_id = ?", userID).
		Where("revoked = ?", false).
		Where("expires_at > ?", time.Now()).
		Count(ctx)
	return count, err
}

// DeleteExpired remove sessões expiradas ou revogadas há mais tempo que a
// retenção. Retorna as sessões removidas para auditoria.
func (r *SessionRepository) DeleteExpired(ctx context.Context, retention time.Duration) (int64, []models.Session, error) {
	cutoff := time.Now().Add(-retention)

	var expired []models.Session
	err := r.db.NewSelect().
		Model(&expired).
		Where("expires_at < ?", cutoff).
		WhereOr("revoked = ? AND COALESCE(revoked_at, created_at) < ?", true, cutoff).
		Scan(ctx)
	if err != nil {
		return 0, nil, err
	}

	res, err := r.db.NewDelete().
		Model((*models.Session)(nil)).
		Where("expires_at < ?", cutoff).
		WhereOr("revoked = ? AND COALESCE(revoked_at, created_at) < ?", true, cutoff).
		Exec(ctx)
	if err != nil {
		return 0, nil, err
	}
	n, _ := res.RowsAffected()
	return n, expired, nil
}

func (r *SessionRepository) RevokeAllByUser(ctx context.Context, userID int64) error {
	now := time.Now()
	_, err := r.db.NewUpdate().
		Model((*models.Session)(nil)).
		Set("revoked = ?", true).
		Set("revoked_at = ?", now).
		Where("user_id = ?", userID).
		Where("revoked = ?", false).
		Exec(ctx)
	return err
}
