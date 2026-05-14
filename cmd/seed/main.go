package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/factory"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}

	ctx := context.Background()

	if err := truncateAll(db, ctx); err != nil {
		log.Fatalf("truncating tables: %v", err)
	}

	counts, err := seedAll(db, ctx)
	if err != nil {
		log.Fatalf("seeding: %v", err)
	}

	printSummary(counts)
}

type seedCounts struct {
	Organizations   int
	Users           int
	Roles           int
	QuestionBanks   int
	Questions       int
	Classes         int
	ClassMembers    int
	ClassSessions   int
	Quizzes         int
	QuizRules       int
	QuizRooms       int
	QuizSubmissions int
}

func truncateAll(db *gorm.DB, ctx context.Context) error {
	tables := []string{
		"quiz_submissions",
		"quiz_rooms",
		"quiz_rules",
		"quizzes",
		"class_sessions",
		"class_members",
		"classes",
		"questions",
		"question_banks",
		"role_permissions",
		"roles",
		"live_recordings",
		"live_participants",
		"live_rooms",
		"media",
		"users",
		"organizations",
		"permissions",
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, t := range tables {
			if err := tx.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", t)).Error; err != nil {
				return fmt.Errorf("truncating %s: %w", t, err)
			}
		}
		return nil
	})
}

func seedAll(db *gorm.DB, ctx context.Context) (*seedCounts, error) {
	counts := &seedCounts{}

	// 1. Organizations
	demoOrg := factory.NewOrganization(func(o *domain.Organization) {
		o.Name = "Zoora Demo"
		o.Description = "Demo organization for development"
	})
	randomOrg := factory.NewOrganization()
	orgs := []*domain.Organization{demoOrg, randomOrg}
	for _, org := range orgs {
		if err := db.WithContext(ctx).Create(org).Error; err != nil {
			return nil, fmt.Errorf("creating organization %s: %w", org.Name, err)
		}
	}
	counts.Organizations = len(orgs)

	// 2. Users (5 per org — first org gets fixed admin)
	type orgUsers struct {
		teachers []*domain.User
		students []*domain.User
	}
	usersByOrg := make(map[uuid.UUID]*orgUsers)

	for i, org := range orgs {
		ou := &orgUsers{}
		usersToCreate := 5

		for j := 0; j < usersToCreate; j++ {
			var u *domain.User
			if i == 0 && j == 0 {
				u = factory.NewUser(org.ID, func(u *domain.User) {
					u.OrganizationID = nil
					u.Username = "admin1"
					u.Name = "Admin User 1"
					u.IsAdmin = true
				})
			} else if i == 0 && j == 1 {
				u = factory.NewUser(org.ID, func(u *domain.User) {
					u.OrganizationID = nil
					u.Username = "admin2"
					u.Name = "Admin User 2"
					u.IsAdmin = true
				})
			} else if i == 1 && j == 0 {
				u = factory.NewUser(org.ID, func(u *domain.User) {
					u.OrganizationID = &org.ID
					u.Username = "staff1"
					u.Name = "Staff User 1"
				})
			} else if i == 1 && j == 1 {
				u = factory.NewUser(org.ID, func(u *domain.User) {
					u.OrganizationID = &org.ID
					u.Username = "staff2"
					u.Name = "Staff User 2"
				})
			} else {
				u = factory.NewUser(org.ID)
			}
			if err := db.WithContext(ctx).Create(u).Error; err != nil {
				return nil, fmt.Errorf("creating user %s: %w", u.Username, err)
			}
			if j < 2 {
				ou.teachers = append(ou.teachers, u)
			} else {
				ou.students = append(ou.students, u)
			}
			counts.Users++
		}
		usersByOrg[org.ID] = ou
	}

	// 3. Permissions — seed from domain.AllPermissions
	permissions := make([]domain.Permission, 0, len(domain.AllPermissions))
	for _, name := range domain.AllPermissions {
		permissions = append(permissions, domain.Permission{Name: name})
	}
	if err := db.WithContext(ctx).Create(&permissions).Error; err != nil {
		return nil, fmt.Errorf("seeding permissions: %w", err)
	}

	permByName := make(map[domain.PermissionName]uuid.UUID, len(permissions))
	for _, p := range permissions {
		permByName[p.Name] = p.ID
	}

	resolvePermIDs := func(names []domain.PermissionName) []uuid.UUID {
		ids := make([]uuid.UUID, 0, len(names))
		for _, n := range names {
			if id, ok := permByName[n]; ok {
				ids = append(ids, id)
			}
		}
		return ids
	}

	assignPerms := func(roleID uuid.UUID, permIDs []uuid.UUID) error {
		for _, pid := range permIDs {
			rp := domain.RolePermission{RoleID: roleID, PermissionID: pid}
			if err := db.WithContext(ctx).Create(&rp).Error; err != nil {
				return fmt.Errorf("assigning permission: %w", err)
			}
		}
		return nil
	}

	// 4. Preset roles (Staff + Teacher) — global, no org
	presetDefs := []struct {
		Name  string
		Perms []domain.PermissionName
	}{
		{domain.PresetRoleStaff, domain.StaffPermissions},
		{domain.PresetRoleTeacher, domain.TeacherPermissions},
	}

	presetRoles := make(map[string]*domain.Role)
	for _, def := range presetDefs {
		role := &domain.Role{Name: def.Name, IsPreset: true}
		if err := db.WithContext(ctx).Create(role).Error; err != nil {
			return nil, fmt.Errorf("creating preset role %s: %w", def.Name, err)
		}
		if err := assignPerms(role.ID, resolvePermIDs(def.Perms)); err != nil {
			return nil, fmt.Errorf("assigning preset %s permissions: %w", def.Name, err)
		}
		presetRoles[def.Name] = role
		counts.Roles++
	}

	// 5. Org-scoped roles (Student per org) + assign roles to users
	studentPermNames := []domain.PermissionName{
		domain.PermLiveSessionsView, domain.PermLiveSessionsJoin,
		domain.PermRecordingsView,
		domain.PermClassesView, domain.PermClassesJoin,
		domain.PermUsersView,
		domain.PermQuizzesView,
		domain.PermPollsView,
	}
	studentPermIDs := resolvePermIDs(studentPermNames)

	for _, org := range orgs {
		studentRole := factory.NewRole(&org.ID, func(r *domain.Role) {
			r.Name = "Student"
		})
		if err := db.WithContext(ctx).Create(studentRole).Error; err != nil {
			return nil, fmt.Errorf("creating student role: %w", err)
		}
		if err := assignPerms(studentRole.ID, studentPermIDs); err != nil {
			return nil, fmt.Errorf("assigning student permissions: %w", err)
		}
		counts.Roles++

		ou := usersByOrg[org.ID]
		for _, teacher := range ou.teachers {
			if teacher.IsAdmin {
				continue
			}
			teacher.RoleID = &presetRoles[domain.PresetRoleStaff].ID
			if err := db.WithContext(ctx).Save(teacher).Error; err != nil {
				return nil, fmt.Errorf("assigning staff role: %w", err)
			}
		}
		for _, student := range ou.students {
			student.RoleID = &studentRole.ID
			if err := db.WithContext(ctx).Save(student).Error; err != nil {
				return nil, fmt.Errorf("assigning student role: %w", err)
			}
		}
	}

	// 6. QuestionBanks (2 per org)
	type orgBankData struct {
		banks     []*domain.QuestionBank
		questions [][]*domain.Question
	}
	banksByOrg := make(map[uuid.UUID]*orgBankData)

	for _, org := range orgs {
		bd := &orgBankData{}
		for b := 0; b < 2; b++ {
			bank := factory.NewQuestionBank(org.ID)
			if err := db.WithContext(ctx).Create(bank).Error; err != nil {
				return nil, fmt.Errorf("creating question bank: %w", err)
			}
			counts.QuestionBanks++
			bd.banks = append(bd.banks, bank)

			// 7. Questions (7 per bank)
			var bankQuestions []*domain.Question
			questionTypes := []domain.QuestionType{
				domain.QuestionTypeChoice,
				domain.QuestionTypeChoice,
				domain.QuestionTypeChoice,
				domain.QuestionTypeShortAnswer,
				domain.QuestionTypeShortAnswer,
				domain.QuestionTypeDescriptive,
				domain.QuestionTypeChoice,
			}
			for _, qt := range questionTypes {
				q := factory.NewQuestion(bank.ID, org.ID, func(q *domain.Question) {
					q.Type = qt
				})
				if err := db.WithContext(ctx).Create(q).Error; err != nil {
					return nil, fmt.Errorf("creating question: %w", err)
				}
				bankQuestions = append(bankQuestions, q)
				counts.Questions++
			}
			bd.questions = append(bd.questions, bankQuestions)
		}
		banksByOrg[org.ID] = bd
	}

	// 8. Classes (2 per org, assigned to first teacher)
	for _, org := range orgs {
		ou := usersByOrg[org.ID]
		bd := banksByOrg[org.ID]

		for c := 0; c < 2; c++ {
			teacher := ou.teachers[c%len(ou.teachers)]
			class := factory.NewClass(org.ID, teacher.ID, func(cl *domain.Class) {
				cl.TotalUsers = 30
			})
			if err := db.WithContext(ctx).Create(class).Error; err != nil {
				return nil, fmt.Errorf("creating class: %w", err)
			}
			counts.Classes++

			// 9. ClassMembers — enroll all students
			for _, student := range ou.students {
				m := factory.NewClassMember(class.ID, student.ID)
				if err := db.WithContext(ctx).Create(m).Error; err != nil {
					return nil, fmt.Errorf("creating class member: %w", err)
				}
				counts.ClassMembers++
			}

			// 10. ClassSessions (3 per class)
			var quizSession *domain.ClassSession
			for i := 0; i < 3; i++ {
				s := factory.NewClassSession(class.ID, func(s *domain.ClassSession) {
					s.StartTime = time.Now().Add(time.Duration(counts.ClassSessions+1) * 24 * time.Hour)
				})
				if err := db.WithContext(ctx).Create(s).Error; err != nil {
					return nil, fmt.Errorf("creating class session: %w", err)
				}
				if i == 1 {
					quizSession = s
				}
				counts.ClassSessions++
			}

			// 11. Quiz (1 per class)
			quiz := factory.NewQuiz(org.ID, teacher.ID, class.ID)
			if err := db.WithContext(ctx).Create(quiz).Error; err != nil {
				return nil, fmt.Errorf("creating quiz: %w", err)
			}
			counts.Quizzes++

			// 12. QuizRules — 1 manual rule using questions from first bank
			bank := bd.banks[c%len(bd.banks)]
			bankQuestions := bd.questions[c%len(bd.questions)]
			var questionIDs []uuid.UUID
			for _, q := range bankQuestions {
				questionIDs = append(questionIDs, q.ID)
			}
			rule := factory.NewQuizRule(quiz.ID, func(r *domain.QuizRule) {
				r.Type = domain.QuizRuleTypeManual
				r.BankID = &bank.ID
				r.QuestionIDs = questionIDs
				r.Count = len(questionIDs)
			})
			if err := db.WithContext(ctx).Create(rule).Error; err != nil {
				return nil, fmt.Errorf("creating quiz rule: %w", err)
			}
			counts.QuizRules++

			// 13. QuizRoom (1 per quiz, linked to quiz session)
			if quizSession != nil {
				room := factory.NewQuizRoom(quiz.ID, quizSession.ID)
				if err := db.WithContext(ctx).Create(room).Error; err != nil {
					return nil, fmt.Errorf("creating quiz room: %w", err)
				}
				counts.QuizRooms++

				// 14. QuizSubmissions (1 per enrolled student, max 2)
				submissionCount := min(2, len(ou.students))
				for s := 0; s < submissionCount; s++ {
					student := ou.students[s]
					var answers []domain.SubmissionAnswer
					totalScore := 0.0
					for _, q := range bankQuestions {
						earned := 0.0
						var selectedIDs []string
						var value string
						if q.Type == domain.QuestionTypeChoice && len(q.Options) > 0 {
							selectedIDs = []string{q.Options[0].ID}
							earned = q.Options[0].Score
						} else if q.Type == domain.QuestionTypeShortAnswer {
							value = "Sample answer"
							earned = 2.0
						}
						totalScore += earned
						answers = append(answers, domain.SubmissionAnswer{
							QuestionID:        q.ID,
							SelectedOptionIDs: selectedIDs,
							Value:             value,
							EarnedScore:       earned,
							SpentSeconds:      60,
						})
					}
					sub := factory.NewQuizSubmission(quiz.ID, student.ID, func(qs *domain.QuizSubmission) {
						qs.Answers = answers
						qs.TotalScore = totalScore
						qs.Status = domain.SubmissionStatusGraded
					})
					if err := db.WithContext(ctx).Create(sub).Error; err != nil {
						return nil, fmt.Errorf("creating quiz submission: %w", err)
					}
					counts.QuizSubmissions++
				}
			}
		}
	}

	return counts, nil
}

func printSummary(c *seedCounts) {
	fmt.Println("\nSeeded successfully:")
	fmt.Printf("  Organizations:    %d\n", c.Organizations)
	fmt.Printf("  Users:            %d\n", c.Users)
	fmt.Printf("  Roles:            %d\n", c.Roles)
	fmt.Printf("  QuestionBanks:    %d\n", c.QuestionBanks)
	fmt.Printf("  Questions:        %d\n", c.Questions)
	fmt.Printf("  Classes:          %d\n", c.Classes)
	fmt.Printf("  ClassMembers:     %d\n", c.ClassMembers)
	fmt.Printf("  ClassSessions:    %d\n", c.ClassSessions)
	fmt.Printf("  Quizzes:          %d\n", c.Quizzes)
	fmt.Printf("  QuizRules:        %d\n", c.QuizRules)
	fmt.Printf("  QuizRooms:        %d\n", c.QuizRooms)
	fmt.Printf("  QuizSubmissions:  %d\n", c.QuizSubmissions)
	fmt.Println("\nAdmin logins:  username=admin1 password=password | username=admin2 password=password")
	fmt.Println("Staff logins:  username=staff1 password=password | username=staff2 password=password (Staff preset role)")
}
