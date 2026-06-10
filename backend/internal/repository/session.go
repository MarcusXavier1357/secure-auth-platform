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

// RotateToken substitui o hash do refresh token na mesma sessão (rotação).
func (r *SessionRepository) RotateToken(ctx context.Context, sessionID int64, newHash string, newExpiry time.Time) error {
	_, err := r.db.NewUpdate().
		Model((*models.Session)(nil)).
		Set("refresh_token_hash = ?", newHash).
		Set("expires_at = ?", newExpiry).
		Where("id = ?", sessionID).
		Exec(ctx)
	return err
}

func (r *SessionRepository) Revoke(ctx context.Context, sessionID int64) error {
	_, err := r.db.NewUpdate().
		Model((*models.Session)(nil)).
		Set("revoked = ?", true).
		Where("id = ?", sessionID).
		Exec(ctx)
	return err
}

// DeleteExpired remove sessões expiradas ou revogadas há mais tempo que a
// retenção. Retorna o número de linhas removidas.
func (r *SessionRepository) DeleteExpired(ctx context.Context, retention time.Duration) (int64, error) {
	cutoff := time.Now().Add(-retention)
	res, err := r.db.NewDelete().
		Model((*models.Session)(nil)).
		Where("expires_at < ?", cutoff).
		WhereOr("revoked = ? AND created_at < ?", true, cutoff).
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *SessionRepository) RevokeAllByUser(ctx context.Context, userID int64) error {
	_, err := r.db.NewUpdate().
		Model((*models.Session)(nil)).
		Set("revoked = ?", true).
		Where("user_id = ?", userID).
		Where("revoked = ?", false).
		Exec(ctx)
	return err
}
