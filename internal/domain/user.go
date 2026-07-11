package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	OrganizationID      *uuid.UUID     `gorm:"type:uuid;index" json:"organization_id,omitempty"`
	RoleID              *uuid.UUID     `gorm:"type:uuid;index" json:"role_id,omitempty"`
	Username            string         `gorm:"not null" json:"username"`
	Name                string         `gorm:"not null" json:"name"`
	Password            string         `gorm:"not null" json:"-"`
	IsAdmin             bool           `gorm:"not null;default:false" json:"is_admin"`
	Role                *Role          `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	ChangelogLastSeenAt *time.Time     `gorm:"column:changelog_last_seen_at" json:"-"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
	DisabledAt          *time.Time     `gorm:"index" json:"disabled_at,omitempty"`
	DisabledBy          *uuid.UUID     `gorm:"type:uuid" json:"disabled_by,omitempty"`
	DisabledReason      *string        `json:"disabled_reason,omitempty"`
}

type CreateUserDTO struct {
	OrganizationID *uuid.UUID `json:"organization_id"`
	Username       string     `json:"username" binding:"required,username"`
	Name           string     `json:"name" binding:"required,min=2"`
	Password       string     `json:"password" binding:"required,min=8"`
	IsAdmin        bool       `json:"is_admin"`
	RoleID         *uuid.UUID `json:"role_id"`
}

type UpdateUserDTO struct {
	Username *string    `json:"username" binding:"omitempty,username"`
	Name     *string    `json:"name" binding:"omitempty,min=2"`
	RoleID   *uuid.UUID `json:"role_id"`
}

type ChangePasswordDTO struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,strongpassword"`
}

// ChangePasswordResponse carries a freshly-issued token. Changing the password
// revokes every session issued before the change (see SessionTokenService); the
// current device swaps in this token so it stays signed in while other devices
// are logged out.
type ChangePasswordResponse struct {
	Token string `json:"token"`
}

type LoginDTO struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AdminCreateUserDTO struct {
	OrganizationID *uuid.UUID `json:"organization_id"`
	Username       string     `json:"username" binding:"required,username"`
	Name           string     `json:"name" binding:"required,min=2"`
	Password       string     `json:"password" binding:"required,strongpassword"`
	IsAdmin        bool       `json:"is_admin"`
	RoleID         *uuid.UUID `json:"role_id"`
}

type AdminUpdateUserDTO struct {
	Username *string    `json:"username" binding:"omitempty,username"`
	Name     *string    `json:"name" binding:"omitempty,min=2"`
	Password *string    `json:"password" binding:"omitempty,strongpassword"`
	IsAdmin  *bool      `json:"is_admin"`
	RoleID   *uuid.UUID `json:"role_id"`
}

type AdminForceResetPasswordDTO struct {
	NewPassword string `json:"new_password" binding:"required,strongpassword"`
}

// UserListScope is the role-resolved scope produced by the service for
// GET /users. The repository is role-agnostic and only translates this
// into SQL filters.
type UserListScope struct {
	All            bool
	OrganizationID *uuid.UUID
	UserID         *uuid.UUID
	IncludeDeleted bool
	// Disabled filters by lockout state: nil = all, false = active only,
	// true = disabled only. Not applied to status-count queries.
	Disabled *bool
}

// UserStatusCounts breaks the caller-scoped user set down by lockout state.
// All = Active + Disabled. Backs the status tabs on GET /users.
type UserStatusCounts struct {
	All      int64 `json:"all"`
	Active   int64 `json:"active"`
	Disabled int64 `json:"disabled"`
}

type AdminListUsersQuery struct {
	OrganizationID *uuid.UUID `form:"-"`
	RoleID         *uuid.UUID `form:"-"`
	IsAdmin        *bool      `form:"is_admin"`
	Disabled       *bool      `form:"disabled"`
	IncludeDeleted bool       `form:"include_deleted"`
	ListParams     ListParams `form:"-"`
}

type AssignRoleDTO struct {
	RoleID uuid.UUID `json:"role_id" binding:"required"`
}

type DisableUserDTO struct {
	Reason string `json:"reason" binding:"omitempty,max=500"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByIDWithPermissions(ctx context.Context, id uuid.UUID) (*User, error)
	// FindByUsernameAndOrg looks up an active org member by (org, username).
	FindByUsernameAndOrg(ctx context.Context, username string, orgID uuid.UUID) (*User, error)
	// SearchActiveInOrg returns up to `limit` active (non-disabled) users in the
	// org whose username or name ILIKE-matches query, ordered by username asc.
	// An empty query returns the first page of active org users.
	SearchActiveInOrg(ctx context.Context, orgID uuid.UUID, query string, limit int) ([]User, error)
	// FilterIDsInOrg returns the subset of ids that belong to users in the org,
	// in one query. Unknown and cross-org ids are silently dropped.
	FilterIDsInOrg(ctx context.Context, orgID uuid.UUID, ids []uuid.UUID) ([]uuid.UUID, error)
	// FindAdminByUsername looks up a platform admin (org_id IS NULL, is_admin).
	FindAdminByUsername(ctx context.Context, username string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, scope UserListScope, p ListParams) ([]User, int64, error)
	StatusCounts(ctx context.Context, scope UserListScope) (UserStatusCounts, error)

	// Admin-only operations.
	HardDelete(ctx context.Context, id uuid.UUID) error
	FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*User, error)
	AdminList(ctx context.Context, q AdminListUsersQuery) ([]User, int64, error)
	CountAll(ctx context.Context) (int64, error)
}

type UserService interface {
	Create(ctx context.Context, dto CreateUserDTO) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	Update(ctx context.Context, id uuid.UUID, dto UpdateUserDTO) (*User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, p ListParams, disabled *bool) ([]User, int64, error)
	StatusCounts(ctx context.Context) (UserStatusCounts, error)
	GetProfile(ctx context.Context, id uuid.UUID) (*User, error)
	// ChangePassword verifies the current password, sets the new one, revokes
	// all sessions issued before now, and returns a fresh token for the caller's
	// current device so it stays signed in.
	ChangePassword(ctx context.Context, id uuid.UUID, dto ChangePasswordDTO) (string, error)
	AssignRole(ctx context.Context, userID uuid.UUID, dto AssignRoleDTO) (*User, error)
	RemoveRole(ctx context.Context, userID uuid.UUID) (*User, error)
	Disable(ctx context.Context, id uuid.UUID, dto DisableUserDTO) (*User, error)
	Enable(ctx context.Context, id uuid.UUID) (*User, error)

	// Admin surface.
	AdminGetByID(ctx context.Context, id uuid.UUID) (*User, error)
	AdminList(ctx context.Context, q AdminListUsersQuery) ([]User, int64, error)
	AdminCreate(ctx context.Context, dto AdminCreateUserDTO) (*User, error)
	AdminUpdate(ctx context.Context, id uuid.UUID, dto AdminUpdateUserDTO) (*User, error)
	AdminForceResetPassword(ctx context.Context, id uuid.UUID, dto AdminForceResetPasswordDTO) error
	AdminHardDelete(ctx context.Context, id uuid.UUID) error
}

// SessionTokenService issues auth tokens and revokes a user's own sessions.
// It is implemented in the auth package and injected into features that must
// re-issue or invalidate a caller's own token after proving identity
// themselves (e.g. self password change) — no caller authz is performed here.
type SessionTokenService interface {
	GenerateToken(userID uuid.UUID) (string, error)
	// RevokeUserSessions invalidates every token issued before now for userID.
	RevokeUserSessions(ctx context.Context, userID uuid.UUID) error
}

type AuthService interface {
	// Login authenticates within a host scope. orgID nil = admin-host login
	// (org_id IS NULL, is_admin); non-nil = tenant-host login scoped to that org.
	Login(ctx context.Context, dto LoginDTO, orgID *uuid.UUID) (*User, string, error)
	AdminRevokeSessions(ctx context.Context, userID uuid.UUID) error
}
