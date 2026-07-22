//go:build integration

package integration

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/attendance"
	"github.com/4H1R/zoora/internal/classes"
	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/gradebook"
	"github.com/4H1R/zoora/internal/organizations"
	"github.com/4H1R/zoora/internal/platform/authz"
	"github.com/4H1R/zoora/internal/quizzes"
	"github.com/4H1R/zoora/internal/users"
	"github.com/4H1R/zoora/tests/testutil"
)

func TestIntegration_QuizRepo_ListByMemberWithRooms(t *testing.T) {
	r := setupQuizzesDB(t)
	ctx := context.Background()
	f := seedQuizFixture(t, r)

	// A second class the student is NOT enrolled in, with its own quiz.
	otherTeacher := seedTeacher(t, r.users, f.org.ID, "teacher2")
	otherClass := &domain.Class{OrganizationID: f.org.ID, UserID: otherTeacher.ID, Name: "Science", TotalUsers: 10}
	require.NoError(t, r.classes.Create(ctx, otherClass))
	otherQuiz := &domain.Quiz{OrganizationID: f.org.ID, UserID: otherTeacher.ID, ClassID: otherClass.ID, Title: "Other", DurationMinutes: 30}
	require.NoError(t, r.quizzes.Create(ctx, otherQuiz))

	got, err := r.quizzes.ListByMemberWithRooms(ctx, f.student.ID, nil, domain.ListParams{Page: 1, PageSize: 50})
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, f.quiz.ID, got[0].ID)
	require.NotNil(t, got[0].Class)
	assert.Equal(t, "Math", got[0].Class.Name)
}

func TestIntegration_QuizService_ListMine_States(t *testing.T) {
	r := setupQuizzesDB(t)
	ctx := context.Background()
	f := seedQuizFixture(t, r)

	// Second quiz in the same (enrolled) class, which we'll grade.
	gradedQuiz := &domain.Quiz{OrganizationID: f.org.ID, UserID: f.teacher.ID, ClassID: f.class.ID, Title: "Graded", DurationMinutes: 30, TotalScore: 20}
	require.NoError(t, r.quizzes.Create(ctx, gradedQuiz))

	// Open room for the first quiz (started in the past, ends in the future).
	start := time.Now().Add(-time.Minute).UTC()
	end := time.Now().Add(time.Hour).UTC()
	require.NoError(t, r.rooms.Create(ctx, &domain.QuizRoom{QuizID: f.quiz.ID, ClassSessionID: f.session.ID, StartedAt: &start, EndedAt: &end}))

	// Graded submission for the second quiz by the student.
	submittedAt := time.Now().UTC()
	require.NoError(t, r.submissions.Create(ctx, &domain.QuizSubmission{
		QuizID:      gradedQuiz.ID,
		UserID:      f.student.ID,
		Status:      domain.SubmissionStatusGraded,
		TotalScore:  18,
		StartedAt:   submittedAt,
		SubmittedAt: &submittedAt,
	}))

	svc := quizzes.NewService(r.quizzes, r.rules, r.rooms, r.submissions, r.questions, r.classes, r.members, nil, nil, nil, slog.Default())

	callerCtx := domain.WithCaller(ctx, domain.Caller{
		UserID:      f.student.ID,
		Permissions: []string{string(domain.PermQuizzesTake)},
	})

	exams, total, err := svc.ListMine(callerCtx, domain.ListMyExamsQuery{ListParams: domain.ListParams{Page: 1, PageSize: 50}})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, exams, 2)

	byID := map[string]domain.MyExam{}
	for _, e := range exams {
		byID[e.QuizID.String()] = e
	}

	open := byID[f.quiz.ID.String()]
	assert.Equal(t, domain.MyExamStateOpen, open.State)
	require.NotNil(t, open.Room)
	assert.True(t, open.Room.IsOpen)
	assert.Equal(t, f.session.ID, open.Room.ClassSessionID)

	graded := byID[gradedQuiz.ID.String()]
	assert.Equal(t, domain.MyExamStateGraded, graded.State)
	require.NotNil(t, graded.Score)
	assert.Equal(t, 18.0, *graded.Score)
}

func TestIntegration_GradebookService_GetMine(t *testing.T) {
	db := testutil.SetupPostgres(t)
	require.NoError(t, db.AutoMigrate(
		&domain.Organization{},
		&domain.User{},
		&domain.Class{},
		&domain.ClassSession{},
		&domain.ClassMember{},
		&domain.GradebookColumn{},
		&domain.GradebookCell{},
	))
	// The (column_id, student_id) uniqueness lives in migration SQL, not the
	// model tags, so AutoMigrate doesn't create it — add it for the Upsert path.
	require.NoError(t, db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS uq_gradebook_cells_column_student ON gradebook_cells(column_id, student_id)").Error)
	ctx := context.Background()

	orgRepo := organizations.NewRepository(db)
	userRepo := users.NewRepository(db)
	classRepo := classes.NewRepository(db)
	memberRepo := classes.NewMemberRepository(db)
	columnRepo := gradebook.NewColumnRepository(db)
	cellRepo := gradebook.NewCellRepository(db)

	org := seedOrg(t, orgRepo, "Acme")
	teacher := seedTeacher(t, userRepo, org.ID, "teacher")
	student := seedTeacher(t, userRepo, org.ID, "student")
	other := seedTeacher(t, userRepo, org.ID, "other")

	cls := &domain.Class{OrganizationID: org.ID, UserID: teacher.ID, Name: "Math", TotalUsers: 30}
	require.NoError(t, classRepo.Create(ctx, cls))
	require.NoError(t, memberRepo.Create(ctx, &domain.ClassMember{ClassID: cls.ID, UserID: student.ID}))
	require.NoError(t, memberRepo.Create(ctx, &domain.ClassMember{ClassID: cls.ID, UserID: other.ID}))

	col := &domain.GradebookColumn{ClassID: cls.ID, Title: "Midterm", Type: domain.GradebookColumnManualGrade}
	require.NoError(t, columnRepo.Create(ctx, col))
	require.NoError(t, cellRepo.Upsert(ctx, &domain.GradebookCell{ColumnID: col.ID, StudentID: student.ID, Value: "18"}))
	require.NoError(t, cellRepo.Upsert(ctx, &domain.GradebookCell{ColumnID: col.ID, StudentID: other.ID, Value: "11"}))

	svc := gradebook.NewService(columnRepo, cellRepo, classRepo, memberRepo, nil, nil, nil, nil, nil, nil, authz.NewResolver(memberRepo), slog.Default())

	callerCtx := domain.WithCaller(ctx, domain.Caller{
		UserID:      student.ID,
		Permissions: []string{string(domain.PermGradebookView)},
	})

	rc, err := svc.GetMine(callerCtx)
	require.NoError(t, err)
	require.Len(t, rc.Classes, 1)
	assert.Equal(t, "Math", rc.Classes[0].ClassName)
	require.Len(t, rc.Classes[0].Columns, 1)
	assert.Equal(t, "18", rc.Classes[0].Cells[col.ID.String()])
	// The other student's value must not leak in.
	assert.NotContains(t, []string{rc.Classes[0].Cells[col.ID.String()]}, "11")
}

func TestIntegration_Attendance_ListMine(t *testing.T) {
	db := testutil.SetupPostgres(t)
	require.NoError(t, db.AutoMigrate(
		&domain.Organization{},
		&domain.User{},
		&domain.Class{},
		&domain.ClassSession{},
		&domain.ClassMember{},
		&domain.Attendance{},
	))
	ctx := context.Background()

	orgRepo := organizations.NewRepository(db)
	userRepo := users.NewRepository(db)
	classRepo := classes.NewRepository(db)
	sessRepo := classes.NewSessionRepository(db)
	attRepo := attendance.NewRepository(db)

	org := seedOrg(t, orgRepo, "Acme")
	teacher := seedTeacher(t, userRepo, org.ID, "teacher")
	student := seedTeacher(t, userRepo, org.ID, "student")

	cls := &domain.Class{OrganizationID: org.ID, UserID: teacher.ID, Name: "Math", TotalUsers: 30}
	require.NoError(t, classRepo.Create(ctx, cls))
	sess := &domain.ClassSession{ClassID: cls.ID, Name: "S1", StartTime: time.Now().UTC()}
	require.NoError(t, sessRepo.Create(ctx, sess))

	statuses := []domain.AttendanceStatus{
		domain.AttendanceStatusPresent,
		domain.AttendanceStatusPresent,
		domain.AttendanceStatusAbsent,
		domain.AttendanceStatusLate,
	}
	for _, st := range statuses {
		require.NoError(t, attRepo.Create(ctx, &domain.Attendance{
			OrganizationID: org.ID,
			ClassID:        cls.ID,
			ClassSessionID: sess.ID,
			UserID:         student.ID,
			Status:         st,
		}))
	}
	// A record for a different user must not be counted.
	other := seedTeacher(t, userRepo, org.ID, "other")
	require.NoError(t, attRepo.Create(ctx, &domain.Attendance{
		OrganizationID: org.ID, ClassID: cls.ID, ClassSessionID: sess.ID, UserID: other.ID, Status: domain.AttendanceStatusPresent,
	}))

	// Repo-level scoping.
	rows, total, err := attRepo.ListByUser(ctx, student.ID, domain.ListMyAttendanceQuery{
		ListParams: domain.ListParams{Page: 1, PageSize: 50},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(4), total)
	assert.Len(t, rows, 4)

	// Status filter narrows the set.
	present := domain.AttendanceStatusPresent
	rows, total, err = attRepo.ListByUser(ctx, student.ID, domain.ListMyAttendanceQuery{
		Status:     &present,
		ListParams: domain.ListParams{Page: 1, PageSize: 50},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, rows, 2)

	// Service-level summary.
	svc := attendance.NewService(attRepo, classRepo, sessRepo, nil, nil, nil, nil, nil, nil, authz.NewResolver(nil), slog.Default())
	callerCtx := domain.WithCaller(ctx, domain.Caller{
		UserID:      student.ID,
		Permissions: []string{string(domain.PermAttendanceView)},
	})
	res, err := svc.ListMine(callerCtx, domain.ListMyAttendanceQuery{
		ListParams: domain.ListParams{Page: 1, PageSize: 50},
	})
	require.NoError(t, err)
	assert.Equal(t, 2, res.Summary.Present)
	assert.Equal(t, 1, res.Summary.Absent)
	assert.Equal(t, 1, res.Summary.Late)
	assert.Equal(t, 0, res.Summary.Excused)
	assert.Len(t, res.Items, 4)
	assert.Equal(t, int64(4), res.Total)

	// Summary must aggregate the FULL filtered set even when the page is
	// smaller than the result set.
	res, err = svc.ListMine(callerCtx, domain.ListMyAttendanceQuery{
		ListParams: domain.ListParams{Page: 1, PageSize: 2},
	})
	require.NoError(t, err)
	assert.Len(t, res.Items, 2)
	assert.Equal(t, int64(4), res.Total)
	assert.Equal(t, 2, res.Summary.Present)
	assert.Equal(t, 1, res.Summary.Absent)
	assert.Equal(t, 1, res.Summary.Late)

	// Status filter narrows items/total, but the summary stays a full
	// breakdown by status over the class/session scope.
	res, err = svc.ListMine(callerCtx, domain.ListMyAttendanceQuery{
		Status:     &present,
		ListParams: domain.ListParams{Page: 1, PageSize: 50},
	})
	require.NoError(t, err)
	assert.Len(t, res.Items, 2)
	assert.Equal(t, int64(2), res.Total)
	assert.Equal(t, 2, res.Summary.Present)
	assert.Equal(t, 1, res.Summary.Absent)
}
