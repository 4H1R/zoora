package domain

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestCallerContextAndPermissionChecks(t *testing.T) {
	ctx := context.Background()
	if _, ok := CallerFromCtx(ctx); ok {
		t.Fatal("CallerFromCtx returned ok for context without caller")
	}

	roleID := uuid.New()
	caller := Caller{
		UserID:      uuid.New(),
		RoleID:      &roleID,
		Permissions: []string{string(PermUsersView), string(PermClassesCreate)},
	}
	ctx = WithCaller(ctx, caller)

	got, ok := CallerFromCtx(ctx)
	if !ok {
		t.Fatal("CallerFromCtx did not find caller")
	}
	if got.UserID != caller.UserID || got.RoleID == nil || *got.RoleID != roleID {
		t.Fatalf("caller = %#v, want original caller", got)
	}
	if !got.HasPermission(PermUsersView) {
		t.Fatal("HasPermission returned false for assigned permission")
	}
	if got.HasPermission(PermUsersDelete) {
		t.Fatal("HasPermission returned true for missing permission")
	}
}

func TestValidationErrorFormattingAndUnwrap(t *testing.T) {
	err := NewValidationError(map[string]string{"name": "required", "status": "invalid"})

	if !errors.Is(err, ErrValidation) {
		t.Fatal("ValidationError should unwrap to ErrValidation")
	}
	msg := err.Error()
	if !strings.Contains(msg, "validation failed:") || !strings.Contains(msg, "name: required") || !strings.Contains(msg, "status: invalid") {
		t.Fatalf("unexpected validation message: %q", msg)
	}

	empty := NewValidationError(nil)
	if empty.Error() != ErrValidation.Error() {
		t.Fatalf("empty validation error = %q, want %q", empty.Error(), ErrValidation.Error())
	}
}

func TestListParamsOffsetAndLimitBoundaries(t *testing.T) {
	tests := []struct {
		name       string
		params     ListParams
		wantOffset int
		wantLimit  int
	}{
		{"zero values use defaults", ListParams{}, 0, DefaultPageSize},
		{"negative page clamps offset", ListParams{Page: -3, PageSize: 10}, 0, 10},
		{"first page offset zero", ListParams{Page: 1, PageSize: 25}, 0, 25},
		{"later page uses effective limit", ListParams{Page: 4, PageSize: 15}, 45, 15},
		{"later page with default limit", ListParams{Page: 2}, DefaultPageSize, DefaultPageSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.params.Offset(); got != tt.wantOffset {
				t.Fatalf("Offset() = %d, want %d", got, tt.wantOffset)
			}
			if got := tt.params.Limit(); got != tt.wantLimit {
				t.Fatalf("Limit() = %d, want %d", got, tt.wantLimit)
			}
		})
	}
}

func TestAllPermissionsAreUniqueAndNonEmpty(t *testing.T) {
	seen := map[PermissionName]bool{}
	for _, perm := range AllPermissions {
		if perm == "" {
			t.Fatal("AllPermissions contains an empty permission")
		}
		if seen[perm] {
			t.Fatalf("AllPermissions contains duplicate permission %q", perm)
		}
		seen[perm] = true
	}

	for _, required := range []PermissionName{
		PermUsersView,
		PermOrganizationsUpdate,
		PermClassesJoin,
		PermLiveSessionsManage,
		PermMediaCreate,
		PermAttendanceDelete,
	} {
		if !seen[required] {
			t.Fatalf("AllPermissions missing required permission %q", required)
		}
	}
}

func TestMediaS3KeyIncludesModelCollectionAndFile(t *testing.T) {
	modelID := uuid.New()
	m := Media{
		ModelType:      "practice",
		ModelID:        modelID,
		CollectionName: "attachments",
		FileName:       "solution.pdf",
	}

	want := "practice/" + modelID.String() + "/attachments/solution.pdf"
	if got := m.S3Key(); got != want {
		t.Fatalf("S3Key() = %q, want %q", got, want)
	}
}

func TestMediaS3KeyNamespacesByOrganization(t *testing.T) {
	orgID := uuid.New()
	modelID := uuid.New()
	m := Media{
		OrganizationID: &orgID,
		ModelType:      "practice",
		ModelID:        modelID,
		CollectionName: "attachments",
		FileName:       "solution.pdf",
	}

	want := "orgs/" + orgID.String() + "/practice/" + modelID.String() + "/attachments/solution.pdf"
	if got := m.S3Key(); got != want {
		t.Fatalf("S3Key() = %q, want %q", got, want)
	}
}
