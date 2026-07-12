package imports

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// processClassMembers enrolls members (by username) into existing classes
// (by name). Unlike processClasses it never creates classes — a class_name
// that doesn't resolve to exactly one existing class errors the row.
func (s *service) processClassMembers(ctx context.Context, job *domain.ImportJob, data []byte) error {
	rows, err := ParseClassMembersFile(data)
	if err != nil {
		return s.fail(ctx, job, err.Error())
	}
	job.TotalRows = len(rows)
	if err := s.repo.Update(ctx, job); err != nil {
		return err
	}

	// resolve every referenced username in one query
	usernameSet := map[string]bool{}
	for i := range rows {
		rows[i].MemberUsername = strings.ToLower(strings.TrimSpace(rows[i].MemberUsername))
		if rows[i].MemberUsername != "" {
			usernameSet[rows[i].MemberUsername] = true
		}
	}
	usernames := make([]string, 0, len(usernameSet))
	for u := range usernameSet {
		usernames = append(usernames, u)
	}
	found, err := s.users.FindByUsernames(ctx, job.OrganizationID, usernames)
	if err != nil {
		return s.fail(ctx, job, "could not resolve usernames")
	}
	userIDByName := make(map[string]uuid.UUID, len(found))
	for _, u := range found {
		userIDByName[u.Username] = u.ID
	}

	nameSet := map[string]bool{}
	for _, r := range rows {
		if r.ClassName != "" {
			nameSet[r.ClassName] = true
		}
	}
	names := make([]string, 0, len(nameSet))
	for n := range nameSet {
		names = append(names, n)
	}
	existingClasses, err := s.classes.ListByNames(ctx, job.OrganizationID, names)
	if err != nil {
		return s.fail(ctx, job, "could not resolve classes")
	}
	classesByName := map[string][]domain.Class{}
	for _, c := range existingClasses {
		classesByName[c.Name] = append(classesByName[c.Name], c)
	}

	results := make(map[int]RowResult, len(rows))
	memberCount := map[uuid.UUID]int64{} // lazily loaded per capacity-limited class
	created, skipped, failed, processed := 0, 0, 0, 0

	progress := func() {
		processed++
		if processed%progressBatch == 0 {
			_ = s.repo.UpdateProgress(ctx, job.ID, processed, created, skipped, failed)
		}
	}

	for _, r := range rows {
		if r.ClassName == "" || r.MemberUsername == "" {
			results[r.RowNum] = RowResult{Status: RowError, Message: "class_name and member_username are required"}
			failed++
			progress()
			continue
		}
		var class domain.Class
		switch ex := classesByName[r.ClassName]; {
		case len(ex) == 1:
			class = ex[0]
		case len(ex) > 1:
			results[r.RowNum] = RowResult{Status: RowError, Message: "multiple classes with this name exist"}
			failed++
			progress()
			continue
		default:
			results[r.RowNum] = RowResult{Status: RowError, Message: "class not found"}
			failed++
			progress()
			continue
		}
		userID, ok := userIDByName[r.MemberUsername]
		if !ok {
			results[r.RowNum] = RowResult{Status: RowError, Message: "member username not found"}
			failed++
			progress()
			continue
		}
		if limit := class.TotalUsers; limit > 0 {
			if _, loaded := memberCount[class.ID]; !loaded {
				n, countErr := s.members.CountByClass(ctx, class.ID)
				if countErr != nil {
					results[r.RowNum] = RowResult{Status: RowError, Message: "internal error"}
					failed++
					progress()
					continue
				}
				memberCount[class.ID] = n
			}
			if memberCount[class.ID] >= int64(limit) {
				results[r.RowNum] = RowResult{Status: RowError, Message: "class is full"}
				failed++
				progress()
				continue
			}
		}
		switch enrollErr := s.members.Create(ctx, &domain.ClassMember{ClassID: class.ID, UserID: userID}); {
		case enrollErr == nil:
			results[r.RowNum] = RowResult{Status: RowCreated}
			memberCount[class.ID]++
			created++
		case errors.Is(enrollErr, domain.ErrConflict):
			results[r.RowNum] = RowResult{Status: RowSkipped, Message: "already a member"}
			skipped++
		default:
			s.logger.Error("import enroll failed", "job_id", job.ID.String(), "row", r.RowNum, "error", enrollErr)
			results[r.RowNum] = RowResult{Status: RowError, Message: "internal error"}
			failed++
		}
		progress()
	}

	resultFile, err := BuildClassMembersResult(rows, results)
	if err != nil {
		return s.fail(ctx, job, "could not build result file")
	}
	if err := s.results.Set(ctx, job.ID, resultFile); err != nil {
		s.logger.Error("import result store failed", "job_id", job.ID.String(), "error", err)
	}

	job.Status = domain.ImportStatusCompleted
	job.ProcessedRows = processed
	job.CreatedCount = created
	job.SkippedCount = skipped
	job.FailedCount = failed
	return s.repo.Update(ctx, job)
}
