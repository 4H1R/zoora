package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
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

	loc := chooseLocale()
	// Publish the resolved choice to the env the factory reads, so interactive
	// and SEED_LANG-driven runs both flow through the same config source.
	os.Setenv(factory.SeedLangEnv, string(loc))
	fmt.Printf("Seeding in %s.\n", localeLabel(loc))

	ctx := context.Background()

	if err := truncateAll(db, ctx); err != nil {
		log.Fatalf("truncating tables: %v", err)
	}

	if err := flushRedis(ctx); err != nil {
		log.Fatalf("flushing redis: %v", err)
	}

	counts, err := seedAll(db, ctx)
	if err != nil {
		log.Fatalf("seeding: %v", err)
	}

	printSummary(counts)
}

// flushRedis clears all cached state (permissions, tenant, sessions, queue) so a
// fresh seed isn't shadowed by stale cache from a previous dataset.
func flushRedis(ctx context.Context) error {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatal("REDIS_URL is required")
	}
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return fmt.Errorf("parsing redis URL: %w", err)
	}
	client := redis.NewClient(opts)
	defer client.Close()
	if err := client.FlushAll(ctx).Err(); err != nil {
		return fmt.Errorf("flushall: %w", err)
	}
	fmt.Println("Redis flushed.")
	return nil
}

// chooseLocale resolves the seed language. SEED_LANG (en/fa) takes priority for
// non-interactive runs; otherwise it prompts on stdin. Empty input (or EOF, as
// happens under `docker compose exec -T`) defaults to Persian.
func chooseLocale() factory.Locale {
	if v := strings.TrimSpace(os.Getenv("SEED_LANG")); v != "" {
		return parseLocale(v)
	}
	fmt.Print("Seed data language? [P]ersian (default) / [E]nglish: ")
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return parseLocale(line)
}

func parseLocale(s string) factory.Locale {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "e", "en", "eng", "english":
		return factory.LocaleEn
	default:
		return factory.LocaleFa
	}
}

func localeLabel(l factory.Locale) string {
	if l == factory.LocaleEn {
		return "English"
	}
	return "Persian"
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
		o.Name = factory.T("Zoora Demo", "زورا دمو")
		o.Slug = "acme"
		o.Description = factory.T("Demo organization for development", "سازمان نمونه برای توسعه")
	})
	randomOrg := factory.NewOrganization(func(o *domain.Organization) {
		o.Slug = "globex"
	})
	orgs := []*domain.Organization{demoOrg, randomOrg}
	for _, org := range orgs {
		if err := db.WithContext(ctx).Create(org).Error; err != nil {
			return nil, fmt.Errorf("creating organization %s: %w", org.Name, err)
		}
		if err := db.WithContext(ctx).Create(factory.NewOrganizationSettings(org.ID)).Error; err != nil {
			return nil, fmt.Errorf("creating org settings for %s: %w", org.Name, err)
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

	// 3. Preset roles — global, no org
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
		u.Name = factory.T("Admin User", "کاربر مدیر کل")
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
				u.Name = factory.T("Manager User", "کاربر مدیر")
				u.RoleID = &presetRoles[domain.PresetRoleManager].ID
			})
			if err := db.WithContext(ctx).Create(staff).Error; err != nil {
				return nil, fmt.Errorf("creating manager1: %w", err)
			}
			ou.teachers = append(ou.teachers, staff)
			counts.Users++

			// Second manager — manager2 (Manager preset) in acme org.
			staff2 := factory.NewUser(org.ID, func(u *domain.User) {
				u.OrganizationID = &org.ID
				u.Username = "manager2"
				u.Name = factory.T("Manager Two", "کاربر مدیر دو")
				u.RoleID = &presetRoles[domain.PresetRoleManager].ID
			})
			if err := db.WithContext(ctx).Create(staff2).Error; err != nil {
				return nil, fmt.Errorf("creating manager2: %w", err)
			}
			ou.teachers = append(ou.teachers, staff2)
			counts.Users++

			// Fixed user1 — debug student in Zoora Demo org
			studentRole := presetRoles[domain.PresetRoleStudent]
			user1 := factory.NewUser(org.ID, func(u *domain.User) {
				u.OrganizationID = &org.ID
				u.Username = "user1"
				u.Name = factory.T("User One", "کاربر یک")
				u.RoleID = &studentRole.ID
			})
			if err := db.WithContext(ctx).Create(user1).Error; err != nil {
				return nil, fmt.Errorf("creating user1: %w", err)
			}
			ou.students = append(ou.students, user1)
			counts.Users++

			// Fixed user2 — second debug student in Zoora Demo org
			user2 := factory.NewUser(org.ID, func(u *domain.User) {
				u.OrganizationID = &org.ID
				u.Username = "user2"
				u.Name = factory.T("User Two", "کاربر دو")
				u.RoleID = &studentRole.ID
			})
			if err := db.WithContext(ctx).Create(user2).Error; err != nil {
				return nil, fmt.Errorf("creating user2: %w", err)
			}
			ou.students = append(ou.students, user2)
			counts.Users++
		}

		teacher := factory.NewUser(org.ID, func(u *domain.User) {
			u.OrganizationID = &org.ID
			u.RoleID = &presetRoles[domain.PresetRoleTeacher].ID
		})
		if err := db.WithContext(ctx).Create(teacher).Error; err != nil {
			return nil, fmt.Errorf("creating teacher: %w", err)
		}
		ou.teachers = append(ou.teachers, teacher)
		counts.Users++

		studentRole := presetRoles[domain.PresetRoleStudent]
		for range 4 {
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

	// 6. Media — avatars for each user. Media objects are tenant-namespaced by
	// organization_id (S3 key prefix orgs/{org_id}/…), so carry each user's org
	// through to the media row. The platform admin belongs to no org (nil).
	type avatarOwner struct {
		user  *domain.User
		orgID *uuid.UUID
	}
	avatarOwners := []avatarOwner{{user: admin, orgID: nil}}
	for _, org := range orgs {
		orgID := org.ID
		ou := usersByOrg[org.ID]
		for _, t := range ou.teachers {
			avatarOwners = append(avatarOwners, avatarOwner{user: t, orgID: &orgID})
		}
		for _, st := range ou.students {
			avatarOwners = append(avatarOwners, avatarOwner{user: st, orgID: &orgID})
		}
	}
	for _, ao := range avatarOwners {
		m := factory.NewMedia(func(m *domain.Media) {
			m.OrganizationID = ao.orgID
			m.ModelType = "user"
			m.ModelID = ao.user.ID
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
		for range 2 {
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
					orgID := org.ID
					photo := factory.NewMedia(func(m *domain.Media) {
						m.OrganizationID = &orgID
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

					// Demo per-option image on the first choice option, backed by a
					// real option-photos media row.
					if q.Type == domain.QuestionTypeChoice && len(q.Options) > 0 {
						optPhoto := factory.NewMedia(func(m *domain.Media) {
							m.OrganizationID = &orgID
							m.ModelType = domain.QuestionMediaModelType
							m.ModelID = q.ID
							m.CollectionName = domain.QuestionOptionPhotosCollection
							m.MimeType = "image/png"
							m.FileName = "option.png"
						})
						if err := db.WithContext(ctx).Create(optPhoto).Error; err != nil {
							return nil, fmt.Errorf("creating option photo: %w", err)
						}
						counts.Media++
						q.Options[0].ImageMediaID = &optPhoto.ID
						// Demo negative marking for this choice question.
						q.NegativeMarkMode = domain.NegativeMarkPerWrong
						q.NegativeValue = domain.FractionFor(len(q.Options))
						q.WrongsPerPoint = 0
					}

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

		for c := range 2 {
			teacher := ou.teachers[c%len(ou.teachers)]
			class := factory.NewClass(org.ID, teacher.ID, func(cl *domain.Class) {
				cl.TotalUsers = 30
			})
			if err := db.WithContext(ctx).Create(class).Error; err != nil {
				return nil, fmt.Errorf("creating class: %w", err)
			}
			counts.Classes++

			// 9. ClassMembers
			for _, student := range ou.students {
				m := factory.NewClassMember(class.ID, student.ID)
				if err := db.WithContext(ctx).Create(m).Error; err != nil {
					return nil, fmt.Errorf("creating class member: %w", err)
				}
				counts.ClassMembers++
			}

			// 10. ClassSessions
			var sessions []*domain.ClassSession
			for range 3 {
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

			// 11. Quiz — first quiz per org demos quiz-wide negative marking.
			quiz := factory.NewQuiz(org.ID, teacher.ID, class.ID)
			if c == 0 {
				quiz.NegativeMarkMode = domain.NegativeMarkAccumulative
				quiz.NegativeValue = 0
				quiz.WrongsPerPoint = 3
			}
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
				// Demo per-question override (Layer 2a) on the first question.
				if len(questionIDs) > 0 {
					r.NegativeOverrides = []domain.QuizQuestionNegativeOverride{
						{QuestionID: questionIDs[0], Mode: domain.NegativeMarkPerWrong, NegativeValue: 0.5},
					}
				}
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
			for s := range subCount {
				student := ou.students[s]
				var answers []domain.SubmissionAnswer
				var questionSet []domain.SubmissionQuestion
				totalScore := 0.0
				for _, q := range bankQuestions {
					earned := 0.0
					var selectedIDs []string
					var value string
					optionOrder := make([]string, len(q.Options))
					for i, opt := range q.Options {
						optionOrder[i] = opt.ID
					}
					questionSet = append(questionSet, domain.SubmissionQuestion{
						QuestionID:    q.ID,
						OptionIDOrder: optionOrder,
					})
					if q.Type == domain.QuestionTypeChoice && len(q.Options) > 0 {
						selectedIDs = []string{q.Options[0].ID}
						earned = q.Options[0].Score
					} else if q.Type == domain.QuestionTypeShortAnswer {
						value = factory.T("Sample answer", "پاسخ نمونه")
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
				// Frozen per-student question set mirroring the answers, as
				// StartSubmission would have built it.
				questionSet := make([]domain.SubmissionQuestion, 0, len(bankQuestions))
				for _, q := range bankQuestions {
					var order []string
					if q.Type == domain.QuestionTypeChoice {
						order = make([]string, len(q.Options))
						for i, o := range q.Options {
							order[i] = o.ID
						}
					}
					questionSet = append(questionSet, domain.SubmissionQuestion{QuestionID: q.ID, OptionIDOrder: order})
				}
				sub := factory.NewQuizSubmission(quiz.ID, student.ID, func(qs *domain.QuizSubmission) {
					qs.Answers = answers
					qs.QuestionSet = questionSet
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
				p.Name = factory.T("Class feedback", "بازخورد کلاس")
				p.AllowedAnswersCount = 1
				p.Options = []domain.PollOption{
					{Label: factory.T("Great", "عالی"), Value: "great"},
					{Label: factory.T("Okay", "متوسط"), Value: "okay"},
					{Label: factory.T("Bad", "ضعیف"), Value: "bad"},
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
				c.Name = class.Name + factory.T(" Chat", " گفتگو")
			})
			if err := db.WithContext(ctx).Create(chat).Error; err != nil {
				return nil, fmt.Errorf("creating chat: %w", err)
			}
			counts.Chats++
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
			tid := teacher.ID
			welcome := factory.NewMessage(chat.ID, &tid, func(m *domain.Message) {
				m.Content = factory.T("Welcome to the class!", "به کلاس خوش آمدید!")
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
			for _, student := range ou.students {
				r := factory.NewMessageReaction(welcome.ID, student.ID, "👍")
				if err := db.WithContext(ctx).Create(r).Error; err != nil {
					return nil, fmt.Errorf("creating reaction: %w", err)
				}
				counts.MessageReactions++
			}

			// 17. LiveRoom on liveSession + participants + recording
			liveRoom := factory.NewLiveRoom(liveSession.ID, func(lr *domain.LiveRoom) {
				lr.Name = factory.T("Past session", "جلسه گذشته")
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
			// Mirror CreateRoom: every live room owns a live_session chat so the
			// in-room chat panel has a backing chat to post to.
			liveRoomChat := factory.NewChat("live_session", liveRoom.ID, func(c *domain.Chat) {
				c.Name = factory.T("Chat – Past session", "گفتگو – جلسه گذشته")
			})
			if err := db.WithContext(ctx).Create(liveRoomChat).Error; err != nil {
				return nil, fmt.Errorf("creating live room chat: %w", err)
			}
			counts.Chats++
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
				lr.Name = factory.T("Scheduled session", "جلسه زمان‌بندی‌شده")
				lr.Status = domain.LiveRoomStatusCreated
				at := time.Now().Add(24 * time.Hour)
				lr.ScheduledStartTime = &at
			})
			if err := db.WithContext(ctx).Create(scheduledRoom).Error; err != nil {
				return nil, fmt.Errorf("creating scheduled live room: %w", err)
			}
			counts.LiveRooms++
			scheduledRoomChat := factory.NewChat("live_session", scheduledRoom.ID, func(c *domain.Chat) {
				c.Name = factory.T("Chat – Scheduled session", "گفتگو – جلسه زمان‌بندی‌شده")
			})
			if err := db.WithContext(ctx).Create(scheduledRoomChat).Error; err != nil {
				return nil, fmt.Errorf("creating scheduled live room chat: %w", err)
			}
			counts.Chats++

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
				{factory.T("Attendance", "حضور و غیاب"), domain.GradebookColumnAutoAttendance},
				{factory.T("Quiz Score", "نمره آزمون"), domain.GradebookColumnAutoQuiz},
				{factory.T("Midterm", "میان‌ترم"), domain.GradebookColumnManualGrade},
				{factory.T("Notes", "یادداشت‌ها"), domain.GradebookColumnManualText},
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
						value = factory.T("OK", "خوب")
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
	fmt.Println("  manager2 / password (Manager preset in Zoora Demo org)")
	fmt.Println("  user1 / password    (Student in Zoora Demo org)")
	fmt.Println("  user2 / password    (Student in Zoora Demo org)")
}
