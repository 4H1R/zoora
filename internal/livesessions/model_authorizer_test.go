package livesessions_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/livesessions"
)

func TestModelAuthorizer_UnsupportedModelType(t *testing.T) {
	auth := livesessions.NewModelAuthorizer(nil, nil, nil, nil)

	_, err := auth.CanModerate(context.Background(), domain.Caller{}, "chat", uuid.New())
	assert.ErrorIs(t, err, domain.ErrUnsupportedModelType)

	_, err = auth.CanParticipate(context.Background(), domain.Caller{}, "chat", uuid.New())
	assert.ErrorIs(t, err, domain.ErrUnsupportedModelType)
}

func TestModelAuthorizer_ModeratorIsClassOwnerWithManage(t *testing.T) {
	roomID := uuid.New()
	sessionID := uuid.New()
	classID := uuid.New()
	ownerID := uuid.New()

	rooms := &mockRoomRepo{}
	rooms.On("FindByID", mock.Anything, roomID).
		Return(&domain.LiveRoom{ID: roomID, ClassSessionID: sessionID}, nil)
	sessions := &mockClassSessionRepo{}
	sessions.On("FindByID", mock.Anything, sessionID).
		Return(&domain.ClassSession{ID: sessionID, ClassID: classID}, nil)
	classes := &mockClassRepo{}
	classes.On("FindByID", mock.Anything, classID).
		Return(&domain.Class{ID: classID, UserID: ownerID}, nil)
	members := &mockMemberRepo{}

	auth := livesessions.NewModelAuthorizer(rooms, sessions, classes, members)

	owner := domain.Caller{UserID: ownerID, Permissions: []string{string(domain.PermLiveSessionsManage)}}
	ok, err := auth.CanModerate(context.Background(), owner, domain.QAModelLiveSession, roomID)
	require.NoError(t, err)
	assert.True(t, ok)

	// A non-owner without manage:any is not a moderator.
	stranger := domain.Caller{UserID: uuid.New(), Permissions: []string{string(domain.PermLiveSessionsManage)}}
	ok, err = auth.CanModerate(context.Background(), stranger, domain.QAModelLiveSession, roomID)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestModelAuthorizer_ParticipantIsEnrolledMember(t *testing.T) {
	roomID := uuid.New()
	sessionID := uuid.New()
	classID := uuid.New()
	ownerID := uuid.New()
	studentID := uuid.New()

	rooms := &mockRoomRepo{}
	rooms.On("FindByID", mock.Anything, roomID).
		Return(&domain.LiveRoom{ID: roomID, ClassSessionID: sessionID}, nil)
	sessions := &mockClassSessionRepo{}
	sessions.On("FindByID", mock.Anything, sessionID).
		Return(&domain.ClassSession{ID: sessionID, ClassID: classID}, nil)
	classes := &mockClassRepo{}
	classes.On("FindByID", mock.Anything, classID).
		Return(&domain.Class{ID: classID, UserID: ownerID}, nil)
	members := &mockMemberRepo{}
	members.On("Exists", mock.Anything, classID, studentID).Return(true, nil)

	auth := livesessions.NewModelAuthorizer(rooms, sessions, classes, members)

	student := domain.Caller{UserID: studentID}
	ok, err := auth.CanParticipate(context.Background(), student, domain.QAModelLiveSession, roomID)
	require.NoError(t, err)
	assert.True(t, ok)
}
