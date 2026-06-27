//go:build integration

package integration

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/classes"
	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/livesessions"
	"github.com/4H1R/zoora/internal/organizations"
	"github.com/4H1R/zoora/internal/users"
	"github.com/4H1R/zoora/tests/testutil"
)

type liveRoomRepos struct {
	rooms        domain.LiveRoomRepository
	participants domain.LiveParticipantRepository
	sessions     domain.ClassSessionRepository
	classes      domain.ClassRepository
	users        domain.UserRepository
	orgs         domain.OrganizationRepository
}

func setupLiveRoomsDB(t *testing.T) liveRoomRepos {
	t.Helper()
	db := testutil.SetupPostgres(t)
	require.NoError(t, db.AutoMigrate(
		&domain.Organization{},
		&domain.User{},
		&domain.Class{},
		&domain.ClassSession{},
		&domain.ClassMember{},
		&domain.LiveRoom{},
		&domain.LiveParticipant{},
	))
	return liveRoomRepos{
		rooms:        livesessions.NewRoomRepository(db),
		participants: livesessions.NewParticipantRepository(db),
		sessions:     classes.NewSessionRepository(db),
		classes:      classes.NewRepository(db),
		users:        users.NewRepository(db),
		orgs:         organizations.NewRepository(db),
	}
}

func seedLiveRoom(t *testing.T, r liveRoomRepos, sessionID uuid.UUID, name string, start time.Time) *domain.LiveRoom {
	t.Helper()
	scheduled := start
	room := &domain.LiveRoom{
		ClassSessionID:     sessionID,
		Name:               name,
		LiveKitRoomName:    "lk-" + uuid.NewString()[:8],
		Status:             domain.LiveRoomStatusCreated,
		Config:             domain.DefaultLiveRoomConfig(),
		ScheduledStartTime: &scheduled,
	}
	require.NoError(t, r.rooms.Create(context.Background(), room))
	return room
}

// TestIntegration_LiveRoomRepo_List_PreloadsHost verifies the list preloads the
// full ClassSession -> Class -> User chain so the UI can show the host teacher.
func TestIntegration_LiveRoomRepo_List_PreloadsHost(t *testing.T) {
	r := setupLiveRoomsDB(t)
	ctx := context.Background()
	org := seedOrg(t, r.orgs, "Acme")
	teacher := seedTeacher(t, r.users, org.ID, "ms-host")
	c := &domain.Class{OrganizationID: org.ID, UserID: teacher.ID, Name: "Algebra"}
	require.NoError(t, r.classes.Create(ctx, c))
	sess := &domain.ClassSession{ClassID: c.ID, Name: "Sess1", StartTime: time.Now()}
	require.NoError(t, r.sessions.Create(ctx, sess))
	seedLiveRoom(t, r, sess.ID, "Room A", time.Now().Add(time.Hour))

	rooms, total, err := r.rooms.List(ctx, domain.LiveRoomListScope{All: true}, domain.ListParams{Page: 1, PageSize: 50})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, rooms, 1)
	require.NotNil(t, rooms[0].ClassSession, "class_session must be preloaded")
	require.NotNil(t, rooms[0].ClassSession.Class, "class must be preloaded")
	require.NotNil(t, rooms[0].ClassSession.Class.User, "host teacher (class.user) must be preloaded")
	assert.Equal(t, "ms-host", rooms[0].ClassSession.Class.User.Name)
}

// TestIntegration_LiveRoomRepo_List_OrderAndSearch verifies the repository can
// order by scheduled_start_time and search by room name (the fields the online
// classes page relies on).
func TestIntegration_LiveRoomRepo_List_OrderAndSearch(t *testing.T) {
	r := setupLiveRoomsDB(t)
	ctx := context.Background()
	org := seedOrg(t, r.orgs, "Acme")
	teacher := seedTeacher(t, r.users, org.ID, "teach")
	c := &domain.Class{OrganizationID: org.ID, UserID: teacher.ID, Name: "Algebra"}
	require.NoError(t, r.classes.Create(ctx, c))
	sess := &domain.ClassSession{ClassID: c.ID, Name: "Sess1", StartTime: time.Now()}
	require.NoError(t, r.sessions.Create(ctx, sess))

	base := time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)
	seedLiveRoom(t, r, sess.ID, "Algebra Lecture", base.Add(2*time.Hour))
	seedLiveRoom(t, r, sess.ID, "Biology Lab", base)

	// Order by scheduled_start_time asc -> earlier-scheduled room first.
	ordered, _, err := r.rooms.List(ctx, domain.LiveRoomListScope{All: true},
		domain.ListParams{Page: 1, PageSize: 50, OrderBy: "scheduled_start_time", OrderDir: "asc"})
	require.NoError(t, err)
	require.Len(t, ordered, 2)
	require.NotNil(t, ordered[0].ScheduledStartTime)
	require.NotNil(t, ordered[1].ScheduledStartTime)
	assert.True(t, ordered[0].ScheduledStartTime.Before(*ordered[1].ScheduledStartTime),
		"rooms must be ordered by scheduled_start_time asc")
	assert.Equal(t, "Biology Lab", ordered[0].Name)

	// Search by room name.
	found, total, err := r.rooms.List(ctx, domain.LiveRoomListScope{All: true},
		domain.ListParams{Page: 1, PageSize: 50, Search: "Algebra", SearchFields: []string{"name"}})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, found, 1)
	assert.True(t, strings.Contains(strings.ToLower(found[0].Name), "algebra"))
}

// TestIntegration_LiveParticipantRole_Persistence verifies that
// UpdateParticipantRole, SetHandRaised, and GetActiveParticipant work
// correctly against a real Postgres database.
func TestIntegration_LiveParticipantRole_Persistence(t *testing.T) {
	r := setupLiveRoomsDB(t)
	ctx := context.Background()

	// Seed prerequisite entities.
	org := seedOrg(t, r.orgs, "Acme")
	teacher := seedTeacher(t, r.users, org.ID, "teacher-role-test")
	c := &domain.Class{OrganizationID: org.ID, UserID: teacher.ID, Name: "Role Test Class"}
	require.NoError(t, r.classes.Create(ctx, c))
	sess := &domain.ClassSession{ClassID: c.ID, Name: "Role Test Session", StartTime: time.Now()}
	require.NoError(t, r.sessions.Create(ctx, sess))
	room := seedLiveRoom(t, r, sess.ID, "Role Test Room", time.Now().Add(time.Hour))

	// Insert a participant directly via the participant repository.
	identity := teacher.ID.String()
	p := &domain.LiveParticipant{
		LiveRoomID: room.ID,
		UserID:     teacher.ID,
		Identity:   identity,
		JoinedAt:   time.Now(),
		Role:       domain.ParticipantRoleViewer,
	}
	require.NoError(t, r.participants.Create(ctx, p))

	// 1. UpdateParticipantRole: viewer -> presenter.
	require.NoError(t, r.participants.UpdateParticipantRole(ctx, room.ID, identity, domain.ParticipantRolePresenter))
	got, err := r.participants.GetActiveParticipant(ctx, room.ID, identity)
	require.NoError(t, err)
	assert.Equal(t, domain.ParticipantRolePresenter, got.Role)

	// 2. SetHandRaised true: HandRaisedAt must be populated.
	require.NoError(t, r.participants.SetHandRaised(ctx, room.ID, identity, true))
	got, err = r.participants.GetActiveParticipant(ctx, room.ID, identity)
	require.NoError(t, err)
	assert.NotNil(t, got.HandRaisedAt, "HandRaisedAt must be set after raising hand")

	// 3. SetHandRaised false: HandRaisedAt must be cleared.
	require.NoError(t, r.participants.SetHandRaised(ctx, room.ID, identity, false))
	got, err = r.participants.GetActiveParticipant(ctx, room.ID, identity)
	require.NoError(t, err)
	assert.Nil(t, got.HandRaisedAt, "HandRaisedAt must be nil after lowering hand")

	// 4. GetActiveParticipant for unknown identity returns ErrParticipantNotFound.
	_, err = r.participants.GetActiveParticipant(ctx, room.ID, "no-such-identity")
	assert.ErrorIs(t, err, domain.ErrParticipantNotFound)
}
