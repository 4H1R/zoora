package imports

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func (s *service) processClasses(ctx context.Context, job *domain.ImportJob, data []byte) error {
	classRows, memberRows, err := ParseClassesFile(data)
	if err != nil {
		return s.fail(ctx, job, err.Error())
	}
	job.TotalRows = len(classRows) + len(memberRows)
	if err := s.repo.Update(ctx, job); err != nil {
		return err
	}

	// resolve every referenced username in one query
	usernameSet := map[string]bool{}
	for i := range classRows {
		classRows[i].OwnerUsername = strings.ToLower(strings.TrimSpace(classRows[i].OwnerUsername))
		if classRows[i].OwnerUsername != "" {
			usernameSet[classRows[i].OwnerUsername] = true
		}
	}
	for i := range memberRows {
		memberRows[i].MemberUsername = strings.ToLower(strings.TrimSpace(memberRows[i].MemberUsername))
		if memberRows[i].MemberUsername != "" {
			usernameSet[memberRows[i].MemberUsername] = true
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

	names := make([]string, 0, len(classRows))
	for _, r := range classRows {
		if r.Name != "" {
			names = append(names, r.Name)
		}
	}
	existingClasses, err := s.classes.ListByNames(ctx, job.OrganizationID, names)
	if err != nil {
		return s.fail(ctx, job, "could not check existing classes")
	}
	existingByName := map[string][]domain.Class{}
	for _, c := range existingClasses {
		existingByName[c.Name] = append(existingByName[c.Name], c)
	}

	classResults := make(map[int]RowResult, len(classRows))
	memberResults := make(map[int]RowResult, len(memberRows))
	classIDByName := map[string]uuid.UUID{} // classes member rows may target
	capacityByID := map[uuid.UUID]int{}
	memberCount := map[uuid.UUID]int64{} // lazily loaded per capacity-limited class
	seenClass := map[string]bool{}
	created, skipped, failed, processed := 0, 0, 0, 0

	progress := func() {
		processed++
		if processed%progressBatch == 0 {
			_ = s.repo.UpdateProgress(ctx, job.ID, processed, created, skipped, failed)
		}
	}

	for _, r := range classRows {
		msg, capacity := validateClassRow(r)
		if msg != "" {
			classResults[r.RowNum] = RowResult{Status: RowError, Message: msg}
			failed++
			progress()
			continue
		}
		if seenClass[r.Name] {
			classResults[r.RowNum] = RowResult{Status: RowError, Message: "duplicate class_name in file"}
			failed++
			progress()
			continue
		}
		seenClass[r.Name] = true

		switch ex := existingByName[r.Name]; {
		case len(ex) == 1:
			// reuse: existing owner kept, sheet description/capacity ignored —
			// import never mutates existing records
			classIDByName[r.Name] = ex[0].ID
			capacityByID[ex[0].ID] = ex[0].TotalUsers
			classResults[r.RowNum] = RowResult{Status: RowSkipped, Message: "class already exists; members will be enrolled"}
			skipped++
		case len(ex) > 1:
			classResults[r.RowNum] = RowResult{Status: RowError, Message: "multiple classes with this name already exist"}
			failed++
		default:
			ownerID, ok := userIDByName[r.OwnerUsername]
			if !ok {
				classResults[r.RowNum] = RowResult{Status: RowError, Message: "owner username not found"}
				failed++
				progress()
				continue
			}
			class := &domain.Class{
				OrganizationID: job.OrganizationID,
				UserID:         ownerID,
				Name:           r.Name,
				Description:    r.Description,
				TotalUsers:     capacity,
			}
			if createErr := s.classes.Create(ctx, class); createErr != nil {
				s.logger.Error("import class create failed", "job_id", job.ID.String(), "row", r.RowNum, "error", createErr)
				classResults[r.RowNum] = RowResult{Status: RowError, Message: "internal error"}
				failed++
			} else {
				classIDByName[r.Name] = class.ID
				capacityByID[class.ID] = capacity
				classResults[r.RowNum] = RowResult{Status: RowCreated}
				created++
			}
		}
		progress()
	}

	for _, r := range memberRows {
		if r.ClassName == "" || r.MemberUsername == "" {
			memberResults[r.RowNum] = RowResult{Status: RowError, Message: "class_name and member_username are required"}
			failed++
			progress()
			continue
		}
		classID, ok := classIDByName[r.ClassName]
		if !ok {
			memberResults[r.RowNum] = RowResult{Status: RowError, Message: "class not found in Classes sheet (or its row failed)"}
			failed++
			progress()
			continue
		}
		userID, ok := userIDByName[r.MemberUsername]
		if !ok {
			memberResults[r.RowNum] = RowResult{Status: RowError, Message: "member username not found"}
			failed++
			progress()
			continue
		}
		if limit := capacityByID[classID]; limit > 0 {
			if _, loaded := memberCount[classID]; !loaded {
				n, countErr := s.members.CountByClass(ctx, classID)
				if countErr != nil {
					memberResults[r.RowNum] = RowResult{Status: RowError, Message: "internal error"}
					failed++
					progress()
					continue
				}
				memberCount[classID] = n
			}
			if memberCount[classID] >= int64(limit) {
				memberResults[r.RowNum] = RowResult{Status: RowError, Message: "class is full"}
				failed++
				progress()
				continue
			}
		}
		switch enrollErr := s.members.Create(ctx, &domain.ClassMember{ClassID: classID, UserID: userID}); {
		case enrollErr == nil:
			memberResults[r.RowNum] = RowResult{Status: RowCreated}
			memberCount[classID]++
			created++
		case errors.Is(enrollErr, domain.ErrConflict):
			memberResults[r.RowNum] = RowResult{Status: RowSkipped, Message: "already a member"}
			skipped++
		default:
			s.logger.Error("import enroll failed", "job_id", job.ID.String(), "row", r.RowNum, "error", enrollErr)
			memberResults[r.RowNum] = RowResult{Status: RowError, Message: "internal error"}
			failed++
		}
		progress()
	}

	resultFile, err := BuildClassesResult(classRows, classResults, memberRows, memberResults)
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
	return s.complete(ctx, job)
}

func validateClassRow(r ClassRow) (string, int) {
	if len([]rune(strings.TrimSpace(r.Name))) < 2 {
		return "class_name must be at least 2 characters", 0
	}
	if r.OwnerUsername == "" {
		return "owner_username is required", 0
	}
	capacity := 0
	if r.Capacity != "" {
		n, err := strconv.Atoi(r.Capacity)
		if err != nil || n < 0 {
			return "capacity must be a whole number of 0 or more", 0
		}
		capacity = n
	}
	return "", capacity
}
