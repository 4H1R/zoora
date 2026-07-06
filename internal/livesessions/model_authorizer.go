package livesessions

import (
	"context"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// modelAuthorizer resolves polymorphic authz for models owned by this feature.
// Today it handles domain.QAModelLiveSession ("live_session"), where model_id
// is a LiveRoom ID. It reuses the same ownership rules as the live-session
// service: manage permission on the owning class => moderator; enrolled member
// or viewer permission => participant.
type modelAuthorizer struct {
	rooms    domain.LiveRoomRepository
	sessions domain.ClassSessionRepository
	classes  domain.ClassRepository
	members  domain.ClassMemberRepository
}

// NewModelAuthorizer builds a ModelAuthorizer backed by live-session ownership.
func NewModelAuthorizer(
	rooms domain.LiveRoomRepository,
	sessions domain.ClassSessionRepository,
	classes domain.ClassRepository,
	members domain.ClassMemberRepository,
) domain.ModelAuthorizer {
	return &modelAuthorizer{rooms: rooms, sessions: sessions, classes: classes, members: members}
}

// classForModel resolves model_type + model_id to the owning class. Only
// live_session is supported; anything else yields ErrUnsupportedModelType.
func (a *modelAuthorizer) classForModel(ctx context.Context, modelType string, modelID uuid.UUID) (*domain.Class, error) {
	if modelType != domain.QAModelLiveSession {
		return nil, domain.ErrUnsupportedModelType
	}
	room, err := a.rooms.FindByID(ctx, modelID)
	if err != nil {
		return nil, err
	}
	session, err := a.sessions.FindByID(ctx, room.ClassSessionID)
	if err != nil {
		return nil, err
	}
	class, err := a.classes.FindByID(ctx, session.ClassID)
	if err != nil {
		return nil, err
	}
	return class, nil
}

func (a *modelAuthorizer) CanModerate(ctx context.Context, caller domain.Caller, modelType string, modelID uuid.UUID) (bool, error) {
	class, err := a.classForModel(ctx, modelType, modelID)
	if err != nil {
		return false, err
	}
	return caller.CanManageOwned(class.UserID, domain.PermLiveSessionsManage, domain.PermLiveSessionsManageAny), nil
}

func (a *modelAuthorizer) CanParticipate(ctx context.Context, caller domain.Caller, modelType string, modelID uuid.UUID) (bool, error) {
	class, err := a.classForModel(ctx, modelType, modelID)
	if err != nil {
		return false, err
	}
	if caller.IsAdmin || caller.HasPermission(domain.PermLiveSessionsViewAny) {
		return true, nil
	}
	if caller.CanManageOwned(class.UserID, domain.PermLiveSessionsManage, domain.PermLiveSessionsManageAny) {
		return true, nil
	}
	return a.members.Exists(ctx, class.ID, caller.UserID)
}
