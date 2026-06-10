package models

import (
	"time"

	"github.com/uptrace/bun"
)

type Role struct {
	bun.BaseModel `bun:"table:roles,alias:r"`

	ID          int64     `bun:"id,pk,autoincrement" json:"id"`
	Name        string    `bun:"name,notnull,unique" json:"name"`
	Description string    `bun:"description" json:"description"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:current_timestamp" json:"createdAt"`
}

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID           int64     `bun:"id,pk,autoincrement" json:"id"`
	RoleID       *int64    `bun:"role_id" json:"roleId"`
	Name         string    `bun:"name,notnull" json:"name"`
	Email        string    `bun:"email,notnull,unique" json:"email"`
	PasswordHash string    `bun:"password_hash,notnull" json:"-"`
	Active       bool      `bun:"active,notnull,default:true" json:"active"`
	CreatedAt    time.Time `bun:"created_at,notnull,default:current_timestamp" json:"createdAt"`
	UpdatedAt    time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updatedAt"`

	Role        *Role        `bun:"rel:belongs-to,join:role_id=id" json:"role,omitempty"`
	Permissions []Permission `bun:"m2m:user_permissions,join:User=Permission" json:"permissions,omitempty"`
}

type Permission struct {
	bun.BaseModel `bun:"table:permissions,alias:p"`

	ID          int64     `bun:"id,pk,autoincrement" json:"id"`
	Code        string    `bun:"code,notnull,unique" json:"code"`
	Description string    `bun:"description" json:"description"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:current_timestamp" json:"createdAt"`
}

type UserPermission struct {
	bun.BaseModel `bun:"table:user_permissions,alias:up"`

	UserID       int64       `bun:"user_id,pk"`
	User         *User       `bun:"rel:belongs-to,join:user_id=id"`
	PermissionID int64       `bun:"permission_id,pk"`
	Permission   *Permission `bun:"rel:belongs-to,join:permission_id=id"`
}

type Session struct {
	bun.BaseModel `bun:"table:sessions,alias:s"`

	ID               int64     `bun:"id,pk,autoincrement" json:"id"`
	UserID           int64     `bun:"user_id,notnull" json:"userId"`
	RefreshTokenHash string    `bun:"refresh_token_hash,notnull" json:"-"`
	ExpiresAt        time.Time `bun:"expires_at,notnull" json:"expiresAt"`
	Revoked          bool      `bun:"revoked,notnull,default:false" json:"revoked"`
	CreatedAt        time.Time `bun:"created_at,notnull,default:current_timestamp" json:"createdAt"`
}

type AuditLog struct {
	bun.BaseModel `bun:"table:audit_logs,alias:al"`

	ID        int64          `bun:"id,pk,autoincrement" json:"id"`
	UserID    *int64         `bun:"user_id" json:"userId"`
	Action    string         `bun:"action,notnull" json:"action"`
	Entity    string         `bun:"entity,notnull" json:"entity"`
	EntityID  *int64         `bun:"entity_id" json:"entityId"`
	OldData   map[string]any `bun:"old_data,type:jsonb" json:"oldData"`
	NewData   map[string]any `bun:"new_data,type:jsonb" json:"newData"`
	CreatedAt time.Time      `bun:"created_at,notnull,default:current_timestamp" json:"createdAt"`
}
