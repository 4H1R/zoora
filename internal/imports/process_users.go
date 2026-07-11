package imports

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
)

func (s *service) processUsers(ctx context.Context, job *domain.ImportJob, data []byte) error {
	caller, _ := domain.CallerFromCtx(ctx)

	rows, err := ParseUsersFile(data)
	if err != nil {
		return s.fail(ctx, job, err.Error())
	}
	job.TotalRows = len(rows)
	if err := s.repo.Update(ctx, job); err != nil {
		return err
	}

	roleByName, err := s.roleLookup(ctx, job.OrganizationID)
	if err != nil {
		return s.fail(ctx, job, "could not load roles")
	}

	// normalize usernames once; Excel data is routinely mixed-case
	usernames := make([]string, 0, len(rows))
	for i := range rows {
		rows[i].Username = strings.ToLower(strings.TrimSpace(rows[i].Username))
		usernames = append(usernames, rows[i].Username)
	}
	existingUsers, err := s.users.FindByUsernames(ctx, job.OrganizationID, usernames)
	if err != nil {
		return s.fail(ctx, job, "could not check existing users")
	}
	existing := make(map[string]bool, len(existingUsers))
	for _, u := range existingUsers {
		existing[u.Username] = true
	}

	type pendingUser struct {
		row    UserRow
		roleID *uuid.UUID
		gen    string
	}
	results := make(map[int]RowResult, len(rows))
	seen := make(map[string]bool, len(rows))
	var toCreate []pendingUser
	skipped, failed := 0, 0

	for _, r := range rows {
		if msg := validateUserRow(r); msg != "" {
			results[r.RowNum] = RowResult{Status: RowError, Message: msg}
			failed++
			continue
		}
		if seen[r.Username] {
			results[r.RowNum] = RowResult{Status: RowError, Message: "duplicate username in file"}
			failed++
			continue
		}
		seen[r.Username] = true
		if existing[r.Username] {
			results[r.RowNum] = RowResult{Status: RowSkipped, Message: "username already exists"}
			skipped++
			continue
		}
		roleID, msg := resolveRole(caller, roleByName, r.Role)
		if msg != "" {
			results[r.RowNum] = RowResult{Status: RowError, Message: msg}
			failed++
			continue
		}
		gen := ""
		if r.Password == "" {
			gen, err = generatePassword()
			if err != nil {
				return s.fail(ctx, job, "password generation failed")
			}
		}
		toCreate = append(toCreate, pendingUser{row: r, roleID: roleID, gen: gen})
	}

	// Seat limit is all-or-nothing: never fill "up to the limit" — which rows
	// would win the remaining seats is nondeterministic.
	if len(toCreate) > 0 {
		if err := s.ent.CheckUserLimitN(ctx, job.OrganizationID, caller.Ent, int64(len(toCreate))); err != nil {
			return s.fail(ctx, job, fmt.Sprintf("plan seat limit: importing %d new users exceeds the remaining seats", len(toCreate)))
		}
	}

	created := 0
	processed := len(rows) - len(toCreate) // validation-phase rows count as processed
	for i, p := range toCreate {
		pass := p.row.Password
		if pass == "" {
			pass = p.gen
		}
		hashed, hashErr := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		if hashErr != nil {
			results[p.row.RowNum] = RowResult{Status: RowError, Message: "internal error"}
			failed++
		} else {
			user := &domain.User{
				OrganizationID: &job.OrganizationID,
				Username:       p.row.Username,
				Name:           p.row.Name,
				Password:       string(hashed),
				RoleID:         p.roleID,
			}
			switch createErr := s.users.Create(ctx, user); {
			case createErr == nil:
				results[p.row.RowNum] = RowResult{Status: RowCreated, GeneratedPassword: p.gen}
				created++
			case errors.Is(createErr, domain.ErrConflict):
				results[p.row.RowNum] = RowResult{Status: RowSkipped, Message: "username already exists"}
				skipped++
			default:
				s.logger.Error("import user create failed", "job_id", job.ID.String(), "row", p.row.RowNum, "error", createErr)
				results[p.row.RowNum] = RowResult{Status: RowError, Message: "internal error"}
				failed++
			}
		}
		processed++
		if (i+1)%progressBatch == 0 {
			_ = s.repo.UpdateProgress(ctx, job.ID, processed, created, skipped, failed)
		}
	}

	resultFile, err := BuildUsersResult(rows, results)
	if err != nil {
		return s.fail(ctx, job, "could not build result file")
	}
	if err := s.results.Set(ctx, job.ID, resultFile); err != nil {
		// counters still land; user loses the passwords file — log loudly
		s.logger.Error("import result store failed", "job_id", job.ID.String(), "error", err)
	}

	job.Status = domain.ImportStatusCompleted
	job.ProcessedRows = len(rows)
	job.CreatedCount = created
	job.SkippedCount = skipped
	job.FailedCount = failed
	return s.repo.Update(ctx, job)
}

func validateUserRow(r UserRow) string {
	if len([]rune(strings.TrimSpace(r.Name))) < 2 {
		return "name must be at least 2 characters"
	}
	if !httpx.ValidUsername(r.Username) {
		return "username must be 3-30 chars: lowercase letters, digits, dot or underscore"
	}
	if r.Password != "" && len(r.Password) < 8 {
		return "password must be at least 8 characters"
	}
	if r.Role == "" {
		return `role is required; use "-" for no role`
	}
	return ""
}

// resolveRole maps a role cell to a role ID. Mirrors the manual-create
// guards (internal/users/service.go) but per-row and explicit instead of
// silently dropping the role.
func resolveRole(caller domain.Caller, roleByName map[string][]domain.Role, cellValue string) (*uuid.UUID, string) {
	if cellValue == "-" {
		return nil, ""
	}
	matches := roleByName[strings.ToLower(strings.TrimSpace(cellValue))]
	if len(matches) == 0 {
		return nil, "unknown role"
	}
	if len(matches) > 1 {
		return nil, "ambiguous role name"
	}
	role := matches[0]
	if !caller.IsAdmin {
		if !caller.HasPermission(domain.PermRolesUpdate) {
			return nil, "you are not allowed to assign roles"
		}
		if role.IsPreset && role.Name == domain.PresetRoleManager {
			return nil, "the Manager role cannot be assigned via import"
		}
	}
	return &role.ID, ""
}

func (s *service) roleLookup(ctx context.Context, orgID uuid.UUID) (map[string][]domain.Role, error) {
	list, err := s.roles.List(ctx, domain.RoleFilter{OrganizationID: &orgID, IncludePreset: true})
	if err != nil {
		return nil, err
	}
	m := make(map[string][]domain.Role, len(list))
	for _, r := range list {
		key := strings.ToLower(r.Name)
		m[key] = append(m[key], r)
	}
	return m, nil
}
