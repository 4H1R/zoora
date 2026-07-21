package imports

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// MaxRows caps data rows per sheet — bcrypt cost makes bigger files a worker hog.
const MaxRows = 5000

// unzipSizeLimit and unzipXMLSizeLimit bound how much excelize will inflate
// from a zip-bomb .xlsx: the archive as a whole and any single worksheet XML
// member are each capped at 100MB, well above any legitimate MaxRows-sized
// import but far short of a maliciously crafted archive's default ceiling.
var parseOptions = excelize.Options{
	UnzipSizeLimit:    100 << 20,
	UnzipXMLSizeLimit: 100 << 20,
}

type UserRow struct {
	RowNum   int // 1-based Excel row
	Name     string
	Username string
	Password string
	Role     string
}

type ClassRow struct {
	RowNum        int
	Name          string
	OwnerUsername string
	Description   string
	Capacity      string // raw cell; numeric validation happens in the service
}

type MemberRow struct {
	RowNum         int
	ClassName      string
	MemberUsername string
}

func ParseUsersFile(data []byte) ([]UserRow, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data), parseOptions)
	if err != nil {
		return nil, fmt.Errorf("could not read xlsx file: %w", err)
	}
	defer f.Close()

	rows, err := sheetRows(f, "", 0)
	if err != nil {
		return nil, err
	}
	cols, err := headerIndex(rows[0], []string{"name", "username", "role"})
	if err != nil {
		return nil, err
	}

	var out []UserRow
	for i, row := range rows[1:] {
		r := UserRow{
			RowNum:   i + 2,
			Name:     cell(row, cols, "name"),
			Username: cell(row, cols, "username"),
			Password: cell(row, cols, "password"),
			Role:     cell(row, cols, "role"),
		}
		if r.Name == "" && r.Username == "" && r.Password == "" && r.Role == "" {
			continue // fully blank row
		}
		out = append(out, r)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("file has no data rows")
	}
	if len(out) > MaxRows {
		return nil, fmt.Errorf("file exceeds %d data rows", MaxRows)
	}
	return out, nil
}

func ParseClassesFile(data []byte) ([]ClassRow, []MemberRow, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data), parseOptions)
	if err != nil {
		return nil, nil, fmt.Errorf("could not read xlsx file: %w", err)
	}
	defer f.Close()

	classRows, err := sheetRows(f, "Classes", 0)
	if err != nil {
		return nil, nil, err
	}
	classCols, err := headerIndex(classRows[0], []string{"class_name", "owner_username"})
	if err != nil {
		return nil, nil, fmt.Errorf("reading Classes sheet: %w", err)
	}

	var classes []ClassRow
	for i, row := range classRows[1:] {
		r := ClassRow{
			RowNum:        i + 2,
			Name:          cell(row, classCols, "class_name"),
			OwnerUsername: cell(row, classCols, "owner_username"),
			Description:   cell(row, classCols, "description"),
			Capacity:      cell(row, classCols, "capacity"),
		}
		if r.Name == "" && r.OwnerUsername == "" && r.Description == "" && r.Capacity == "" {
			continue
		}
		classes = append(classes, r)
	}
	if len(classes) == 0 {
		return nil, nil, fmt.Errorf("classes sheet has no data rows")
	}
	if len(classes) > MaxRows {
		return nil, nil, fmt.Errorf("classes sheet exceeds %d data rows", MaxRows)
	}

	var members []MemberRow
	memberRows, err := sheetRows(f, "Members", 1)
	if err == nil { // Members sheet is optional
		memberCols, herr := headerIndex(memberRows[0], []string{"class_name", "member_username"})
		if herr != nil {
			return nil, nil, fmt.Errorf("reading Members sheet: %w", herr)
		}
		for i, row := range memberRows[1:] {
			r := MemberRow{
				RowNum:         i + 2,
				ClassName:      cell(row, memberCols, "class_name"),
				MemberUsername: cell(row, memberCols, "member_username"),
			}
			if r.ClassName == "" && r.MemberUsername == "" {
				continue
			}
			members = append(members, r)
		}
		if len(members) > MaxRows {
			return nil, nil, fmt.Errorf("members sheet exceeds %d data rows", MaxRows)
		}
	}
	return classes, members, nil
}

// ParseClassMembersFile reads a members-only import: a single sheet
// (named "Members" when present, otherwise the first sheet) with
// class_name + member_username columns targeting existing classes.
func ParseClassMembersFile(data []byte) ([]MemberRow, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data), parseOptions)
	if err != nil {
		return nil, fmt.Errorf("could not read xlsx file: %w", err)
	}
	defer f.Close()

	rows, err := sheetRows(f, "Members", 0)
	if err != nil {
		return nil, err
	}
	cols, err := headerIndex(rows[0], []string{"class_name", "member_username"})
	if err != nil {
		return nil, err
	}

	var out []MemberRow
	for i, row := range rows[1:] {
		r := MemberRow{
			RowNum:         i + 2,
			ClassName:      cell(row, cols, "class_name"),
			MemberUsername: cell(row, cols, "member_username"),
		}
		if r.ClassName == "" && r.MemberUsername == "" {
			continue
		}
		out = append(out, r)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("file has no data rows")
	}
	if len(out) > MaxRows {
		return nil, fmt.Errorf("file exceeds %d data rows", MaxRows)
	}
	return out, nil
}

// sheetRows finds a sheet by name (case-insensitive), falling back to index,
// and returns its rows. Errors when the sheet is missing or has no header row.
func sheetRows(f *excelize.File, name string, index int) ([][]string, error) {
	sheet := ""
	if name != "" {
		for _, s := range f.GetSheetList() {
			if strings.EqualFold(s, name) {
				sheet = s
				break
			}
		}
	}
	if sheet == "" {
		list := f.GetSheetList()
		if index >= len(list) {
			return nil, fmt.Errorf("sheet %q not found", name)
		}
		sheet = list[index]
	}
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("could not read sheet %q: %w", sheet, err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("sheet %q is empty", sheet)
	}
	return rows, nil
}

// headerIndex maps lowercase trimmed header names to column indexes and
// enforces the required set. Unknown columns are ignored; on duplicate
// headers the first occurrence wins.
func headerIndex(header []string, required []string) (map[string]int, error) {
	idx := make(map[string]int, len(header))
	for i, h := range header {
		key := strings.ToLower(strings.TrimSpace(h))
		if key == "" {
			continue
		}
		if _, exists := idx[key]; !exists {
			idx[key] = i
		}
	}
	for _, r := range required {
		if _, ok := idx[r]; !ok {
			return nil, fmt.Errorf("missing required column %q", r)
		}
	}
	return idx, nil
}

func cell(row []string, cols map[string]int, key string) string {
	i, ok := cols[key]
	if !ok || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}
