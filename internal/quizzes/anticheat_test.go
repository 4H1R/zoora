package quizzes

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestHaversineWithin50m(t *testing.T) {
	// two points ~30m apart
	d := haversineMeters(35.700000, 51.400000, 35.700270, 51.400000)
	assert.InDelta(t, 30, d, 5)
}

func TestSameLocationClustering(t *testing.T) {
	uA, uB, uC := uuid.New(), uuid.New(), uuid.New()
	acc := 20.0
	subs := []gpsPoint{
		{UserID: uA, Lat: 35.700000, Lng: 51.400000, Acc: &acc},
		{UserID: uB, Lat: 35.700270, Lng: 51.400000, Acc: &acc}, // ~30m from A
		{UserID: uC, Lat: 35.800000, Lng: 51.500000, Acc: &acc}, // far
	}
	got := clusterSameLocation(subs)
	assert.ElementsMatch(t, []uuid.UUID{uB}, got[uA])
	assert.ElementsMatch(t, []uuid.UUID{uA}, got[uB])
	assert.Empty(t, got[uC])
}

func TestSameLocationClustering_IgnoresCoarseAccuracy(t *testing.T) {
	uA, uB := uuid.New(), uuid.New()
	coarse := 500.0 // worse than SameLocationMaxAccuracy → excluded
	subs := []gpsPoint{
		{UserID: uA, Lat: 35.700000, Lng: 51.400000, Acc: &coarse},
		{UserID: uB, Lat: 35.700270, Lng: 51.400000, Acc: &coarse},
	}
	got := clusterSameLocation(subs)
	assert.Empty(t, got[uA])
	assert.Empty(t, got[uB])
}
