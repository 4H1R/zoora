package imports_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"

	"github.com/4H1R/zoora/internal/imports"
)

func usersXLSX(t *testing.T, rows [][]string) []byte {
	t.Helper()
	f := excelize.NewFile()
	for i, row := range rows {
		cell, err := excelize.CoordinatesToCellName(1, i+1)
		require.NoError(t, err)
		require.NoError(t, f.SetSheetRow("Sheet1", cell, &row))
	}
	var buf bytes.Buffer
	require.NoError(t, f.Write(&buf))
	return buf.Bytes()
}

func TestParseUsersFile(t *testing.T) {
	t.Run("happy path with shuffled headers and extra column", func(t *testing.T) {
		data := usersXLSX(t, [][]string{
			{"Role", "extra", "USERNAME", "name", "password"},
			{"student", "x", "ali.r", "Ali Rezaei", ""},
			{"-", "", "sara.k", "Sara K", "secret123"},
		})
		rows, err := imports.ParseUsersFile(data)
		require.NoError(t, err)
		require.Len(t, rows, 2)
		assert.Equal(t, imports.UserRow{RowNum: 2, Name: "Ali Rezaei", Username: "ali.r", Password: "", Role: "student"}, rows[0])
		assert.Equal(t, imports.UserRow{RowNum: 3, Name: "Sara K", Username: "sara.k", Password: "secret123", Role: "-"}, rows[1])
	})

	t.Run("missing required header", func(t *testing.T) {
		data := usersXLSX(t, [][]string{{"name", "username"}, {"Ali", "ali"}})
		_, err := imports.ParseUsersFile(data)
		assert.ErrorContains(t, err, "role")
	})

	t.Run("blank rows skipped", func(t *testing.T) {
		data := usersXLSX(t, [][]string{
			{"name", "username", "role"},
			{"", "", ""},
			{"Ali", "ali", "-"},
		})
		rows, err := imports.ParseUsersFile(data)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		assert.Equal(t, 3, rows[0].RowNum)
	})

	t.Run("no data rows", func(t *testing.T) {
		data := usersXLSX(t, [][]string{{"name", "username", "role"}})
		_, err := imports.ParseUsersFile(data)
		assert.Error(t, err)
	})

	t.Run("not an xlsx", func(t *testing.T) {
		_, err := imports.ParseUsersFile([]byte("definitely,a,csv"))
		assert.Error(t, err)
	})
}

func classesXLSX(t *testing.T, classes, members [][]string) []byte {
	t.Helper()
	f := excelize.NewFile()
	require.NoError(t, f.SetSheetName("Sheet1", "Classes"))
	for i, row := range classes {
		cell, err := excelize.CoordinatesToCellName(1, i+1)
		require.NoError(t, err)
		require.NoError(t, f.SetSheetRow("Classes", cell, &row))
	}
	if members != nil {
		_, err := f.NewSheet("Members")
		require.NoError(t, err)
		for i, row := range members {
			cell, err := excelize.CoordinatesToCellName(1, i+1)
			require.NoError(t, err)
			require.NoError(t, f.SetSheetRow("Members", cell, &row))
		}
	}
	var buf bytes.Buffer
	require.NoError(t, f.Write(&buf))
	return buf.Bytes()
}

func TestParseClassesFile(t *testing.T) {
	t.Run("both sheets", func(t *testing.T) {
		data := classesXLSX(t,
			[][]string{
				{"class_name", "owner_username", "description", "capacity"},
				{"Math-A", "ali.r", "Algebra", "30"},
			},
			[][]string{
				{"class_name", "member_username"},
				{"Math-A", "sara.k"},
			})
		classes, members, err := imports.ParseClassesFile(data)
		require.NoError(t, err)
		require.Len(t, classes, 1)
		require.Len(t, members, 1)
		assert.Equal(t, imports.ClassRow{RowNum: 2, Name: "Math-A", OwnerUsername: "ali.r", Description: "Algebra", Capacity: "30"}, classes[0])
		assert.Equal(t, imports.MemberRow{RowNum: 2, ClassName: "Math-A", MemberUsername: "sara.k"}, members[0])
	})

	t.Run("members sheet optional", func(t *testing.T) {
		data := classesXLSX(t, [][]string{
			{"class_name", "owner_username"},
			{"Math-A", "ali.r"},
		}, nil)
		classes, members, err := imports.ParseClassesFile(data)
		require.NoError(t, err)
		assert.Len(t, classes, 1)
		assert.Empty(t, members)
	})
}

// membersXLSX builds a single-sheet class-members import file. The sheet is
// left as the default "Sheet1" to exercise the first-sheet fallback.
func membersXLSX(t *testing.T, rows [][]string) []byte {
	t.Helper()
	f := excelize.NewFile()
	for i, row := range rows {
		cell, err := excelize.CoordinatesToCellName(1, i+1)
		require.NoError(t, err)
		require.NoError(t, f.SetSheetRow("Sheet1", cell, &row))
	}
	var buf bytes.Buffer
	require.NoError(t, f.Write(&buf))
	return buf.Bytes()
}

func TestParseClassMembersFile(t *testing.T) {
	t.Run("happy path with extra column and blank row", func(t *testing.T) {
		data := membersXLSX(t, [][]string{
			{"CLASS_NAME", "extra", "member_username"},
			{"Math-A", "x", "sara.k"},
			{"", "", ""},
			{"Math-B", "", "ali.r"},
		})
		rows, err := imports.ParseClassMembersFile(data)
		require.NoError(t, err)
		require.Len(t, rows, 2)
		assert.Equal(t, imports.MemberRow{RowNum: 2, ClassName: "Math-A", MemberUsername: "sara.k"}, rows[0])
		assert.Equal(t, imports.MemberRow{RowNum: 4, ClassName: "Math-B", MemberUsername: "ali.r"}, rows[1])
	})

	t.Run("missing required column", func(t *testing.T) {
		data := membersXLSX(t, [][]string{
			{"class_name"},
			{"Math-A"},
		})
		_, err := imports.ParseClassMembersFile(data)
		require.ErrorContains(t, err, "member_username")
	})

	t.Run("no data rows", func(t *testing.T) {
		data := membersXLSX(t, [][]string{
			{"class_name", "member_username"},
		})
		_, err := imports.ParseClassMembersFile(data)
		require.ErrorContains(t, err, "no data rows")
	})
}
