package admin

import (
	"testing"

	"github.com/gin-gonic/gin"
)

type registrarFunc func(group *gin.RouterGroup)

func (f registrarFunc) RegisterAdminRoutes(group *gin.RouterGroup) {
	f(group)
}

func TestRegisterRoutesInvokesRegistrarsInOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/admin")
	calls := []string{}

	RegisterRoutes(group,
		registrarFunc(func(*gin.RouterGroup) { calls = append(calls, "first") }),
		registrarFunc(func(*gin.RouterGroup) { calls = append(calls, "second") }),
	)

	if len(calls) != 2 || calls[0] != "first" || calls[1] != "second" {
		t.Fatalf("registrar calls = %#v, want first then second", calls)
	}
}
