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
	Organizations       int
	Users               int
	Roles               int
	QuestionBanks       int
	Questions           int
	Classes             int
	ClassMembers        int
	ClassSessions       int
	Quizzes             int
	QuizRules           int
	QuizRooms           int
	QuizSubmissions     int
	Polls               int
	PollAnswers         int
	Chats               int
	ChatMembers         int
	Messages            int
	MessageReactions    int
	Media               int
	LiveRooms           int
	LiveParticipants    int
	LiveRecordings      int
	PracticeRooms       int
	PracticeSubmissions int
	OfflineRooms        int
	OfflineRoomViews    int
	Attendances         int
	GradebookColumns    int
	GradebookCells      int
}

func truncateAll(db *gorm.DB, ctx context.Context) error {
	tables := []string{
		"gradebook_cells",
		"gradebook_columns",
		"attendances",
		"offline_room_views",
		"offline_rooms",
		"practice_submissions",
		"practice_rooms",
		"message_reactions",
		"messages",
		"chat_members",
		"chats",
		"poll_answers",
		"polls",
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

	// 2. Permissions
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

	// 3. Preset roles (Manager + Teacher) — global, no org
	presetDefs := []struct {
		Name  string
		Perms []domain.PermissionName
	}{
		{domain.PresetRoleManager, domain.ManagerPermissions},
		{domain.PresetRoleTeacher, domain.TeacherPermissions},
		{domain.PresetRoleStudent, domain.StudentPermissions},
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

	// 4. Users
	type orgUsers struct {
		teachers []*domain.User
		students []*domain.User
	}
	usersByOrg := make(map[uuid.UUID]*orgUsers)

	// Fixed admin1 — global admin, no org
	admin := factory.NewUser(uuid.Nil, func(u *domain.User) {
		u.OrganizationID = nil
		u.Username = "admin1"
		u.Name = "Admin User"
		u.IsAdmin = true
	})
	if err := db.WithContext(ctx).Create(admin).Error; err != nil {
		return nil, fmt.Errorf("creating admin1: %w", err)
	}
	counts.Users++

	for i, org := range orgs {
		ou := &orgUsers{}

		// First org: manager1 = manager (acts as teacher).
		if i == 0 {
			staff := factory.NewUser(org.ID, func(u *domain.User) {
				u.OrganizationID = &org.ID
				u.Username = "manager1"
				u.Name = "Manager User"
				u.RoleID = &presetRoles[domain.PresetRoleManager].ID
			})
			if err := db.WithContext(ctx).Create(staff).Error; err != nil {
				return nil, fmt.Errorf("creating manager1: %w", err)
			}
			ou.teachers = append(ou.teachers, staff)
			counts.Users++

			// Fixed user1 — debug student in Zoora Demo org
			studentRole := presetRoles[domain.PresetRoleStudent]
			user1 := factory.NewUser(org.ID, func(u *domain.User) {
				u.OrganizationID = &org.ID
				u.Username = "user1"
				u.Name = "User One"
				u.RoleID = &studentRole.ID
			})
			if err := db.WithContext(ctx).Create(user1).Error; err != nil {
				return nil, fmt.Errorf("creating user1: %w", err)
			}
			ou.students = append(ou.students, user1)
			counts.Users++
		}

		// One extra teacher per org with Teacher preset role
		teacher := factory.NewUser(org.ID, func(u *domain.User) {
			u.OrganizationID = &org.ID
			u.RoleID = &presetRoles[domain.PresetRoleTeacher].ID
		})
		if err := db.WithContext(ctx).Create(teacher).Error; err != nil {
			return nil, fmt.Errorf("creating teacher: %w", err)
		}
		ou.teachers = append(ou.teachers, teacher)
		counts.Users++

		// 4 students per org
		studentRole := presetRoles[domain.PresetRoleStudent]
		for s := 0; s < 4; s++ {
			st := factory.NewUser(org.ID, func(u *domain.User) {
				u.OrganizationID = &org.ID
				u.RoleID = &studentRole.ID
			})
			if err := db.WithContext(ctx).Create(st).Error; err != nil {
				return nil, fmt.Errorf("creating student: %w", err)
			}
			ou.students = append(ou.students, st)
			counts.Users++
		}
		usersByOrg[org.ID] = ou
	}

	// 6. Media — avatars for each user
	allUsers := []*domain.User{admin}
	for _, org := range orgs {
		ou := usersByOrg[org.ID]
		allUsers = append(allUsers, ou.teachers...)
		allUsers = append(allUsers, ou.students...)
	}
	for _, u := range allUsers {
		m := factory.NewMedia(func(m *domain.Media) {
			m.ModelType = "user"
			m.ModelID = u.ID
			m.CollectionName = "avatar"
		})
		if err := db.WithContext(ctx).Create(m).Error; err != nil {
			return nil, fmt.Errorf("creating media: %w", err)
		}
		counts.Media++
	}

	// 7. QuestionBanks + Questions
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

			questionTypes := []domain.QuestionType{
				domain.QuestionTypeChoice, domain.QuestionTypeChoice, domain.QuestionTypeChoice,
				domain.QuestionTypeShortAnswer, domain.QuestionTypeShortAnswer,
				domain.QuestionTypeDescriptive,
				domain.QuestionTypeChoice,
			}
			var bankQuestions []*domain.Question
			for i, qt := range questionTypes {
				q := factory.NewQuestion(bank.ID, org.ID, func(q *domain.Question) {
					q.Type = qt
				})
				if err := db.WithContext(ctx).Create(q).Error; err != nil {
					return nil, fmt.Errorf("creating question: %w", err)
				}
				if i == 0 {
					photo := factory.NewMedia(func(m *domain.Media) {
						m.ModelType = domain.QuestionMediaModelType
						m.ModelID = q.ID
						m.CollectionName = domain.QuestionPhotosCollection
						m.MimeType = "image/png"
						m.FileName = "diagram.png"
					})
					if err := db.WithContext(ctx).Create(photo).Error; err != nil {
						return nil, fmt.Errorf("creating question photo: %w", err)
					}
					counts.Media++
					q.Metadata = []domain.QuestionMetadata{{Type: domain.QuestionMetadataPhoto, MediaID: photo.ID}}
					if err := db.WithContext(ctx).Save(q).Error; err != nil {
						return nil, fmt.Errorf("updating question metadata: %w", err)
					}
				}
				bankQuestions = append(bankQuestions, q)
				counts.Questions++
			}
			bd.questions = append(bd.questions, bankQuestions)
		}
		banksByOrg[org.ID] = bd
	}

	// 8. Classes per org
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
			var sessions []*domain.ClassSession
			for i := 0; i < 3; i++ {
				s := factory.NewClassSession(class.ID, func(s *domain.ClassSession) {
					s.StartTime = time.Now().Add(time.Duration(counts.ClassSessions+1) * 24 * time.Hour)
				})
				if err := db.WithContext(ctx).Create(s).Error; err != nil {
					return nil, fmt.Errorf("creating class session: %w", err)
				}
				sessions = append(sessions, s)
				counts.ClassSessions++
			}
			quizSession := sessions[1]
			liveSession := sessions[0]
			practiceSession := sessions[2]

			// 11. Quiz
			quiz := factory.NewQuiz(org.ID, teacher.ID, class.ID)
			if err := db.WithContext(ctx).Create(quiz).Error; err != nil {
				return nil, fmt.Errorf("creating quiz: %w", err)
			}
			counts.Quizzes++

			// 12. QuizRule
			bank := bd.banks[c%len(bd.banks)]
			bankQuestions := bd.questions[c%len(bd.questions)]
			questionIDs := make([]uuid.UUID, 0, len(bankQuestions))
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

			var quizTotal float64
			for _, q := range bankQuestions {
				quizTotal += q.MaxScore()
			}
			quiz.TotalScore = quizTotal
			if err := db.WithContext(ctx).Save(quiz).Error; err != nil {
				return nil, fmt.Errorf("updating quiz total_score: %w", err)
			}

			// 13. QuizRoom
			room := factory.NewQuizRoom(quiz.ID, quizSession.ID)
			if err := db.WithContext(ctx).Create(room).Error; err != nil {
				return nil, fmt.Errorf("creating quiz room: %w", err)
			}
			counts.QuizRooms++

			// 14. QuizSubmissions
			subCount := min(2, len(ou.students))
			for s := 0; s < subCount; s++ {
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

			// 15. Poll attached to class
			poll := factory.NewPoll(teacher.ID, "class", class.ID, func(p *domain.Poll) {
				p.Name = "Class feedback"
				p.AllowedAnswersCount = 1
				p.Options = []domain.PollOption{
					{Label: "Great", Value: "great"},
					{Label: "Okay", Value: "okay"},
					{Label: "Bad", Value: "bad"},
				}
			})
			if err := db.WithContext(ctx).Create(poll).Error; err != nil {
				return nil, fmt.Errorf("creating poll: %w", err)
			}
			counts.Polls++
			for _, student := range ou.students {
				ans := factory.NewPollAnswer(student.ID, poll.ID, poll.Options[int(student.ID.ID())%len(poll.Options)].Value)
				if err := db.WithContext(ctx).Create(ans).Error; err != nil {
					return nil, fmt.Errorf("creating poll answer: %w", err)
				}
				counts.PollAnswers++
			}

			// 16. Chat for class + members + messages + reactions
			chat := factory.NewChat("class", class.ID, func(c *domain.Chat) {
				c.Name = class.Name + " Chat"
			})
			if err := db.WithContext(ctx).Create(chat).Error; err != nil {
				return nil, fmt.Errorf("creating chat: %w", err)
			}
			counts.Chats++
			// teacher as admin member
			teacherMember := factory.NewChatMember(chat.ID, teacher.ID, domain.ChatMemberRoleAdmin)
			if err := db.WithContext(ctx).Create(teacherMember).Error; err != nil {
				return nil, fmt.Errorf("creating chat member: %w", err)
			}
			counts.ChatMembers++
			for _, student := range ou.students {
				cm := factory.NewChatMember(chat.ID, student.ID, domain.ChatMemberRoleMember)
				if err := db.WithContext(ctx).Create(cm).Error; err != nil {
					return nil, fmt.Errorf("creating chat member: %w", err)
				}
				counts.ChatMembers++
			}
			// messages
			tid := teacher.ID
			welcome := factory.NewMessage(chat.ID, &tid, func(m *domain.Message) {
				m.Content = "Welcome to the class!"
			})
			if err := db.WithContext(ctx).Create(welcome).Error; err != nil {
				return nil, fmt.Errorf("creating message: %w", err)
			}
			counts.Messages++
			for _, student := range ou.students {
				sid := student.ID
				reply := factory.NewMessage(chat.ID, &sid)
				if err := db.WithContext(ctx).Create(reply).Error; err != nil {
					return nil, fmt.Errorf("creating message: %w", err)
				}
				counts.Messages++
			}
			// reactions on welcome message
			for _, student := range ou.students {
				r := factory.NewMessageReaction(welcome.ID, student.ID, "👍")
				if err := db.WithContext(ctx).Create(r).Error; err != nil {
					return nil, fmt.Errorf("creating reaction: %w", err)
				}
				counts.MessageReactions++
			}

			// 17. LiveRoom on liveSession + participants + recording
			liveRoom := factory.NewLiveRoom(liveSession.ID, func(lr *domain.LiveRoom) {
				lr.Status = domain.LiveRoomStatusFinished
				start := time.Now().Add(-2 * time.Hour)
				end := time.Now().Add(-1 * time.Hour)
				lr.ActualStartTime = &start
				lr.ActualEndTime = &end
			})
			if err := db.WithContext(ctx).Create(liveRoom).Error; err != nil {
				return nil, fmt.Errorf("creating live room: %w", err)
			}
			counts.LiveRooms++
			// teacher + all students as participants
			lpUsers := append([]*domain.User{teacher}, ou.students...)
			for _, u := range lpUsers {
				lp := factory.NewLiveParticipant(liveRoom.ID, u.ID)
				if err := db.WithContext(ctx).Create(lp).Error; err != nil {
					return nil, fmt.Errorf("creating live participant: %w", err)
				}
				counts.LiveParticipants++
			}
			rec := factory.NewLiveRecording(liveRoom.ID)
			if err := db.WithContext(ctx).Create(rec).Error; err != nil {
				return nil, fmt.Errorf("creating live recording: %w", err)
			}
			counts.LiveRecordings++

			// 17b. A scheduled (not-yet-started) live room so the lobby's
			// host-start / student-wait + countdown flow has seed data.
			scheduledRoom := factory.NewLiveRoom(liveSession.ID, func(lr *domain.LiveRoom) {
				lr.Name = "Scheduled session"
				lr.Status = domain.LiveRoomStatusCreated
				at := time.Now().Add(24 * time.Hour)
				lr.ScheduledStartTime = &at
			})
			if err := db.WithContext(ctx).Create(scheduledRoom).Error; err != nil {
				return nil, fmt.Errorf("creating scheduled live room: %w", err)
			}
			counts.LiveRooms++

			// 18. PracticeRoom + submissions
			pr := factory.NewPracticeRoom(org.ID, class.ID, practiceSession.ID, teacher.ID)
			if err := db.WithContext(ctx).Create(pr).Error; err != nil {
				return nil, fmt.Errorf("creating practice room: %w", err)
			}
			counts.PracticeRooms++
			for _, student := range ou.students {
				ps := factory.NewPracticeSubmission(pr.ID, student.ID)
				if err := db.WithContext(ctx).Create(ps).Error; err != nil {
					return nil, fmt.Errorf("creating practice submission: %w", err)
				}
				counts.PracticeSubmissions++
			}

			// 19. OfflineRoom + views
			or := factory.NewOfflineRoom(org.ID, class.ID, liveSession.ID, teacher.ID)
			if err := db.WithContext(ctx).Create(or).Error; err != nil {
				return nil, fmt.Errorf("creating offline room: %w", err)
			}
			counts.OfflineRooms++
			for _, student := range ou.students {
				v := factory.NewOfflineRoomView(or.ID, student.ID)
				if err := db.WithContext(ctx).Create(v).Error; err != nil {
					return nil, fmt.Errorf("creating offline room view: %w", err)
				}
				counts.OfflineRoomViews++
			}

			// 20. Attendance — for each session, each student
			statuses := []domain.AttendanceStatus{
				domain.AttendanceStatusPresent,
				domain.AttendanceStatusLate,
				domain.AttendanceStatusAbsent,
				domain.AttendanceStatusExcused,
			}
			for sIdx, sess := range sessions {
				for stIdx, student := range ou.students {
					a := factory.NewAttendance(org.ID, class.ID, sess.ID, student.ID, func(a *domain.Attendance) {
						a.Status = statuses[(sIdx+stIdx)%len(statuses)]
					})
					if err := db.WithContext(ctx).Create(a).Error; err != nil {
						return nil, fmt.Errorf("creating attendance: %w", err)
					}
					counts.Attendances++
				}
			}

			// 21. Gradebook columns + cells
			colDefs := []struct {
				Title string
				Type  domain.GradebookColumnType
			}{
				{"Attendance", domain.GradebookColumnAutoAttendance},
				{"Quiz Score", domain.GradebookColumnAutoQuiz},
				{"Midterm", domain.GradebookColumnManualGrade},
				{"Notes", domain.GradebookColumnManualText},
			}
			for idx, cd := range colDefs {
				col := factory.NewGradebookColumn(class.ID, cd.Type, func(c *domain.GradebookColumn) {
					c.Title = cd.Title
					c.OrderIndex = idx
					if cd.Type == domain.GradebookColumnAutoQuiz {
						qid := quiz.ID
						c.SourceID = &qid
					}
				})
				if err := db.WithContext(ctx).Create(col).Error; err != nil {
					return nil, fmt.Errorf("creating gradebook column: %w", err)
				}
				counts.GradebookColumns++

				for _, student := range ou.students {
					value := fmt.Sprintf("%d", 70+int(student.ID.ID())%30)
					if cd.Type == domain.GradebookColumnManualText {
						value = "OK"
					}
					cell := factory.NewGradebookCell(col.ID, student.ID, value)
					if err := db.WithContext(ctx).Create(cell).Error; err != nil {
						return nil, fmt.Errorf("creating gradebook cell: %w", err)
					}
					counts.GradebookCells++
				}
			}
		}
	}

	return counts, nil
}

func printSummary(c *seedCounts) {
	fmt.Println("\nSeeded successfully:")
	fmt.Printf("  Organizations:        %d\n", c.Organizations)
	fmt.Printf("  Users:                %d\n", c.Users)
	fmt.Printf("  Roles:                %d\n", c.Roles)
	fmt.Printf("  Media:                %d\n", c.Media)
	fmt.Printf("  QuestionBanks:        %d\n", c.QuestionBanks)
	fmt.Printf("  Questions:            %d\n", c.Questions)
	fmt.Printf("  Classes:              %d\n", c.Classes)
	fmt.Printf("  ClassMembers:         %d\n", c.ClassMembers)
	fmt.Printf("  ClassSessions:        %d\n", c.ClassSessions)
	fmt.Printf("  Quizzes:              %d\n", c.Quizzes)
	fmt.Printf("  QuizRules:            %d\n", c.QuizRules)
	fmt.Printf("  QuizRooms:            %d\n", c.QuizRooms)
	fmt.Printf("  QuizSubmissions:      %d\n", c.QuizSubmissions)
	fmt.Printf("  Polls:                %d\n", c.Polls)
	fmt.Printf("  PollAnswers:          %d\n", c.PollAnswers)
	fmt.Printf("  Chats:                %d\n", c.Chats)
	fmt.Printf("  ChatMembers:          %d\n", c.ChatMembers)
	fmt.Printf("  Messages:             %d\n", c.Messages)
	fmt.Printf("  MessageReactions:     %d\n", c.MessageReactions)
	fmt.Printf("  LiveRooms:            %d\n", c.LiveRooms)
	fmt.Printf("  LiveParticipants:     %d\n", c.LiveParticipants)
	fmt.Printf("  LiveRecordings:       %d\n", c.LiveRecordings)
	fmt.Printf("  PracticeRooms:        %d\n", c.PracticeRooms)
	fmt.Printf("  PracticeSubmissions:  %d\n", c.PracticeSubmissions)
	fmt.Printf("  OfflineRooms:         %d\n", c.OfflineRooms)
	fmt.Printf("  OfflineRoomViews:     %d\n", c.OfflineRoomViews)
	fmt.Printf("  Attendances:          %d\n", c.Attendances)
	fmt.Printf("  GradebookColumns:     %d\n", c.GradebookColumns)
	fmt.Printf("  GradebookCells:       %d\n", c.GradebookCells)
	fmt.Println("\nLogins:")
	fmt.Println("  admin1 / password   (super admin)")
	fmt.Println("  manager1 / password (Manager preset in Zoora Demo org)")
	fmt.Println("  user1 / password    (Student in Zoora Demo org)")
}
