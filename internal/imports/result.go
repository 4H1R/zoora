package imports

import (
	"bytes"
	"fmt"

	"github.com/xuri/excelize/v2"
)

type RowStatus string

const (
	RowCreated RowStatus = "created"
	RowSkipped RowStatus = "skipped"
	RowError   RowStatus = "error"
)

// RowResult is the per-input-row outcome shown in the result file.
// GeneratedPassword is only set for users the system invented a password for
// — it exists nowhere else once the Redis copy expires.
type RowResult struct {
	Status            RowStatus
	Message           string
	GeneratedPassword string
}

func BuildUsersResult(rows []UserRow, results map[int]RowResult) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	sheet := f.GetSheetList()[0]

	header := []string{"name", "username", "role", "status", "message", "generated_password"}
	if err := writeRow(f, sheet, 1, header); err != nil {
		return nil, err
	}
	for i, r := range rows {
		res := results[r.RowNum]
		// input password intentionally omitted — never echo manual passwords
		row := []string{r.Name, r.Username, r.Role, string(res.Status), res.Message, res.GeneratedPassword}
		if err := writeRow(f, sheet, i+2, row); err != nil {
			return nil, err
		}
	}
	return save(f)
}

func BuildClassesResult(classRows []ClassRow, classResults map[int]RowResult, memberRows []MemberRow, memberResults map[int]RowResult) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	if err := f.SetSheetName(f.GetSheetList()[0], "Classes"); err != nil {
		return nil, fmt.Errorf("imports.BuildClassesResult: %w", err)
	}
	if _, err := f.NewSheet("Members"); err != nil {
		return nil, fmt.Errorf("imports.BuildClassesResult: %w", err)
	}

	if err := writeRow(f, "Classes", 1, []string{"class_name", "owner_username", "description", "capacity", "status", "message"}); err != nil {
		return nil, err
	}
	for i, r := range classRows {
		res := classResults[r.RowNum]
		if err := writeRow(f, "Classes", i+2, []string{r.Name, r.OwnerUsername, r.Description, r.Capacity, string(res.Status), res.Message}); err != nil {
			return nil, err
		}
	}

	if err := writeRow(f, "Members", 1, []string{"class_name", "member_username", "status", "message"}); err != nil {
		return nil, err
	}
	for i, r := range memberRows {
		res := memberResults[r.RowNum]
		if err := writeRow(f, "Members", i+2, []string{r.ClassName, r.MemberUsername, string(res.Status), res.Message}); err != nil {
			return nil, err
		}
	}
	return save(f)
}

func BuildClassMembersResult(rows []MemberRow, results map[int]RowResult) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	if err := f.SetSheetName(f.GetSheetList()[0], "Members"); err != nil {
		return nil, fmt.Errorf("imports.BuildClassMembersResult: %w", err)
	}

	if err := writeRow(f, "Members", 1, []string{"class_name", "member_username", "status", "message"}); err != nil {
		return nil, err
	}
	for i, r := range rows {
		res := results[r.RowNum]
		if err := writeRow(f, "Members", i+2, []string{r.ClassName, r.MemberUsername, string(res.Status), res.Message}); err != nil {
			return nil, err
		}
	}
	return save(f)
}

func writeRow(f *excelize.File, sheet string, rowNum int, values []string) error {
	cell, err := excelize.CoordinatesToCellName(1, rowNum)
	if err != nil {
		return fmt.Errorf("imports.writeRow: %w", err)
	}
	if err := f.SetSheetRow(sheet, cell, &values); err != nil {
		return fmt.Errorf("imports.writeRow: %w", err)
	}
	return nil
}

func save(f *excelize.File) ([]byte, error) {
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("imports.save result: %w", err)
	}
	return buf.Bytes(), nil
}
