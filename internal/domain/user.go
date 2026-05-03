package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	OrganizationID *uuid.UUID     `gorm:"type:uuid;index" json:"organization_id,omitempty"`
	RoleID         *uuid.UUID     `gorm:"type:uuid;index" json:"role_id,omitempty"`
	Username       string         `gorm:"not null" json:"username"`
	Name           string         `gorm:"not null" json:"name"`
	Password       string         `gorm:"not null" json:"-"`
	IsAdmin        bool           `gorm:"not null;default:false" json:"is_admin"`
	Role           *Role          `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type CreateUserDTO struct {
	OrganizationID *uuid.UUID `json:"organization_id"`
	Username       string     `json:"username" binding:"required,min=3"`
	Name           string     `json:"name" binding:"required,min=2"`
	Password       string     `json:"password" binding:"required,min=6"`
	IsAdmin        bool       `json:"is_admin"`
	RoleID         *uuid.UUID `json:"role_id"`
}

type UpdateUserDTO struct {
	Username *string    `json:"username" binding:"omitempty,min=3"`
	Name     *string    `json:"name" binding:"omitempty,min=2"`
	RoleID   *uuid.UUID `json:"role_id"`
}

type UpdateProfileDTO struct {
	Name string `json:"name" binding:"omitempty,min=2"`
}

type ChangePasswordDTO struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,strongpassword"`
}

type LoginDTO struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AdminCreateUserDTO struct {
	OrganizationID *uuid.UUID `json:"organization_id"`
	Username       string     `json:"username" binding:"required,min=3"`
	Name           string     `json:"name" binding:"required,min=2"`
	Password       string     `json:"password" binding:"required,strongpassword"`
	IsAdmin        bool       `json:"is_admin"`
}

type AdminUpdateUserDTO struct {
	Username *string    `json:"username" binding:"omitempty,min=3"`
	Name     *string    `json:"name" binding:"omitempty,min=2"`
	Password *string    `json:"password" binding:"omitempty,strongpassword"`
	IsAdmin  *bool      `json:"is_admin"`
	RoleID   *uuid.UUID `json:"role_id"`
}

type AdminForceResetPasswordDTO struct {
	NewPassword string `json:"new_password" binding:"required,strongpassword"`
}

type AdminListUsersQuery struct {
	OrganizationID string `form:"organization_id"`
	IsAdmin        *bool  `form:"is_admin"`
	IncludeDeleted bool   `form:"include_deleted"`
	ListParams     ListParams `form:"-"`
}

type ListUsersQuery struct {
	OrganizationID string `form:"organization_id"`
	ListParams     ListParams `form:"-"`
}

type AssignRoleDTO struct {
	RoleID uuid.UUID `json:"role_id" binding:"required"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, q ListUsersQuery) ([]User, int64, error)

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
	List(ctx context.Context, q ListUsersQuery) ([]User, int64, error)
	GetProfile(ctx context.Context, id uuid.UUID) (*User, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, dto UpdateProfileDTO) (*User, error)
	ChangePassword(ctx context.Context, id uuid.UUID, dto ChangePasswordDTO) error
	AssignRole(ctx context.Context, userID uuid.UUID, dto AssignRoleDTO) (*User, error)
	RemoveRole(ctx context.Context, userID uuid.UUID) (*User, error)

	// Admin surface.
	AdminGetByID(ctx context.Context, id uuid.UUID) (*User, error)
	AdminList(ctx context.Context, q AdminListUsersQuery) ([]User, int64, error)
	AdminCreate(ctx context.Context, dto AdminCreateUserDTO) (*User, error)
	AdminUpdate(ctx context.Context, id uuid.UUID, dto AdminUpdateUserDTO) (*User, error)
	AdminForceResetPassword(ctx context.Context, id uuid.UUID, dto AdminForceResetPasswordDTO) error
	AdminHardDelete(ctx context.Context, id uuid.UUID) error
}

type AuthService interface {
	Login(ctx context.Context, dto LoginDTO) (*User, string, error)
	AdminRevokeSessions(ctx context.Context, userID uuid.UUID) error
}
