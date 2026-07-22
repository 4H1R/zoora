package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/domain"
)

func TestCustomFieldsManagePermissionIsRegistered(t *testing.T) {
	require.Contains(t, domain.AllPermissions, domain.PermCustomFieldsManage)
	require.Contains(t, domain.ManagerPermissions, domain.PermCustomFieldsManage)
	require.NotContains(t, domain.StudentPermissions, domain.PermCustomFieldsManage)
	require.NotContains(t, domain.TeacherPermissions, domain.PermCustomFieldsManage)
}

func customDef(t domain.CustomFieldType, opts ...string) domain.UserCustomFieldDefinition {
	return domain.UserCustomFieldDefinition{FieldType: t, Options: opts}
}

func TestValidateCustomFieldValue(t *testing.T) {
	tests := []struct {
		name    string
		def     domain.UserCustomFieldDefinition
		value   any
		wantErr bool
	}{
		{"text ok", customDef(domain.CustomFieldTypeText), "hello", false},
		{"text rejects number", customDef(domain.CustomFieldTypeText), 3.0, true},
		{"number ok", customDef(domain.CustomFieldTypeNumber), 42.0, false},
		{"number rejects string", customDef(domain.CustomFieldTypeNumber), "42", true},
		{"boolean ok", customDef(domain.CustomFieldTypeBoolean), true, false},
		{"boolean rejects string", customDef(domain.CustomFieldTypeBoolean), "true", true},
		{"date ok", customDef(domain.CustomFieldTypeDate), "2026-07-21", false},
		{"date rejects garbage", customDef(domain.CustomFieldTypeDate), "not-a-date", true},
		{"select ok", customDef(domain.CustomFieldTypeSelect, "a", "b"), "a", false},
		{"select rejects unknown", customDef(domain.CustomFieldTypeSelect, "a", "b"), "c", true},
		{"unknown type rejected", customDef("weird"), "x", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := domain.ValidateCustomFieldValue(tc.def, tc.value)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCustomFieldValueToText(t *testing.T) {
	require.Equal(t, "42", domain.CustomFieldValueToText(42.0))
	require.Equal(t, "5", domain.CustomFieldValueToText(5.0))
	require.Equal(t, "true", domain.CustomFieldValueToText(true))
	require.Equal(t, "hi", domain.CustomFieldValueToText("hi"))
}
