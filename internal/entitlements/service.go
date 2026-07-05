package entitlements

import (
	"context"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// Service enforces numeric plan limits that require a live count. Boolean
// feature gates do NOT use this — they read Caller.HasFeature directly.
type Service interface {
	CheckUserLimit(ctx context.Context, orgID uuid.UUID, ent domain.Entitlements) error
	CheckStorageLimit(ctx context.Context, orgID uuid.UUID, ent domain.Entitlements, addBytes int64) error
	CheckConcurrentRoomsLimit(ctx context.Context, orgID uuid.UUID, ent domain.Entitlements) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CheckUserLimit(ctx context.Context, orgID uuid.UUID, ent domain.Entitlements) error {
	if ent.Unlimited(domain.LimitMaxUsers) {
		return nil
	}
	n, err := s.repo.CountUsers(ctx, orgID)
	if err != nil {
		return err
	}
	if !ent.Within(domain.LimitMaxUsers, n) {
		return domain.NewLimitError(ent.Plan, domain.LimitMaxUsers, n, ent.Limit(domain.LimitMaxUsers))
	}
	return nil
}

func (s *service) CheckStorageLimit(ctx context.Context, orgID uuid.UUID, ent domain.Entitlements, addBytes int64) error {
	if ent.Unlimited(domain.LimitStorageGB) {
		return nil
	}
	used, err := s.repo.SumStorageBytes(ctx, orgID)
	if err != nil {
		return err
	}
	ceilingBytes := ent.Limit(domain.LimitStorageGB) * 1024 * 1024 * 1024
	if used+addBytes > ceilingBytes {
		return domain.NewLimitError(ent.Plan, domain.LimitStorageGB, used+addBytes, ceilingBytes)
	}
	return nil
}

func (s *service) CheckConcurrentRoomsLimit(ctx context.Context, orgID uuid.UUID, ent domain.Entitlements) error {
	if ent.Unlimited(domain.LimitConcurrentRooms) {
		return nil
	}
	n, err := s.repo.CountActiveLiveRooms(ctx, orgID)
	if err != nil {
		return err
	}
	if !ent.Within(domain.LimitConcurrentRooms, n) {
		return domain.NewLimitError(ent.Plan, domain.LimitConcurrentRooms, n, ent.Limit(domain.LimitConcurrentRooms))
	}
	return nil
}
