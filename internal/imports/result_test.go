package imports_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"

	"github.com/4H1R/zoora/internal/imports"
)

func TestBuildUsersResult(t *testing.T) {
	rows := []imports.UserRow{
		{RowNum: 2, Name: "Ali", Username: "ali", Password: "manual-secret", Role: "student"},
		{RowNum: 3, Name: "Sara", Username: "sara", Role: "-"},
	}
	results := map[int]imports.RowResult{
		2: {Status: imports.RowCreated},
		3: {Status: imports.RowCreated, GeneratedPassword: "xk29fmp3aq"},
	}

	data, err := imports.BuildUsersResult(rows, results)
	require.NoError(t, err)

	f, err := excelize.OpenReader(bytes.NewReader(data))
	require.NoError(t, err)
	defer f.Close()
	got, err := f.GetRows(f.GetSheetList()[0])
	require.NoError(t, err)

	assert.Equal(t, []string{"name", "username", "role", "status", "message", "generated_password"}, got[0])
	// manual password never echoed anywhere
	for _, row := range got {
		assert.NotContains(t, row, "manual-secret")
	}
	assert.Equal(t, "xk29fmp3aq", got[2][5])
	assert.Equal(t, "created", got[1][3])
}

func TestBuildClassesResult(t *testing.T) {
	classRows := []imports.ClassRow{{RowNum: 2, Name: "Math-A", OwnerUsername: "ali", Capacity: "30"}}
	memberRows := []imports.MemberRow{{RowNum: 2, ClassName: "Math-A", MemberUsername: "sara"}}

	data, err := imports.BuildClassesResult(
		classRows, map[int]imports.RowResult{2: {Status: imports.RowCreated}},
		memberRows, map[int]imports.RowResult{2: {Status: imports.RowSkipped, Message: "already a member"}},
	)
	require.NoError(t, err)

	f, err := excelize.OpenReader(bytes.NewReader(data))
	require.NoError(t, err)
	defer f.Close()

	classes, err := f.GetRows("Classes")
	require.NoError(t, err)
	assert.Equal(t, "created", classes[1][4])

	members, err := f.GetRows("Members")
	require.NoError(t, err)
	assert.Equal(t, "skipped", members[1][2])
	assert.Equal(t, "already a member", members[1][3])
}
