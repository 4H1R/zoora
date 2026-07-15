package livesessions

import (
	"context"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func (s *service) requireAdmin(ctx context.Context) (domain.Caller, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok || !caller.IsAdmin {
		return domain.Caller{}, domain.ErrForbidden
	}
	return caller, nil
}

func (s *service) AdminList(ctx context.Context, q domain.AdminListLiveRoomsQuery) ([]domain.LiveRoom, int64, error) {
	if _, err := s.requireAdmin(ctx); err != nil {
		return nil, 0, err
	}
	// Pagination defaults are applied by listparams.Bind in the handler.
	return s.rooms.AdminList(ctx, q)
}

func (s *service) AdminEndRoom(ctx context.Context, roomID uuid.UUID) (*domain.LiveRoom, error) {
	caller, err := s.requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	room, err := s.rooms.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room.Status != domain.LiveRoomStatusActive {
		return nil, domain.NewValidationError(map[string]string{"status": "room must be active to end"})
	}
	room, err = s.endRoomInternal(ctx, room)
	if err != nil {
		return nil, err
	}
	s.logger.Warn("admin ended live room",
		"room_id", room.ID.String(),
		"ended_by", caller.UserID.String(),
	)
	return room, nil
}

func (s *service) AdminHardDelete(ctx context.Context, roomID uuid.UUID) error {
	caller, err := s.requireAdmin(ctx)
	if err != nil {
		return err
	}
	if err := s.rooms.HardDelete(ctx, roomID); err != nil {
		return err
	}
	// Hard delete removes the room row for good, so its shared slides (media rows +
	// S3 objects) would otherwise be orphaned — the normal EndRoom teardown never
	// runs on this path. Slide media is polymorphic (no FK to the room), so no
	// cascade reaches it. Best-effort, mirrors the EndRoom cleanup.
	s.enqueueSlidesCleanup(ctx, roomID)
	s.logger.Warn("admin hard-deleted live room",
		"room_id", roomID.String(),
		"deleted_by", caller.UserID.String(),
	)
	return nil
}
