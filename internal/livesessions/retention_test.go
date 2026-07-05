package livesessions_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/livesessions"
)

type fakeRetentionRepo struct {
	rows    []livesessions.RecordingWithPlan
	deleted []uuid.UUID
}

func (f *fakeRetentionRepo) ListRecordingsWithPlan(context.Context) ([]livesessions.RecordingWithPlan, error) {
	return f.rows, nil
}
func (f *fakeRetentionRepo) DeleteRecording(_ context.Context, id uuid.UUID) error {
	f.deleted = append(f.deleted, id)
	return nil
}

type fakeStore struct{ deletedKeys []string }

func (f *fakeStore) DeleteObject(_ context.Context, key string) error {
	f.deletedKeys = append(f.deletedKeys, key)
	return nil
}

func TestRetentionSweep_DeletesExpiredKeepsFresh(t *testing.T) {
	now := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	proOld := uuid.New()   // Pro (30d retention), 31 days old -> delete
	proFresh := uuid.New() // Pro, 29 days old -> keep
	freeOld := uuid.New()  // Free (0 = keep forever), very old -> keep

	repo := &fakeRetentionRepo{rows: []livesessions.RecordingWithPlan{
		{ID: proOld, FileURL: "orgs/a/rec/old.mp4", StartedAt: now.AddDate(0, 0, -31), Plan: domain.PlanPro},
		{ID: proFresh, FileURL: "orgs/a/rec/fresh.mp4", StartedAt: now.AddDate(0, 0, -29), Plan: domain.PlanPro},
		{ID: freeOld, FileURL: "orgs/b/rec/free.mp4", StartedAt: now.AddDate(-2, 0, 0), Plan: domain.PlanFree},
	}}
	store := &fakeStore{}
	sweeper := livesessions.NewRetentionSweeper(repo, store, slog.Default())

	require.NoError(t, sweeper.Sweep(context.Background(), now))

	assert.Equal(t, []uuid.UUID{proOld}, repo.deleted)
	assert.Equal(t, []string{"orgs/a/rec/old.mp4"}, store.deletedKeys)
}

func TestRetentionSweep_ExpiredPlanUsesFreeRetention(t *testing.T) {
	now := time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)
	past := now.AddDate(0, 0, -1) // plan expired yesterday -> effective Free
	rec := uuid.New()

	repo := &fakeRetentionRepo{rows: []livesessions.RecordingWithPlan{
		// Pro plan but expired -> effective Free (keep forever), so a 400-day-old
		// recording must NOT be deleted (grandfather).
		{ID: rec, FileURL: "orgs/a/rec/x.mp4", StartedAt: now.AddDate(0, 0, -400), Plan: domain.PlanPro, PlanExpiresAt: &past},
	}}
	store := &fakeStore{}
	sweeper := livesessions.NewRetentionSweeper(repo, store, slog.Default())

	require.NoError(t, sweeper.Sweep(context.Background(), now))
	assert.Empty(t, repo.deleted)
	assert.Empty(t, store.deletedKeys)
}
