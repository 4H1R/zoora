package audit

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

type service struct {
	repo   domain.AuditRepository
	logger *slog.Logger
}

func NewService(repo domain.AuditRepository, logger *slog.Logger) domain.AuditService {
	return &service{repo: repo, logger: logger}
}

// Record builds and inserts a success entry, deriving actor/org from the Caller
// and IP/UA from RequestInfo. Called inside the caller's tx (via ctx), so it
// commits or rolls back with the change. Returns an error on failure (hard-fail
// at the call site).
func (s *service) Record(ctx context.Context, r domain.AuditRecord) error {
	entry := &domain.AuditEntry{
		Action:      r.Action,
		TargetType:  r.TargetType,
		TargetID:    r.TargetID,
		TargetLabel: r.TargetLabel,
		Outcome:     domain.AuditOutcomeSuccess,
		Metadata:    map[string]any{},
	}
	for k, v := range r.Metadata {
		entry.Metadata[k] = v
	}

	// Actor + org from the Caller when present; System fallback otherwise.
	var orgFromCaller *uuid.UUID
	if caller, ok := domain.CallerFromCtx(ctx); ok {
		actorID := caller.UserID
		entry.ActorID = &actorID
		entry.ActorName = caller.Name
		entry.ActorUsername = caller.Username
		orgFromCaller = caller.OrgID
		if caller.IsAdmin {
			entry.Metadata["platform_admin"] = true
		}
	} else {
		entry.ActorName = domain.AuditActorSystemName
	}

	// org_id = target's org: explicit OrgID override wins (Platform Admin /
	// System / worker), else the caller's org.
	switch {
	case r.OrgID != nil:
		entry.OrganizationID = *r.OrgID
	case orgFromCaller != nil:
		entry.OrganizationID = *orgFromCaller
	default:
		// No org resolvable: refuse rather than file an orphan entry.
		return domain.ErrValidation
	}

	if ri, ok := domain.RequestInfoFromCtx(ctx); ok {
		if ri.IP != "" {
			entry.Metadata["ip"] = ri.IP
		}
		if ri.UserAgent != "" {
			entry.Metadata["user_agent"] = ri.UserAgent
		}
	}

	return s.repo.Create(ctx, entry)
}

func (s *service) List(ctx context.Context, q domain.AuditListQuery) ([]domain.AuditEntry, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrUnauthorized
	}
	if !caller.HasAny(domain.PermAuditViewAny) {
		return nil, 0, domain.ErrForbidden
	}
	if caller.OrgID == nil {
		// Platform Admin has no single org log; this endpoint is org-scoped.
		return nil, 0, domain.ErrForbidden
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = domain.DefaultPageSize
	}
	return s.repo.List(ctx, *caller.OrgID, q)
}
