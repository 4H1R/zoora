package polls

import (
	"context"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// modelTypeClass is the polymorphic owner type for polls attached directly to a
// class. For this type the poll's model_id IS the class ID.
const modelTypeClass = "class"

// modelAuthorizer resolves polymorphic authz for polls. Polls attach to either a
// class (model_id is the class ID) or a live_session (model_id is a LiveRoom ID,
// resolved room -> session -> class). Authorization is then evaluated against the
// owning class using poll permissions: manage on the class => moderator; enrolled
// member => participant.
type modelAuthorizer struct {
	rooms    domain.LiveRoomRepository
	sessions domain.ClassSessionRepository
	classes  domain.ClassRepository
	members  domain.ClassMemberRepository
}

// NewModelAuthorizer builds a poll ModelAuthorizer backed by class ownership.
func NewModelAuthorizer(
	rooms domain.LiveRoomRepository,
	sessions domain.ClassSessionRepository,
	classes domain.ClassRepository,
	members domain.ClassMemberRepository,
) domain.ModelAuthorizer {
	return &modelAuthorizer{rooms: rooms, sessions: sessions, classes: classes, members: members}
}

// classForModel resolves model_type + model_id to the owning class. Supported
// types: "class" (model_id is the class ID) and "live_session" (model_id is a
// LiveRoom ID). Anything else yields ErrUnsupportedModelType.
func (a *modelAuthorizer) classForModel(ctx context.Context, modelType string, modelID uuid.UUID) (*domain.Class, error) {
	switch modelType {
	case modelTypeClass:
		return a.classes.FindByID(ctx, modelID)
	case domain.ChatModelLiveSession:
		room, err := a.rooms.FindByID(ctx, modelID)
		if err != nil {
			return nil, err
		}
		session, err := a.sessions.FindByID(ctx, room.ClassSessionID)
		if err != nil {
			return nil, err
		}
		return a.classes.FindByID(ctx, session.ClassID)
	default:
		return nil, domain.ErrUnsupportedModelType
	}
}

// CanModerate reports whether the caller may perform host/teacher actions on the
// poll: admins, polls:update_any holders, or the owning teacher of the class.
func (a *modelAuthorizer) CanModerate(ctx context.Context, caller domain.Caller, modelType string, modelID uuid.UUID) (bool, error) {
	class, err := a.classForModel(ctx, modelType, modelID)
	if err != nil {
		return false, err
	}
	return caller.CanManage(class.UserID, domain.PermPollsUpdateAny), nil
}

// CanParticipate reports whether the caller may read/vote on the poll: any
// moderator, or an enrolled member of the owning class.
func (a *modelAuthorizer) CanParticipate(ctx context.Context, caller domain.Caller, modelType string, modelID uuid.UUID) (bool, error) {
	class, err := a.classForModel(ctx, modelType, modelID)
	if err != nil {
		return false, err
	}
	if caller.CanManage(class.UserID, domain.PermPollsUpdateAny) {
		return true, nil
	}
	return a.members.Exists(ctx, class.ID, caller.UserID)
}
