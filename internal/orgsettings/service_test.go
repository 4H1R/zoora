package orgsettings_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/domain"
)

func TestUpdate_RecordsAudit(t *testing.T) {
	orgID := uuid.New()
	settingsID := uuid.New()

	repo := &mockSettingsRepo{}
	repo.On("FindByOrgID", mock.Anything, orgID).
		Return(&domain.OrganizationSettings{ID: settingsID, OrganizationID: orgID, AttendancePresentThresholdPercent: 75}, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.OrganizationSettings")).Return(nil)

	audit := &auditSpy{}
	svc := newServiceWithAudit(repo, audit)
	ctx := domain.WithCaller(context.Background(), domain.Caller{UserID: uuid.New(), OrgID: &orgID})

	threshold := 90
	_, err := svc.Update(ctx, orgID, domain.UpdateOrganizationSettingsDTO{AttendancePresentThresholdPercent: &threshold})
	require.NoError(t, err)

	require.Len(t, audit.records, 1)
	rec := audit.records[0]
	require.Equal(t, domain.AuditUpdated, rec.Action)
	require.Equal(t, domain.AuditTargetOrgSettings, rec.TargetType)
	require.Equal(t, "organization settings", rec.TargetLabel)
	require.NotNil(t, rec.TargetID)
	require.Equal(t, settingsID, *rec.TargetID)
	require.NotNil(t, rec.OrgID)
	require.Equal(t, orgID, *rec.OrgID)

	changed, ok := rec.Metadata["changed"].(map[string]any)
	require.True(t, ok)
	require.Contains(t, changed, "attendance_present_threshold_percent")
}
