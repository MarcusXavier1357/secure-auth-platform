package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"auth-backend/cache"
	"auth-backend/internal/geoip"
	"auth-backend/internal/models"
	"auth-backend/internal/password"
	"auth-backend/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrRateLimited        = errors.New("rate limited")
	ErrRateLimitDown      = errors.New("rate limiter unavailable")
	ErrInvalidSession     = errors.New("invalid session")
	ErrTokenReuse         = errors.New("refresh token reuse detected")
)

type TokenClaims struct {
	UserID    int64
	SessionID int64
}

type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	RefreshExpiresAt time.Time
	SessionID        int64
}

type AuthService struct {
	users         *repository.UserRepository
	sessions      *repository.SessionRepository
	permissions   *PermissionService
	tieredLimiter *cache.TieredRateLimiter
	lastLogin     *cache.LastLoginStore
	geo           geoip.Lookup
	audit         *AuditService

	jwtPrivateKey *rsa.PrivateKey
	jwtPublicKey  *rsa.PublicKey
	accessTTL     time.Duration
	refreshTTL    time.Duration
	travelWindow  time.Duration
}

func NewAuthService(
	users *repository.UserRepository,
	sessions *repository.SessionRepository,
	permissions *PermissionService,
	tieredLimiter *cache.TieredRateLimiter,
	lastLogin *cache.LastLoginStore,
	geo geoip.Lookup,
	audit *AuditService,
	keys *JWTKeyPair,
	accessTTL, refreshTTL, travelWindow time.Duration,
) *AuthService {
	if geo == nil {
		geo = geoip.Nop{}
	}
	if travelWindow == 0 {
		travelWindow = 30 * time.Minute
	}
	return &AuthService{
		users:         users,
		sessions:      sessions,
		permissions:   permissions,
		tieredLimiter: tieredLimiter,
		lastLogin:     lastLogin,
		geo:           geo,
		audit:         audit,
		jwtPrivateKey: keys.Private,
		jwtPublicKey:  keys.Public,
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
		travelWindow:  travelWindow,
	}
}

// Login aplica rate limit (antes do hash), valida credenciais, cria a
// sessão no PostgreSQL e retorna o par de tokens.
func (s *AuthService) Login(ctx context.Context, email, passwordPlain, ip, userAgent string) (*TokenPair, error) {
	ipKey := cache.LoginIPKey(ip)
	emailKey := cache.LoginEmailKey(email)

	if err := s.checkLoginRateLimit(ctx, ipKey, email, ip); err != nil {
		return nil, err
	}
	if err := s.checkLoginRateLimit(ctx, emailKey, email, ip); err != nil {
		return nil, err
	}

	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			password.DummyVerify(passwordPlain)
			s.auditLoginFailed(ctx, email, ip, "user_not_found")
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !user.Active {
		password.DummyVerify(passwordPlain)
		s.auditLoginFailed(ctx, email, ip, "user_inactive")
		return nil, ErrInvalidCredentials
	}

	needsRehash, err := password.Verify(user.PasswordHash, passwordPlain)
	if err != nil {
		slog.Warn("login password verification failed", "userId", user.ID, "error", err)
		s.auditLoginFailed(ctx, email, ip, "wrong_password")
		return nil, ErrInvalidCredentials
	}
	if needsRehash {
		newHash, err := password.Hash(passwordPlain)
		if err != nil {
			return nil, err
		}
		if err := s.users.UpdatePasswordHash(ctx, user.ID, newHash); err != nil {
			slog.Error("failed to rehash password to argon2id", "userId", user.ID, "error", err)
		}
	}

	pair, err := s.createSession(ctx, user, ip, userAgent)
	if err != nil {
		return nil, err
	}

	if err := s.tieredLimiter.Reset(ctx, emailKey); err != nil {
		slog.Warn("failed to reset login rate limit", "email", email, "error", err)
	}
	if err := s.tieredLimiter.Reset(ctx, ipKey); err != nil {
		slog.Warn("failed to reset login rate limit", "ip", ip, "error", err)
	}

	s.audit.Log(ctx, &user.ID, "login.success", "auth", &user.ID, nil,
		map[string]any{"email": email, "ip": ip, "sessionId": pair.SessionID})

	s.checkImpossibleTravel(ctx, user.ID, ip)
	s.permissions.InvalidateUserCache(ctx, user.ID)

	return pair, nil
}

func (s *AuthService) checkImpossibleTravel(ctx context.Context, userID int64, ip string) {
	if s.lastLogin == nil {
		return
	}

	country, err := s.geo.CountryCode(ip)
	if err != nil {
		slog.Warn("geoip lookup failed", "ip", ip, "error", err)
		return
	}
	if country == "" {
		return
	}

	prev, err := s.lastLogin.Get(ctx, userID)
	if err != nil {
		slog.Warn("failed to read last login", "userId", userID, "error", err)
	}
	if prev != nil && prev.CountryCode != "" && prev.CountryCode != country {
		elapsed := time.Since(prev.At)
		if elapsed <= s.travelWindow {
			s.audit.LogSecurityAlert(ctx, &userID, "impossible_travel", map[string]any{
				"ip":           ip,
				"country":      country,
				"prevIp":       prev.IP,
				"prevCountry":  prev.CountryCode,
				"elapsedSec":   elapsed.Seconds(),
				"windowMinutes": s.travelWindow.Minutes(),
			})
		}
	}

	if err := s.lastLogin.Set(ctx, userID, cache.LastLoginInfo{
		CountryCode: country,
		IP:          ip,
		At:          time.Now(),
	}); err != nil {
		slog.Warn("failed to store last login", "userId", userID, "error", err)
	}
}

func (s *AuthService) checkLoginRateLimit(ctx context.Context, key, email, ip string) error {
	blocked, retryAfter, count, err := s.tieredLimiter.Check(ctx, key)
	if err != nil {
		slog.Error("rate limiter unavailable during login", "error", err)
		return ErrRateLimitDown
	}
	if blocked {
		s.auditRateLimitBlock(ctx, email, ip, count, retryAfter)
		return ErrRateLimited
	}
	return nil
}

func (s *AuthService) auditRateLimitBlock(ctx context.Context, email, ip string, count int64, retryAfter time.Duration) {
	s.audit.Log(ctx, nil, "login.failed", "auth", nil, nil,
		map[string]any{"email": email, "ip": ip, "reason": "rate_limited", "retryAfterSec": retryAfter.Seconds()})

	if count >= s.tieredLimiter.MaxTierThreshold() {
		s.audit.LogSecurityAlert(ctx, nil, "login_rate_limit", map[string]any{
			"email":         email,
			"ip":            ip,
			"attemptCount":  count,
			"retryAfterSec": retryAfter.Seconds(),
		})
	}
}

// Refresh valida o refresh token, revoga a sessão anterior, cria uma nova
// e detecta reutilização de token já rotacionado.
func (s *AuthService) Refresh(ctx context.Context, refreshToken, ip, userAgent string) (*TokenPair, error) {
	tokenHash := hashToken(refreshToken)

	session, err := s.sessions.FindActiveByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return s.handleRefreshMiss(ctx, tokenHash, ip)
		}
		return nil, err
	}

	user, err := s.users.FindByID(ctx, session.UserID)
	if err != nil || !user.Active {
		return nil, ErrInvalidSession
	}

	oldSessionID := session.ID
	if err := s.sessions.RevokeWithTimestamp(ctx, session.ID); err != nil {
		return nil, err
	}

	pair, err := s.createSession(ctx, user, ip, userAgent)
	if err != nil {
		return nil, err
	}

	s.audit.Log(ctx, &user.ID, "session.refreshed", "session", &oldSessionID, nil,
		map[string]any{"oldSessionId": oldSessionID, "newSessionId": pair.SessionID})

	s.permissions.InvalidateUserCache(ctx, user.ID)

	return pair, nil
}

func (s *AuthService) handleRefreshMiss(ctx context.Context, tokenHash, ip string) (*TokenPair, error) {
	session, err := s.sessions.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidSession
		}
		return nil, err
	}

	// Token conhecido mas sessão revogada/expirada = possível roubo (reuse).
	if session.Revoked || session.ExpiresAt.Before(time.Now()) {
		userID := session.UserID
		if err := s.sessions.RevokeAllByUser(ctx, userID); err != nil {
			slog.Error("failed to revoke all sessions after token reuse", "userId", userID, "error", err)
		}
		s.audit.Log(ctx, &userID, "session.revoked", "session", &session.ID, nil,
			map[string]any{"reason": "refresh_token_reuse", "sessionId": session.ID})
		s.audit.LogSecurityAlert(ctx, &userID, "refresh_token_reuse", map[string]any{
			"sessionId": session.ID,
			"ip":        ip,
		})
		return nil, ErrTokenReuse
	}

	return nil, ErrInvalidSession
}

// Logout revoga a sessão associada ao refresh token.
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	session, err := s.sessions.FindActiveByTokenHash(ctx, hashToken(refreshToken))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		return err
	}

	if err := s.sessions.RevokeWithTimestamp(ctx, session.ID); err != nil {
		return err
	}

	s.audit.Log(ctx, &session.UserID, "logout", "auth", &session.UserID, nil,
		map[string]any{"sessionId": session.ID})
	s.audit.Log(ctx, &session.UserID, "session.revoked", "session", &session.ID, nil,
		map[string]any{"reason": "logout"})
	return nil
}

// ValidateAccessToken valida assinatura RS256 e retorna userId + sessionId.
func (s *AuthService) ValidateAccessToken(tokenString string) (TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtPublicKey, nil
	})
	if err != nil || !token.Valid {
		return TokenClaims{}, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return TokenClaims{}, errors.New("invalid claims")
	}
	sub, ok := claims["sub"].(float64)
	if !ok {
		return TokenClaims{}, errors.New("invalid subject")
	}
	sid, ok := claims["sid"].(float64)
	if !ok {
		return TokenClaims{}, errors.New("invalid session id")
	}
	return TokenClaims{UserID: int64(sub), SessionID: int64(sid)}, nil
}

// TouchSession atualiza last_activity_at da sessão (chamado pelo middleware Auth).
func (s *AuthService) TouchSession(ctx context.Context, sessionID int64) {
	if err := s.sessions.TouchActivity(ctx, sessionID); err != nil {
		slog.Warn("failed to touch session activity", "sessionId", sessionID, "error", err)
	}
}

func (s *AuthService) createSession(ctx context.Context, user *models.User, ip, userAgent string) (*TokenPair, error) {
	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().Add(s.refreshTTL)
	now := time.Now()

	var ipPtr, uaPtr *string
	if ip != "" {
		ipPtr = &ip
	}
	if userAgent != "" {
		uaPtr = &userAgent
	}

	session := &models.Session{
		UserID:           user.ID,
		RefreshTokenHash: hashToken(refreshToken),
		ExpiresAt:        expiresAt,
		IPAddress:        ipPtr,
		UserAgent:        uaPtr,
		LastActivityAt:   &now,
	}
	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, err
	}

	accessToken, err := s.generateAccessToken(user, session.ID)
	if err != nil {
		return nil, err
	}

	s.audit.Log(ctx, &user.ID, "session.created", "session", &session.ID, nil,
		map[string]any{"sessionId": session.ID, "ip": ip})

	return &TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		RefreshExpiresAt: expiresAt,
		SessionID:        session.ID,
	}, nil
}

func (s *AuthService) generateAccessToken(user *models.User, sessionID int64) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"sid":   sessionID,
		"email": user.Email,
		"iat":   now.Unix(),
		"exp":   now.Add(s.accessTTL).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(s.jwtPrivateKey)
}

func (s *AuthService) auditLoginFailed(ctx context.Context, email, ip, reason string) {
	s.audit.Log(ctx, nil, "login.failed", "auth", nil, nil,
		map[string]any{"email": email, "ip": ip, "reason": reason})
}

func generateRefreshToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
