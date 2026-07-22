package middleware

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/domain"
)

func discardLogger() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

// recorderSpy captures every RecordDenied call. Record is unused by the
// middleware but present to satisfy domain.AuditRecorder.
type recorderSpy struct {
	mu      sync.Mutex
	denied  []domain.AuditRecord
	success []domain.AuditRecord
}

func (s *recorderSpy) Record(_ context.Context, r domain.AuditRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.success = append(s.success, r)
	return nil
}

func (s *recorderSpy) RecordDenied(_ context.Context, r domain.AuditRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.denied = append(s.denied, r)
	return nil
}

func newRouterWithCaller(spy domain.AuditRecorder, orgID *uuid.UUID) *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(domain.WithCaller(c.Request.Context(),
			domain.Caller{UserID: uuid.New(), OrgID: orgID}))
		c.Next()
	})
	r.Use(AuditDenied(spy, discardLogger()))
	return r
}

// newRouterWithErrorHandler mounts the OUTER ErrorHandler and the inner
// AuditDenied in the exact order cmd/api uses them: ErrorHandler wraps the whole
// chain (registered first via router.Use), AuditDenied is on the v1 group
// (inner). This reproduces the production path where a service-layer
// domain.ErrForbidden is only an attached c.Error when AuditDenied runs.
func newRouterWithErrorHandler(spy domain.AuditRecorder, orgID *uuid.UUID) *gin.Engine {
	r := gin.New()
	r.Use(ErrorHandler(discardLogger()))
	r.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(domain.WithCaller(c.Request.Context(),
			domain.Caller{UserID: uuid.New(), OrgID: orgID}))
		c.Next()
	})
	r.Use(AuditDenied(spy, discardLogger()))
	return r
}

// TestAuditDeniedRecordsServiceForbidden reproduces the production path: a
// handler returns domain.ErrForbidden via c.Error and writes nothing inline
// (the classic service→handler pattern). The outer ErrorHandler maps it to 403
// only in its own post-Next phase, which runs AFTER AuditDenied. The middleware
// must still record exactly one denied attempt off the attached error.
func TestAuditDeniedRecordsServiceForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	spy := &recorderSpy{}
	orgID := uuid.New()

	r := newRouterWithErrorHandler(spy, &orgID)
	r.DELETE("/api/v1/classes/:id", func(c *gin.Context) {
		_ = c.Error(domain.ErrForbidden) // no inline write; ErrorHandler maps later
	})

	id := uuid.New()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/classes/"+id.String(), nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	require.Len(t, spy.denied, 1)
	require.Empty(t, spy.success)
	rec := spy.denied[0]
	require.Equal(t, domain.AuditDeleted, rec.Action)
	require.Equal(t, domain.AuditTargetClass, rec.TargetType)
	require.NotNil(t, rec.TargetID)
	require.Equal(t, id, *rec.TargetID)
}

// TestAuditDeniedNoDoubleRecordWhenBothSignals asserts that when BOTH the status
// is already 403 AND an ErrForbidden is attached to c.Errors, the middleware
// records the denial exactly once (not once per signal).
func TestAuditDeniedNoDoubleRecordWhenBothSignals(t *testing.T) {
	gin.SetMode(gin.TestMode)
	spy := &recorderSpy{}
	orgID := uuid.New()

	r := newRouterWithCaller(spy, &orgID)
	r.DELETE("/api/v1/classes/:id", func(c *gin.Context) {
		_ = c.Error(domain.ErrForbidden)             // attached error signal
		domain.ErrorResponse(c, domain.ErrForbidden) // inline 403 signal
		c.Abort()
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/classes/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	require.Len(t, spy.denied, 1)
}

func TestAuditDeniedRecordsOn403Mutation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	spy := &recorderSpy{}
	orgID := uuid.New()

	r := newRouterWithCaller(spy, &orgID)
	r.DELETE("/api/v1/classes/:id", func(c *gin.Context) {
		domain.ErrorResponse(c, domain.ErrForbidden) // writes 403
		c.Abort()
	})

	id := uuid.New()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/classes/"+id.String(), nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	require.Len(t, spy.denied, 1)
	require.Empty(t, spy.success)
	rec := spy.denied[0]
	require.Equal(t, domain.AuditDeleted, rec.Action)
	require.Equal(t, domain.AuditTargetClass, rec.TargetType)
	require.NotNil(t, rec.TargetID)
	require.Equal(t, id, *rec.TargetID)
}

func TestAuditDeniedIgnoresSuccessfulMutation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	spy := &recorderSpy{}
	orgID := uuid.New()

	r := newRouterWithCaller(spy, &orgID)
	r.DELETE("/api/v1/classes/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/classes/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	require.Empty(t, spy.denied)
}

func TestAuditDeniedIgnoresReadDenial(t *testing.T) {
	gin.SetMode(gin.TestMode)
	spy := &recorderSpy{}
	orgID := uuid.New()

	r := newRouterWithCaller(spy, &orgID)
	r.GET("/api/v1/classes/:id", func(c *gin.Context) {
		domain.ErrorResponse(c, domain.ErrForbidden)
		c.Abort()
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/classes/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	require.Empty(t, spy.denied)
}

func TestAuditDeniedSkippedWithoutOrg(t *testing.T) {
	gin.SetMode(gin.TestMode)
	spy := &recorderSpy{}

	r := newRouterWithCaller(spy, nil) // no org on the caller
	r.DELETE("/api/v1/classes/:id", func(c *gin.Context) {
		domain.ErrorResponse(c, domain.ErrForbidden)
		c.Abort()
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/classes/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	require.Empty(t, spy.denied)
}

func TestRouteSegmentToTargetType(t *testing.T) {
	cases := map[string]domain.AuditTargetType{
		"/api/v1/classes/:id":                  domain.AuditTargetClass,
		"/api/v1/users/:id":                    domain.AuditTargetUser,
		"/api/v1/roles/:id":                    domain.AuditTargetRole,
		"/api/v1/quizzes/:id":                  domain.AuditTargetQuiz,
		"/api/v1/question-banks/:id/questions": domain.AuditTargetQuestionBank,
		"/api/v1/calendar/events":              domain.AuditTargetCalendarEvent,
		"/api/v1/custom-field-definitions":     domain.AuditTargetCustomField,
		"/api/v1/frobnicate/:id":               domain.AuditTargetType("frobnicate"),
	}
	for path, want := range cases {
		require.Equal(t, want, routeSegmentToTargetType(path), path)
	}
}
