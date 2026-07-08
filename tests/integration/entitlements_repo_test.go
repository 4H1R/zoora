//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/entitlements"
	"github.com/4H1R/zoora/tests/testutil"
)

func setupEntitlementsRepoDB(t *testing.T) (*gorm.DB, entitlements.Repository) {
	t.Helper()
	db := testutil.SetupPostgres(t)
	require.NoError(t, db.AutoMigrate(
		&domain.Organization{},
		&domain.User{},
		&domain.Class{},
		&domain.ClassSession{},
		&domain.LiveRoom{},
		&domain.LiveRecording{},
		&domain.Media{},
	))
	return db, entitlements.NewRepository(db)
}

// seedRoomChain creates org -> class -> session -> room and returns the room ID.
func seedRoomChain(t *testing.T, db *gorm.DB, orgID uuid.UUID, status domain.LiveRoomStatus, roomName string) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	user := &domain.User{OrganizationID: &orgID, Username: "u-" + roomName, Name: "u", Password: "x"}
	require.NoError(t, db.WithContext(ctx).Create(user).Error)
	class := &domain.Class{OrganizationID: orgID, UserID: user.ID, Name: "c"}
	require.NoError(t, db.WithContext(ctx).Create(class).Error)
	session := &domain.ClassSession{ClassID: class.ID, Name: "s", StartTime: time.Unix(1_700_000_000, 0)}
	require.NoError(t, db.WithContext(ctx).Create(session).Error)
	room := &domain.LiveRoom{
		ClassSessionID:  session.ID,
		Name:            roomName,
		LiveKitRoomName: roomName,
		Status:          status,
		Config:          domain.DefaultLiveRoomConfig(),
	}
	require.NoError(t, db.WithContext(ctx).Create(room).Error)
	return room.ID
}

func TestIntegration_Entitlements_SumStorageBytes(t *testing.T) {
	db, repo := setupEntitlementsRepoDB(t)
	ctx := context.Background()

	org := &domain.Organization{Name: "o", Slug: "sum-store", Plan: domain.PlanKey(domain.TierPro, 50)}
	require.NoError(t, db.Create(org).Error)

	// Two media rows for this org.
	for _, size := range []int64{1024, 2048} {
		require.NoError(t, db.Create(&domain.Media{
			OrganizationID: &org.ID, ModelType: "x", ModelID: uuid.New(), FileName: "f", Size: size,
		}).Error)
	}
	// Media belonging to another org must be excluded.
	otherOrg := &domain.Organization{Name: "other", Slug: "sum-store-other", Plan: domain.PlanFree}
	require.NoError(t, db.Create(otherOrg).Error)
	require.NoError(t, db.Create(&domain.Media{
		OrganizationID: &otherOrg.ID, ModelType: "x", ModelID: uuid.New(), FileName: "f", Size: 9999,
	}).Error)

	// One recording (4096) reachable via the class chain.
	roomID := seedRoomChain(t, db, org.ID, domain.LiveRoomStatusFinished, "rec-room")
	require.NoError(t, db.Create(&domain.LiveRecording{
		LiveRoomID: roomID, EgressID: "e", Size: 4096, StartedAt: time.Unix(1_700_000_000, 0),
	}).Error)

	got, err := repo.SumStorageBytes(ctx, org.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1024+2048+4096), got)
}

func TestIntegration_Entitlements_CountActiveLiveRooms(t *testing.T) {
	db, repo := setupEntitlementsRepoDB(t)
	ctx := context.Background()

	org := &domain.Organization{Name: "o", Slug: "count-rooms", Plan: domain.PlanFree}
	require.NoError(t, db.Create(org).Error)

	seedRoomChain(t, db, org.ID, domain.LiveRoomStatusActive, "active-1")
	seedRoomChain(t, db, org.ID, domain.LiveRoomStatusFinished, "finished-1")
	// A soft-deleted active room must not count.
	deletedRoomID := seedRoomChain(t, db, org.ID, domain.LiveRoomStatusActive, "active-deleted")
	require.NoError(t, db.Delete(&domain.LiveRoom{}, "id = ?", deletedRoomID).Error)

	got, err := repo.CountActiveLiveRooms(ctx, org.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), got)
}

func TestIntegration_Entitlements_GetOrgPlan(t *testing.T) {
	db, repo := setupEntitlementsRepoDB(t)
	ctx := context.Background()

	exp := time.Unix(1_800_000_000, 0)
	org := &domain.Organization{Name: "o", Slug: "get-plan", Plan: domain.PlanKey(domain.TierPro, 50), PlanExpiresAt: &exp}
	require.NoError(t, db.Create(org).Error)

	plan, gotExp, err := repo.GetOrgPlan(ctx, org.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.PlanKey(domain.TierPro, 50), plan)
	require.NotNil(t, gotExp)
	assert.WithinDuration(t, exp, *gotExp, time.Second)
}
