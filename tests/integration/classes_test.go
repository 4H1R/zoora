//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/classes"
	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/organizations"
	"github.com/4H1R/zoora/internal/users"
	"github.com/4H1R/zoora/tests/testutil"
)

type classRepos struct {
	classes  domain.ClassRepository
	sessions domain.ClassSessionRepository
	members  domain.ClassMemberRepository
	users    domain.UserRepository
	orgs     domain.OrganizationRepository
}

func setupClassesDB(t *testing.T) classRepos {
	t.Helper()
	db := testutil.SetupPostgres(t)
	require.NoError(t, db.AutoMigrate(
		&domain.Organization{},
		&domain.User{},
		&domain.Class{},
		&domain.ClassSession{},
		&domain.ClassMember{},
	))
	return classRepos{
		classes:  classes.NewRepository(db),
		sessions: classes.NewSessionRepository(db),
		members:  classes.NewMemberRepository(db),
		users:    users.NewRepository(db),
		orgs:     organizations.NewRepository(db),
	}
}

func seedOrg(t *testing.T, repo domain.OrganizationRepository, name string) *domain.Organization {
	t.Helper()
	org := &domain.Organization{Name: name}
	require.NoError(t, repo.Create(context.Background(), org))
	return org
}

func seedTeacher(t *testing.T, repo domain.UserRepository, orgID uuid.UUID, username string) *domain.User {
	t.Helper()
	u := &domain.User{
		OrganizationID: &orgID,
		Username:       username,
		Name:           username,
		Password:       "x",
	}
	require.NoError(t, repo.Create(context.Background(), u))
	return u
}

func seedClass(t *testing.T, r classRepos, orgID, teacherID uuid.UUID, name string, capacity int) *domain.Class {
	t.Helper()
	c := &domain.Class{
		OrganizationID: orgID,
		UserID:         teacherID,
		Name:           name,
		TotalUsers:     capacity,
	}
	require.NoError(t, r.classes.Create(context.Background(), c))
	return c
}

func TestIntegration_ClassRepo_CRUD(t *testing.T) {
	r := setupClassesDB(t)
	ctx := context.Background()
	org := seedOrg(t, r.orgs, "Acme")
	teacher := seedTeacher(t, r.users, org.ID, "t1")

	c := seedClass(t, r, org.ID, teacher.ID, "Algebra", 30)

	got, err := r.classes.FindByID(ctx, c.ID)
	require.NoError(t, err)
	assert.Equal(t, "Algebra", got.Name)
	assert.Equal(t, 30, got.TotalUsers)

	got.Name = "Algebra II"
	require.NoError(t, r.classes.Update(ctx, got))

	require.NoError(t, r.classes.Delete(ctx, c.ID))
	_, err = r.classes.FindByID(ctx, c.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)

	// Soft-deleted row still visible via Unscoped find.
	soft, err := r.classes.FindByIDIncludingDeleted(ctx, c.ID)
	require.NoError(t, err)
	assert.True(t, soft.DeletedAt.Valid)

	require.NoError(t, r.classes.HardDelete(ctx, c.ID))
	_, err = r.classes.FindByIDIncludingDeleted(ctx, c.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestIntegration_ClassRepo_List_ScopeTeacher(t *testing.T) {
	r := setupClassesDB(t)
	ctx := context.Background()
	org := seedOrg(t, r.orgs, "Acme")
	alice := seedTeacher(t, r.users, org.ID, "alice")
	bob := seedTeacher(t, r.users, org.ID, "bob")

	seedClass(t, r, org.ID, alice.ID, "AliceA", 0)
	seedClass(t, r, org.ID, alice.ID, "AliceB", 0)
	seedClass(t, r, org.ID, bob.ID, "BobA", 0)

	bigPage := domain.ListParams{Page: 1, PageSize: 50}
	scope := domain.ClassListScope{TeacherID: &alice.ID}
	classes, total, err := r.classes.List(ctx, scope, bigPage)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, classes, 2)
}

func TestIntegration_ClassRepo_List_ScopeMember(t *testing.T) {
	r := setupClassesDB(t)
	ctx := context.Background()
	org := seedOrg(t, r.orgs, "Acme")
	teacher := seedTeacher(t, r.users, org.ID, "teach")
	student := seedTeacher(t, r.users, org.ID, "stud")

	enrolled := seedClass(t, r, org.ID, teacher.ID, "WithStudent", 0)
	_ = seedClass(t, r, org.ID, teacher.ID, "NoStudent", 0)

	require.NoError(t, r.members.Create(ctx, &domain.ClassMember{ClassID: enrolled.ID, UserID: student.ID}))

	bigPage := domain.ListParams{Page: 1, PageSize: 50}
	scope := domain.ClassListScope{MemberUserID: &student.ID}
	classes, total, err := r.classes.List(ctx, scope, bigPage)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, enrolled.ID, classes[0].ID)
}

func TestIntegration_ClassRepo_List_ScopeAllOrgs(t *testing.T) {
	r := setupClassesDB(t)
	ctx := context.Background()
	orgA := seedOrg(t, r.orgs, "A")
	orgB := seedOrg(t, r.orgs, "B")
	tA := seedTeacher(t, r.users, orgA.ID, "ta")
	tB := seedTeacher(t, r.users, orgB.ID, "tb")
	seedClass(t, r, orgA.ID, tA.ID, "ClassA", 0)
	seedClass(t, r, orgB.ID, tB.ID, "ClassB", 0)

	bigPage := domain.ListParams{Page: 1, PageSize: 50}
	_, total, err := r.classes.List(ctx, domain.ClassListScope{All: true}, bigPage)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
}

func TestIntegration_ClassRepo_AdminList_SearchAndFilter(t *testing.T) {
	r := setupClassesDB(t)
	ctx := context.Background()
	org := seedOrg(t, r.orgs, "Acme")
	teacher := seedTeacher(t, r.users, org.ID, "teach")

	seedClass(t, r, org.ID, teacher.ID, "Algebra", 0)
	seedClass(t, r, org.ID, teacher.ID, "Algorithms", 0)
	seedClass(t, r, org.ID, teacher.ID, "Biology", 0)

	bigPage := domain.ListParams{
		Page:         1,
		PageSize:     50,
		Search:       "algo",
		SearchFields: []string{"name", "description"},
	}
	classes, total, err := r.classes.AdminList(ctx, domain.AdminListClassesQuery{ListParams: bigPage})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "Algorithms", classes[0].Name)

	// Filter by user_id (teacher).
	q := domain.AdminListClassesQuery{
		UserID:     &teacher.ID,
		ListParams: domain.ListParams{Page: 1, PageSize: 50},
	}
	_, total, err = r.classes.AdminList(ctx, q)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
}

func TestIntegration_SessionRepo_CRUD_FilterType(t *testing.T) {
	r := setupClassesDB(t)
	ctx := context.Background()
	org := seedOrg(t, r.orgs, "Acme")
	teacher := seedTeacher(t, r.users, org.ID, "t")
	c := seedClass(t, r, org.ID, teacher.ID, "Math", 0)

	start := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	s1 := &domain.ClassSession{ClassID: c.ID, Name: "Live1", StartTime: start, Type: domain.ClassSessionTypeLive}
	s2 := &domain.ClassSession{ClassID: c.ID, Name: "Quiz1", StartTime: start.Add(time.Hour), Type: domain.ClassSessionTypeQuiz}
	require.NoError(t, r.sessions.Create(ctx, s1))
	require.NoError(t, r.sessions.Create(ctx, s2))

	bigPage := domain.ListParams{Page: 1, PageSize: 50}

	// Unfiltered list.
	_, total, err := r.sessions.ListByClass(ctx, c.ID, domain.ListClassSessionsQuery{ListParams: bigPage})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)

	// Filter by type=live.
	live := domain.ClassSessionTypeLive
	got, total, err := r.sessions.ListByClass(ctx, c.ID, domain.ListClassSessionsQuery{Type: &live, ListParams: bigPage})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "Live1", got[0].Name)
}

func TestIntegration_MemberRepo_UniqueAndCapacityCount(t *testing.T) {
	r := setupClassesDB(t)
	ctx := context.Background()
	org := seedOrg(t, r.orgs, "Acme")
	teacher := seedTeacher(t, r.users, org.ID, "t")
	studentA := seedTeacher(t, r.users, org.ID, "sa")
	studentB := seedTeacher(t, r.users, org.ID, "sb")
	c := seedClass(t, r, org.ID, teacher.ID, "Cap", 0)

	require.NoError(t, r.members.Create(ctx, &domain.ClassMember{ClassID: c.ID, UserID: studentA.ID}))
	require.NoError(t, r.members.Create(ctx, &domain.ClassMember{ClassID: c.ID, UserID: studentB.ID}))

	// Duplicate (class_id, user_id) must surface as ErrConflict via unique violation.
	err := r.members.Create(ctx, &domain.ClassMember{ClassID: c.ID, UserID: studentA.ID})
	assert.ErrorIs(t, err, domain.ErrConflict)

	count, err := r.members.CountByClass(ctx, c.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	exists, err := r.members.Exists(ctx, c.ID, studentA.ID)
	require.NoError(t, err)
	assert.True(t, exists)

	require.NoError(t, r.members.Delete(ctx, c.ID, studentA.ID))
	exists, err = r.members.Exists(ctx, c.ID, studentA.ID)
	require.NoError(t, err)
	assert.False(t, exists)

	// Second delete → ErrNotFound.
	err = r.members.Delete(ctx, c.ID, studentA.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}
