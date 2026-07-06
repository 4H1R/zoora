package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestAudienceValidate(t *testing.T) {
	orgID := uuid.New()
	classID := uuid.New()
	roleID := uuid.New()
	cases := []struct {
		name    string
		a       NotificationAudience
		wantErr bool
	}{
		{"all ok", NotificationAudience{Type: AudienceAll}, false},
		{"org needs org_id", NotificationAudience{Type: AudienceOrg}, true},
		{"org ok", NotificationAudience{Type: AudienceOrg, OrgID: &orgID}, false},
		{"class needs class_id", NotificationAudience{Type: AudienceClass}, true},
		{"class ok", NotificationAudience{Type: AudienceClass, ClassID: &classID}, false},
		{"role needs role_id", NotificationAudience{Type: AudienceRole}, true},
		{"role ok", NotificationAudience{Type: AudienceRole, RoleID: &roleID}, false},
		{"role with org ok", NotificationAudience{Type: AudienceRole, RoleID: &roleID, OrgID: &orgID}, false},
		{"users needs ids", NotificationAudience{Type: AudienceUsers}, true},
		{"users ok", NotificationAudience{Type: AudienceUsers, UserIDs: []uuid.UUID{uuid.New()}}, false},
		{"unknown type", NotificationAudience{Type: "nope"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.a.Validate()
			if (err != nil) != tc.wantErr {
				t.Fatalf("Validate() err = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
