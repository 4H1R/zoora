//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/4H1R/zoora/internal/calendar"
	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/tests/testutil"
)

func setupCalendarDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testutil.SetupPostgres(t)
	require.NoError(t, db.AutoMigrate(
		&domain.Organization{},
		&domain.User{},
		&domain.Class{},
		&domain.ClassSession{},
		&domain.ClassMember{},
		&domain.LiveRoom{},
		&domain.Quiz{},
		&domain.QuizRoom{},
		&domain.PracticeRoom{},
		&domain.OfflineRoom{},
	))
	return db
}

// TestIntegration_CalendarRepo_ListEvents_ScopeAndRange seeds one class with a
// live room inside the window and asserts: org scope returns it, a stranger's
// member scope does not, the owner sees it, and an out-of-range window excludes it.
func TestIntegration_CalendarRepo_ListEvents_ScopeAndRange(t *testing.T) {
	db := setupCalendarDB(t)
	ctx := context.Background()

	org := &domain.Organization{Name: "Acme"}
	require.NoError(t, db.Create(org).Error)

	teacherID := uuid.New()
	teacher := &domain.User{ID: teacherID, OrganizationID: &org.ID, Username: "teach", Name: "teach", Password: "x"}
	require.NoError(t, db.Create(teacher).Error)

	class := &domain.Class{OrganizationID: org.ID, UserID: teacherID, Name: "Math"}
	require.NoError(t, db.Create(class).Error)

	session := &domain.ClassSession{ClassID: class.ID, Name: "S1", StartTime: time.Now()}
	require.NoError(t, db.Create(session).Error)

	start := time.Now().Add(2 * time.Hour)
	live := &domain.LiveRoom{
		ClassSessionID:     session.ID,
		Name:               "Lecture",
		LiveKitRoomName:    "lk-" + uuid.NewString(),
		ScheduledStartTime: &start,
	}
	require.NoError(t, db.Create(live).Error)

	repo := calendar.NewRepository(db)
	rng := domain.CalendarRange{From: time.Now(), To: time.Now().Add(24 * time.Hour)}

	// (1) org-wide scope returns the live room.
	got, err := repo.ListEvents(ctx, domain.ClassListScope{All: true, OrganizationID: &org.ID}, rng)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, domain.CalendarEventLive, got[0].Type)
	require.Equal(t, live.ID, got[0].EntityID)
	require.Equal(t, "Math", got[0].ClassName)

	// (2) a stranger sees nothing.
	stranger := uuid.New()
	got, err = repo.ListEvents(ctx, domain.ClassListScope{TeacherID: &stranger, MemberUserID: &stranger}, rng)
	require.NoError(t, err)
	require.Len(t, got, 0)

	// (3) the owning teacher sees it.
	got, err = repo.ListEvents(ctx, domain.ClassListScope{TeacherID: &teacherID, MemberUserID: &teacherID}, rng)
	require.NoError(t, err)
	require.Len(t, got, 1)

	// (4) an out-of-range window excludes it.
	past := domain.CalendarRange{From: time.Now().Add(-48 * time.Hour), To: time.Now().Add(-24 * time.Hour)}
	got, err = repo.ListEvents(ctx, domain.ClassListScope{All: true, OrganizationID: &org.ID}, past)
	require.NoError(t, err)
	require.Len(t, got, 0)
}
