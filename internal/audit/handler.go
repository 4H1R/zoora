package audit

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

// auditListConfig is the handler-owned white-list for GET /audit. Filtering is
// structured (typed query params below), not free-text search, and the log is
// always newest-first — so no search/order fields are exposed to the client.
var auditListConfig = domain.ListConfig{
	AllowedSearchFields: nil,
	AllowedOrderFields:  nil,
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
}

type Handler struct {
	svc domain.AuditService
}

func NewHandler(svc domain.AuditService) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts the audit log read endpoint. perm is auth.RequirePermission;
// the route-level permission check is defense in depth (the service also checks).
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, perm func(domain.PermissionName) gin.HandlerFunc) {
	authed := rg.Group("", authMiddleware)
	{
		authed.GET("/audit", perm(domain.PermAuditViewAny), h.list)
	}
}

// list returns the org's audit entries, newest first, with optional filters.
//
//	@Summary		List audit log entries
//	@Description	Returns the organization's audit log, newest first. Manager-only (audit:view_any). Filter by actor, action, target_type, target_id, outcome, and date range.
//	@Tags			Audit
//	@Security		BearerAuth
//	@Produce		json
//	@Param			page		query		int		false	"1-based page number"
//	@Param			page_size	query		int		false	"Items per page (default 20)"
//	@Param			actor_id	query		string	false	"Filter by actor user id (uuid)"
//	@Param			action		query		string	false	"Filter by action (created, updated, deleted, ...)"
//	@Param			target_type	query		string	false	"Filter by target type (class, user, role, ...)"
//	@Param			target_id	query		string	false	"Filter by target id (uuid) — per-resource history"
//	@Param			outcome		query		string	false	"Filter by outcome (success, denied)"
//	@Param			from		query		string	false	"Created at >= (RFC3339)"
//	@Param			to			query		string	false	"Created at <= (RFC3339)"
//	@Success		200	{object}	domain.Response{data=domain.PaginatedData{items=[]domain.AuditEntry}}
//	@Failure		401	{object}	domain.Response{error=domain.ErrorBody}
//	@Failure		403	{object}	domain.Response{error=domain.ErrorBody}
//	@Router			/audit [get]
func (h *Handler) list(c *gin.Context) {
	q := domain.AuditListQuery{ListParams: listparams.Bind(c, auditListConfig)}

	if v := c.Query("actor_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			q.ActorID = &id
		}
	}
	if v := c.Query("action"); v != "" {
		a := domain.AuditAction(v)
		if a.Valid() {
			q.Action = &a
		}
	}
	if v := c.Query("target_type"); v != "" {
		t := domain.AuditTargetType(v)
		if t.Valid() {
			q.TargetType = &t
		}
	}
	if v := c.Query("target_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			q.TargetID = &id
		}
	}
	if v := c.Query("outcome"); v != "" {
		o := domain.AuditOutcome(v)
		if o.Valid() {
			q.Outcome = &o
		}
	}
	if v := c.Query("from"); v != "" {
		if ts, err := time.Parse(time.RFC3339, v); err == nil {
			q.From = &ts
		}
	}
	if v := c.Query("to"); v != "" {
		if ts, err := time.Parse(time.RFC3339, v); err == nil {
			q.To = &ts
		}
	}

	entries, total, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}

	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(entries, total, q.ListParams))
}
