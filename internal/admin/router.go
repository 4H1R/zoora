// Package admin registers the /api/v1/admin route tree.
//
// Admin endpoints require caller.IsAdmin = true (enforced by RequireAdmin
// middleware at the group level). Per-feature admin handlers live alongside
// their feature packages and are wired into a single admin group via
// cmd/api/main.go by passing Registrar implementations to RegisterRoutes.
package admin

import (
	"github.com/gin-gonic/gin"
)

// Registrar is implemented by per-feature admin handlers. Each one attaches
// its sub-routes to the shared admin group.
type Registrar interface {
	RegisterAdminRoutes(group *gin.RouterGroup)
}

// RegisterRoutes mounts all admin sub-routers under the given group. The
// group is expected to already have authMiddleware + RequireAdmin applied.
func RegisterRoutes(group *gin.RouterGroup, registrars ...Registrar) {
	for _, r := range registrars {
		r.RegisterAdminRoutes(group)
	}
}
