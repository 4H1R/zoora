package livesessions_test

// Regression tests for the live-session lifecycle fixes: rejoin handling,
// race-safe status transitions, sweep behavior when LiveKit already dropped
// the room, recording double-start, config normalization, and webhook-driven
// participant/egress finalization.

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	lkproto "github.com/livekit/protocol/livekit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/4H1R/zoora/internal/domain"
	lk "github.com/4H1R/zoora/internal/platform/livekit"
)

func activeTestRoom() *domain.LiveRoom {
	room := testRoom()
	room.Status = domain.LiveRoomStatusActive
	return room
}

func stubRoomLookups(f *lkFixture, room *domain.LiveRoom) {
	f.rooms.On("FindByID", mock.Anything, testRoomID).Return(room, nil)
	f.sess.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	f.classes.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)
}

// --- JoinRoom -----------------------------------------------------------

func TestJoinRoom_FirstJoin_CreatesViewerParticipant(t *testing.T) {
	svc, f := newTestServiceLK(t)
	stubRoomLookups(f, activeTestRoom())
	f.members.On("Exists", mock.Anything, testClassID, testStudentID).Return(true, nil)
	f.parts.On("FindActiveByRoomAndUser", mock.Anything, testRoomID, testStudentID).
		Return(nil, domain.ErrNotFound)
	f.parts.On("Create", mock.Anything, mock.MatchedBy(func(p *domain.LiveParticipant) bool {
		return p.Role == domain.ParticipantRoleViewer && p.UserID == testStudentID
	})).Return(nil)
	f.chat.On("FindChatByRoom", mock.Anything, testRoomID).
		Return(nil, domain.ErrNotFound)

	resp, err := svc.JoinRoom(studentCtx(), testRoomID)
	assert.NoError(t, err)
	assert.Equal(t, "test-token", resp.Token)
	f.parts.AssertExpectations(t)
}

func TestJoinRoom_Rejoin_ReusesActiveParticipantRow(t *testing.T) {
	svc, f := newTestServiceLK(t)
	stubRoomLookups(f, activeTestRoom())
	f.members.On("Exists", mock.Anything, testClassID, testStudentID).Return(true, nil)
	existing := &domain.LiveParticipant{
		ID:         uuid.New(),
		LiveRoomID: testRoomID,
		UserID:     testStudentID,
		Identity:   testStudentID.String(),
		Role:       domain.ParticipantRoleViewer,
		JoinedAt:   time.Now().Add(-10 * time.Minute),
	}
	f.parts.On("FindActiveByRoomAndUser", mock.Anything, testRoomID, testStudentID).
		Return(existing, nil)
	f.chat.On("FindChatByRoom", mock.Anything, testRoomID).
		Return(nil, domain.ErrNotFound)

	_, err := svc.JoinRoom(studentCtx(), testRoomID)
	assert.NoError(t, err)
	f.parts.AssertNotCalled(t, "Create")
}

func TestJoinRoom_Rejoin_PreservesPresenterRole(t *testing.T) {
	svc, f := newTestServiceLK(t)
	stubRoomLookups(f, activeTestRoom())
	f.members.On("Exists", mock.Anything, testClassID, testStudentID).Return(true, nil)
	existing := &domain.LiveParticipant{
		ID:         uuid.New(),
		LiveRoomID: testRoomID,
		UserID:     testStudentID,
		Identity:   testStudentID.String(),
		Role:       domain.ParticipantRolePresenter,
		JoinedAt:   time.Now().Add(-10 * time.Minute),
	}
	f.parts.On("FindActiveByRoomAndUser", mock.Anything, testRoomID, testStudentID).
		Return(existing, nil)
	f.chat.On("FindChatByRoom", mock.Anything, testRoomID).
		Return(nil, domain.ErrNotFound)

	_, err := svc.JoinRoom(studentCtx(), testRoomID)
	assert.NoError(t, err)

	// The reissued token must carry the promoted role: publish sources granted
	// and presenter metadata, but no room-admin.
	if assert.Len(t, f.lk.tokenSources, 1) {
		assert.NotEmpty(t, f.lk.tokenSources[0], "presenter must keep publish rights on rejoin")
		assert.Contains(t, f.lk.tokenMetadata[0], string(domain.ParticipantRolePresenter))
		assert.False(t, f.lk.tokenRoomAdmin[0])
	}
}

func TestJoinRoom_FinishedRoom_ValidationError(t *testing.T) {
	svc, f := newTestServiceLK(t)
	room := testRoom()
	room.Status = domain.LiveRoomStatusFinished
	stubRoomLookups(f, room)

	_, err := svc.JoinRoom(studentCtx(), testRoomID)
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestJoinRoom_HostAutoStart_LostRace_ContinuesWhenActive(t *testing.T) {
	svc, f := newTestServiceLK(t)
	created := testRoom() // status created
	f.rooms.On("FindByID", mock.Anything, testRoomID).Return(created, nil).Once()
	f.sess.On("FindByID", mock.Anything, testSessionID).Return(testSession(), nil)
	f.classes.On("FindByID", mock.Anything, testClassID).Return(testClass(), nil)
	// Concurrent moderator won the created→active transition.
	f.rooms.On("Transition", mock.Anything, mock.AnythingOfType("*domain.LiveRoom"), domain.LiveRoomStatusCreated).
		Return(domain.ErrConflict)
	f.rooms.On("FindByID", mock.Anything, testRoomID).Return(activeTestRoom(), nil).Once()
	f.parts.On("FindActiveByRoomAndUser", mock.Anything, testRoomID, testTeacherID).
		Return(nil, domain.ErrNotFound)
	f.parts.On("Create", mock.Anything, mock.AnythingOfType("*domain.LiveParticipant")).Return(nil)
	f.chat.On("FindChatByRoom", mock.Anything, testRoomID).
		Return(nil, domain.ErrNotFound)

	resp, err := svc.JoinRoom(teacherCtx(), testRoomID)
	assert.NoError(t, err)
	assert.Equal(t, "test-token", resp.Token)
}

// --- EndRoom / transitions ------------------------------------------------

func TestEndRoom_LostCloseRace_Conflict(t *testing.T) {
	svc, f := newTestServiceLK(t)
	stubRoomLookups(f, activeTestRoom())
	f.rooms.On("Transition", mock.Anything, mock.AnythingOfType("*domain.LiveRoom"), domain.LiveRoomStatusActive).
		Return(domain.ErrConflict)

	_, err := svc.EndRoom(teacherCtx(), testRoomID)
	assert.ErrorIs(t, err, domain.ErrConflict)
	// Losing the race must not run teardown side effects again.
	f.parts.AssertNotCalled(t, "MarkAllLeft")
	f.chat.AssertNotCalled(t, "ArchiveByRoom")
}

// --- Sweep / no-host close -------------------------------------------------

func TestAutoClose_LiveKitRoomGone_StillClosesRoom(t *testing.T) {
	svc, f := newTestServiceLK(t)
	room := *activeTestRoom()
	f.rooms.On("FindActiveRoomsWithStaleHost", mock.Anything, 15*time.Minute).
		Return([]domain.LiveRoom{room}, nil)
	// LiveKit already dropped the room (missed webhook case): the sweep must
	// treat that as "no host" and finalize, not error forever.
	f.lk.listParticipantsFn = func(context.Context, string) ([]*lkproto.ParticipantInfo, error) {
		return nil, lk.ErrRoomNotFound
	}
	f.recs.On("FindActiveByRoom", mock.Anything, testRoomID).Return(nil, domain.ErrNotFound)
	f.rooms.On("Transition", mock.Anything, mock.AnythingOfType("*domain.LiveRoom"), domain.LiveRoomStatusActive).
		Return(nil)
	f.parts.On("MarkAllLeft", mock.Anything, testRoomID, mock.AnythingOfType("time.Time")).Return(nil)
	f.wb.On("Delete", mock.Anything, testRoomID).Return(nil)
	f.chat.On("ArchiveByRoom", mock.Anything, testRoomID).Return(nil)

	err := svc.AutoCloseStaleRooms(context.Background())
	assert.NoError(t, err)
	f.rooms.AssertExpectations(t)
	// Room-finish teardown must close the room's polls (default Maybe stub in the
	// fixture handles the return; assert it actually fired).
	f.poll.AssertCalled(t, "CloseByModel", mock.Anything, domain.ChatModelLiveSession, testRoomID)
}

func TestAutoClose_HostStillConnected_RefreshesLastSeen(t *testing.T) {
	svc, f := newTestServiceLK(t)
	room := *activeTestRoom()
	f.rooms.On("FindActiveRoomsWithStaleHost", mock.Anything, 15*time.Minute).
		Return([]domain.LiveRoom{room}, nil)
	meta, _ := json.Marshal(map[string]string{"role": string(domain.ParticipantRoleHost)})
	f.lk.listParticipantsFn = func(context.Context, string) ([]*lkproto.ParticipantInfo, error) {
		return []*lkproto.ParticipantInfo{{Identity: "host", Metadata: string(meta)}}, nil
	}
	f.rooms.On("TouchHostLastSeen", mock.Anything, testRoomID, mock.AnythingOfType("time.Time")).Return(nil)

	err := svc.AutoCloseStaleRooms(context.Background())
	assert.NoError(t, err)
	f.rooms.AssertNotCalled(t, "Transition")
	f.rooms.AssertExpectations(t)
}

// --- Recording ---------------------------------------------------------------

func TestStartRecording_ActiveRecordingExists_Conflict(t *testing.T) {
	svc, f := newTestServiceLK(t)
	stubRoomLookups(f, activeTestRoom())
	f.recs.On("FindActiveByRoom", mock.Anything, testRoomID).
		Return(&domain.LiveRecording{ID: uuid.New(), Status: domain.LiveRecordingStatusStarted}, nil)

	_, err := svc.StartRecording(teacherCtx(), testRoomID)
	assert.ErrorIs(t, err, domain.ErrConflict)
	f.recs.AssertNotCalled(t, "Create")
}

func TestStartRecording_FreePlanRejected(t *testing.T) {
	svc, f := newTestServiceLK(t)
	stubRoomLookups(f, activeTestRoom())

	_, err := svc.StartRecording(freeTeacherCtx(), testRoomID)
	assert.ErrorIs(t, err, domain.ErrFeatureNotInPlan)
	f.recs.AssertNotCalled(t, "Create")
}

func TestStartRoom_ConcurrentRoomLimitReached(t *testing.T) {
	svc, f := newTestServiceLKEnt(t, fakeEntSvc{
		concurrentErr: domain.NewLimitError(domain.PlanFree, domain.LimitConcurrentRooms, 1, 1),
	})
	stubRoomLookups(f, testRoom()) // created state

	_, err := svc.StartRoom(teacherCtx(), testRoomID)
	assert.ErrorIs(t, err, domain.ErrPlanLimitReached)
	// The room must not have been promoted to active.
	f.rooms.AssertNotCalled(t, "Transition", mock.Anything, mock.Anything, mock.Anything)
}

func TestUpdateRoomConfig_FreePlanClampsMaxParticipants(t *testing.T) {
	svc, f := newTestServiceLK(t)
	stubRoomLookups(f, activeTestRoom())
	f.rooms.On("UpdateConfig", mock.Anything, testRoomID, mock.MatchedBy(func(cfg domain.LiveRoomConfig) bool {
		return cfg.MaxParticipants == 5 // free_50 ceiling
	})).Return(nil)

	cfg := domain.LiveRoomConfig{MaxParticipants: 50} // above the free_50 ceiling
	room, err := svc.UpdateRoomConfig(freeTeacherCtx(), testRoomID, domain.UpdateLiveRoomConfigDTO{Config: &cfg})
	assert.NoError(t, err)
	assert.Equal(t, 5, room.Config.MaxParticipants)
	f.rooms.AssertExpectations(t)
}

// --- Config ------------------------------------------------------------------

func TestUpdateRoomConfig_NonPositiveMax_BackfillsDefault(t *testing.T) {
	svc, f := newTestServiceLK(t)
	stubRoomLookups(f, activeTestRoom())
	f.rooms.On("UpdateConfig", mock.Anything, testRoomID, mock.MatchedBy(func(cfg domain.LiveRoomConfig) bool {
		return cfg.MaxParticipants == domain.DefaultLiveRoomConfig().MaxParticipants
	})).Return(nil)

	cfg := domain.LiveRoomConfig{MaxParticipants: -5}
	room, err := svc.UpdateRoomConfig(teacherCtx(), testRoomID, domain.UpdateLiveRoomConfigDTO{Config: &cfg})
	assert.NoError(t, err)
	assert.Equal(t, domain.DefaultLiveRoomConfig().MaxParticipants, room.Config.MaxParticipants)
	f.rooms.AssertExpectations(t)
}

// --- Webhook-driven participant + egress finalization -------------------------

func TestOnLiveKitEvent_ParticipantLeft_MarksRowLeft(t *testing.T) {
	svc, f := newTestServiceLK(t)
	room := activeTestRoom()
	f.rooms.On("FindByLiveKitRoomName", mock.Anything, room.LiveKitRoomName).Return(room, nil)
	f.parts.On("MarkLeftByIdentity", mock.Anything, testRoomID, "user-1", mock.AnythingOfType("time.Time")).
		Return(nil)
	// Host still present → no close armed.
	meta, _ := json.Marshal(map[string]string{"role": string(domain.ParticipantRoleHost)})
	f.lk.listParticipantsFn = func(context.Context, string) ([]*lkproto.ParticipantInfo, error) {
		return []*lkproto.ParticipantInfo{{Identity: "host", Metadata: string(meta)}}, nil
	}

	err := svc.OnLiveKitEvent(context.Background(), "participant_left", room.LiveKitRoomName, "user-1")
	assert.NoError(t, err)
	f.parts.AssertExpectations(t)
}

func TestOnEgressEnded_Completed_FinalizesSizeAndDuration(t *testing.T) {
	svc, f := newTestServiceLK(t)
	rec := &domain.LiveRecording{
		ID:         uuid.New(),
		LiveRoomID: testRoomID,
		EgressID:   "EG_1",
		Status:     domain.LiveRecordingStatusStarted,
		StartedAt:  time.Now().Add(-time.Hour),
	}
	f.recs.On("FindByEgressID", mock.Anything, "EG_1").Return(rec, nil)
	f.recs.On("Update", mock.Anything, mock.MatchedBy(func(r *domain.LiveRecording) bool {
		return r.Status == domain.LiveRecordingStatusCompleted &&
			r.Size == 1024 && r.Duration == 3600 && r.EndedAt != nil
	})).Return(nil)

	err := svc.OnEgressEnded(context.Background(), domain.EgressResult{
		EgressID:  "EG_1",
		SizeBytes: 1024,
		Duration:  time.Hour,
	})
	assert.NoError(t, err)
	f.recs.AssertExpectations(t)
}

func TestOnEgressEnded_Failed_MarksFailed(t *testing.T) {
	svc, f := newTestServiceLK(t)
	rec := &domain.LiveRecording{
		ID:       uuid.New(),
		EgressID: "EG_2",
		Status:   domain.LiveRecordingStatusStarted,
	}
	f.recs.On("FindByEgressID", mock.Anything, "EG_2").Return(rec, nil)
	f.recs.On("Update", mock.Anything, mock.MatchedBy(func(r *domain.LiveRecording) bool {
		return r.Status == domain.LiveRecordingStatusFailed
	})).Return(nil)

	err := svc.OnEgressEnded(context.Background(), domain.EgressResult{EgressID: "EG_2", Failed: true})
	assert.NoError(t, err)
	f.recs.AssertExpectations(t)
}

func TestOnEgressEnded_UnknownEgress_NoError(t *testing.T) {
	svc, f := newTestServiceLK(t)
	f.recs.On("FindByEgressID", mock.Anything, "EG_missing").Return(nil, domain.ErrNotFound)

	err := svc.OnEgressEnded(context.Background(), domain.EgressResult{EgressID: "EG_missing"})
	assert.NoError(t, err)
	f.recs.AssertNotCalled(t, "Update")
}
