package auth

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/cache"
)

const (
	ContextKeyUserID  = "user_id"
	ContextKeyIsAdmin = "is_admin"
)

// Middleware validates the JWT, loads user permissions from cache/DB, and
// injects a fully-populated Caller into the request context.
func Middleware(jwt *JWTService, rdb *redis.Client, roleRepo domain.RoleRepository, userRepo domain.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			domain.ErrorResponse(c, domain.ErrUnauthorized)
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			domain.ErrorResponse(c, domain.ErrUnauthorized)
			c.Abort()
			return
		}

		claims, err := jwt.ValidateToken(parts[1])
		if err != nil {
			domain.ErrorResponse(c, domain.ErrUnauthorized)
			c.Abort()
			return
		}

		if rdb != nil && isRevoked(c, rdb, claims) {
			domain.ErrorResponse(c, domain.ErrUnauthorized)
			c.Abort()
			return
		}

		caller := domain.Caller{UserID: claims.UserID}

		if userRepo != nil {
			user, err := userRepo.FindByID(c.Request.Context(), claims.UserID)
			if err != nil {
				domain.ErrorResponse(c, domain.ErrUnauthorized)
				c.Abort()
				return
			}
			caller.OrgID = user.OrganizationID
			caller.IsAdmin = user.IsAdmin
			caller.RoleID = user.RoleID
			caller.Username = user.Username
			caller.Name = user.Name

			if !user.IsAdmin && user.RoleID != nil {
				caller.Permissions = loadPermissions(c, rdb, roleRepo, *user.RoleID)
			}
		}

		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyIsAdmin, caller.IsAdmin)

		c.Request = c.Request.WithContext(domain.WithCaller(c.Request.Context(), caller))
		c.Next()
	}
}

func loadPermissions(c *gin.Context, rdb *redis.Client, roleRepo domain.RoleRepository, roleID uuid.UUID) []string {
	ctx := c.Request.Context()

	if rdb != nil {
		perms, err := cache.GetRolePermissions(ctx, rdb, roleID)
		if err == nil {
			return perms
		}
	}

	perms, err := roleRepo.GetPermissionNames(ctx, roleID)
	if err != nil {
		return nil
	}

	if rdb != nil {
		_ = cache.SetRolePermissions(ctx, rdb, roleID, perms)
	}
	return perms
}

func isRevoked(c *gin.Context, rdb *redis.Client, claims *Claims) bool {
	val, err := rdb.Get(c.Request.Context(), RevokedKey(claims.UserID.String())).Result()
	if err != nil {
		return !errors.Is(err, redis.Nil)
	}
	revokedAt, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return true
	}
	if claims.IssuedAt == nil {
		return true
	}
	return claims.IssuedAt.Unix() < revokedAt
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !GetIsAdmin(c) {
			domain.ErrorResponse(c, domain.ErrForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequirePermission checks that the caller is admin OR has the given permission.
func RequirePermission(permission domain.PermissionName) gin.HandlerFunc {
	return func(c *gin.Context) {
		caller, ok := domain.CallerFromCtx(c.Request.Context())
		if !ok {
			domain.ErrorResponse(c, domain.ErrUnauthorized)
			c.Abort()
			return
		}

		if caller.IsAdmin {
			c.Next()
			return
		}

		if caller.HasPermission(permission) {
			c.Next()
			return
		}

		domain.ErrorResponse(c, domain.ErrForbidden)
		c.Abort()
	}
}

// RequireAnyPermission checks that the caller is admin OR has at least one of the given permissions.
func RequireAnyPermission(permissions ...domain.PermissionName) gin.HandlerFunc {
	return func(c *gin.Context) {
		caller, ok := domain.CallerFromCtx(c.Request.Context())
		if !ok {
			domain.ErrorResponse(c, domain.ErrUnauthorized)
			c.Abort()
			return
		}

		if caller.IsAdmin {
			c.Next()
			return
		}

		for _, p := range permissions {
			if caller.HasPermission(p) {
				c.Next()
				return
			}
		}

		domain.ErrorResponse(c, domain.ErrForbidden)
		c.Abort()
	}
}

// RequireSelfOrPermission checks if the caller is accessing their own resource
// (matched by the URL param) OR has the `_any` permission. For self-access the
// base permission (without _any) is required; for accessing another user's
// resource the anyPerm is required.
//
// Example: RequireSelfOrPermission("users:view", "users:view_any", "id")
//   - GET /users/123 where caller.UserID == 123 → needs "users:view"
//   - GET /users/456 where caller.UserID != 456 → needs "users:view_any"
func RequireSelfOrPermission(selfPerm, anyPerm domain.PermissionName, ownerIDParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		caller, ok := domain.CallerFromCtx(c.Request.Context())
		if !ok {
			domain.ErrorResponse(c, domain.ErrUnauthorized)
			c.Abort()
			return
		}

		if caller.IsAdmin {
			c.Next()
			return
		}

		resourceOwnerID, err := uuid.Parse(c.Param(ownerIDParam))
		if err != nil {
			domain.ErrorResponse(c, domain.ErrForbidden)
			c.Abort()
			return
		}

		if caller.UserID == resourceOwnerID {
			if selfPerm == "" || caller.HasPermission(selfPerm) {
				c.Next()
				return
			}
		}

		if caller.HasPermission(anyPerm) {
			c.Next()
			return
		}

		domain.ErrorResponse(c, domain.ErrForbidden)
		c.Abort()
	}
}

func GetUserID(c *gin.Context) uuid.UUID {
	id, _ := c.Get(ContextKeyUserID)
	return id.(uuid.UUID)
}

func GetIsAdmin(c *gin.Context) bool {
	val, exists := c.Get(ContextKeyIsAdmin)
	if !exists {
		return false
	}
	return val.(bool)
}
