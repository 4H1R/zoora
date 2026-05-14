package attendance

import (
	"context"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// requireAdmin pulls the caller from context and rejects anything that is not
// a verified super-admin. Defense-in-depth — the admin route group already
// mounts RequireAdmin middleware.
func (s *service) requireAdmin(ctx context.Context) (domain.Caller, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok || !caller.IsAdmin {
		return domain.Caller{}, domain.ErrForbidden
	}
	return caller, nil
}

func (s *service) AdminList(ctx context.Context, q domain.AdminListAttendanceQuery) ([]domain.Attendance, int64, error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, 0, err
	}
	if q.ListParams.Page < 1 {
		q.ListParams.Page = 1
	}
	if q.ListParams.PageSize <= 0 {
		q.ListParams.PageSize = domain.DefaultPageSize
	}
	return s.repo.AdminList(ctx, q)
}

func (s *service) AdminUpdate(ctx context.Context, id uuid.UUID, dto domain.AdminUpdateAttendanceDTO) (*domain.Attendance, error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, err
	}
	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if dto.Status != nil {
		a.Status = *dto.Status
	}
	if dto.Remarks != nil {
		a.Remarks = *dto.Remarks
	}
	if err := s.repo.Update(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *service) AdminHardDelete(ctx context.Context, id uuid.UUID) error {
	caller, err := s.requireAdmin(ctx)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.logger.Warn("admin hard-deleted attendance",
		"attendance_id", id.String(),
		"deleted_by", caller.UserID.String(),
	)
	return nil
}
