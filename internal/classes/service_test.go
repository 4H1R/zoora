package classes_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/4H1R/zoora/internal/classes"
	"github.com/4H1R/zoora/internal/domain"
)

type classRepoSvcMock struct{ mock.Mock }

func (m *classRepoSvcMock) Create(ctx context.Context, class *domain.Class) error {
	return m.Called(ctx, class).Error(0)
}

func (m *classRepoSvcMock) FindByID(ctx context.Context, id uuid.UUID) (*domain.Class, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Class), args.Error(1)
}

func (m *classRepoSvcMock) Update(ctx context.Context, class *domain.Class) error {
	return m.Called(ctx, class).Error(0)
}

func (m *classRepoSvcMock) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *classRepoSvcMock) List(ctx context.Context, scope domain.ClassListScope, p domain.ListParams) ([]domain.Class, int64, error) {
	args := m.Called(ctx, scope, p)
	items, _ := args.Get(0).([]domain.Class)
	return items, args.Get(1).(int64), args.Error(2)
}

func (m *classRepoSvcMock) ListByNames(ctx context.Context, orgID uuid.UUID, names []string) ([]domain.Class, error) {
	args := m.Called(ctx, orgID, names)
	items, _ := args.Get(0).([]domain.Class)
	return items, args.Error(1)
}

func (m *classRepoSvcMock) HardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *classRepoSvcMock) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.Class, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Class), args.Error(1)
}

func (m *classRepoSvcMock) AdminList(ctx context.Context, q domain.AdminListClassesQuery) ([]domain.Class, int64, error) {
	args := m.Called(ctx, q)
	items, _ := args.Get(0).([]domain.Class)
	return items, args.Get(1).(int64), args.Error(2)
}

type classSessionRepoSvcMock struct{ mock.Mock }

func (m *classSessionRepoSvcMock) Create(ctx context.Context, session *domain.ClassSession) error {
	return m.Called(ctx, session).Error(0)
}

func (m *classSessionRepoSvcMock) FindByID(ctx context.Context, id uuid.UUID) (*domain.ClassSession, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ClassSession), args.Error(1)
}

func (m *classSessionRepoSvcMock) Update(ctx context.Context, session *domain.ClassSession) error {
	return m.Called(ctx, session).Error(0)
}

func (m *classSessionRepoSvcMock) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *classSessionRepoSvcMock) ListByClass(ctx context.Context, classID uuid.UUID, q domain.ListClassSessionsQuery) ([]domain.ClassSession, int64, error) {
	args := m.Called(ctx, classID, q)
	items, _ := args.Get(0).([]domain.ClassSession)
	return items, args.Get(1).(int64), args.Error(2)
}

func (m *classSessionRepoSvcMock) HardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *classSessionRepoSvcMock) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.ClassSession, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ClassSession), args.Error(1)
}

func (m *classSessionRepoSvcMock) AdminList(ctx context.Context, q domain.AdminListClassSessionsQuery) ([]domain.ClassSession, int64, error) {
	args := m.Called(ctx, q)
	items, _ := args.Get(0).([]domain.ClassSession)
	return items, args.Get(1).(int64), args.Error(2)
}

type classMemberRepoSvcMock struct{ mock.Mock }

func (m *classMemberRepoSvcMock) Create(ctx context.Context, member *domain.ClassMember) error {
	return m.Called(ctx, member).Error(0)
}

func (m *classMemberRepoSvcMock) Delete(ctx context.Context, classID, userID uuid.UUID) error {
	return m.Called(ctx, classID, userID).Error(0)
}

func (m *classMemberRepoSvcMock) Exists(ctx context.Context, classID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, classID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *classMemberRepoSvcMock) CountByClass(ctx context.Context, classID uuid.UUID) (int64, error) {
	args := m.Called(ctx, classID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *classMemberRepoSvcMock) ListByClass(ctx context.Context, classID uuid.UUID, p domain.ListParams) ([]domain.ClassMember, int64, error) {
	args := m.Called(ctx, classID, p)
	items, _ := args.Get(0).([]domain.ClassMember)
	return items, args.Get(1).(int64), args.Error(2)
}

func (m *classMemberRepoSvcMock) ListAllByClass(ctx context.Context, classID uuid.UUID) ([]domain.ClassMember, error) {
	args := m.Called(ctx, classID)
	items, _ := args.Get(0).([]domain.ClassMember)
	return items, args.Error(1)
}

func newClassService(repo *classRepoSvcMock, sessions *classSessionRepoSvcMock, members *classMemberRepoSvcMock, chat ...domain.ClassChatProvisioner) domain.ClassService {
	if repo == nil {
		repo = &classRepoSvcMock{}
	}
	if sessions == nil {
		sessions = &classSessionRepoSvcMock{}
	}
	if members == nil {
		members = &classMemberRepoSvcMock{}
	}
	var provisioner domain.ClassChatProvisioner
	if len(chat) > 0 {
		provisioner = chat[0]
	}
	return classes.NewService(repo, sessions, members, provisioner, slog.Default())
}

// chatProvisionerMock is a testify mock of domain.ClassChatProvisioner.
type chatProvisionerMock struct{ mock.Mock }

func (m *chatProvisionerMock) CreateForClass(ctx context.Context, in domain.ProvisionClassChatDTO) (*domain.Conversation, error) {
	args := m.Called(ctx, in)
	conv, _ := args.Get(0).(*domain.Conversation)
	return conv, args.Error(1)
}

func (m *chatProvisionerMock) SyncClassMembers(ctx context.Context, convID uuid.UUID, memberIDs []uuid.UUID) (*domain.Conversation, error) {
	args := m.Called(ctx, convID, memberIDs)
	conv, _ := args.Get(0).(*domain.Conversation)
	return conv, args.Error(1)
}

func classCtx(userID uuid.UUID, orgID *uuid.UUID, isAdmin bool, perms ...domain.PermissionName) context.Context {
	p := make([]string, 0, len(perms))
	for _, perm := range perms {
		p = append(p, string(perm))
	}
	return domain.WithCaller(context.Background(), domain.Caller{UserID: userID, OrgID: orgID, IsAdmin: isAdmin, Permissions: p})
}

func TestClassCreateRequiresCallerOrgAndCreatePermission(t *testing.T) {
	repo := &classRepoSvcMock{}
	svc := newClassService(repo, nil, nil)
	teacherID := uuid.New()
	orgID := uuid.New()

	_, err := svc.Create(context.Background(), domain.CreateClassDTO{Name: "Math"})
	assert.ErrorIs(t, err, domain.ErrForbidden)

	_, err = svc.Create(classCtx(teacherID, nil, false, domain.PermClassesCreate), domain.CreateClassDTO{Name: "Math"})
	assert.ErrorIs(t, err, domain.ErrForbidden)

	_, err = svc.Create(classCtx(teacherID, &orgID, false), domain.CreateClassDTO{Name: "Math"})
	assert.ErrorIs(t, err, domain.ErrForbidden)

	ctx := classCtx(teacherID, &orgID, false, domain.PermClassesCreate)
	repo.On("Create", ctx, mock.MatchedBy(func(class *domain.Class) bool {
		return class.OrganizationID == orgID &&
			class.UserID == teacherID &&
			class.Name == "Math" &&
			class.TotalUsers == 30
	})).Return(nil)

	created, err := svc.Create(ctx, domain.CreateClassDTO{Name: "Math", TotalUsers: 30})
	assert.NoError(t, err)
	assert.Equal(t, teacherID, created.UserID)
}

func TestClassCreateAnyCanAssignOwner(t *testing.T) {
	repo := &classRepoSvcMock{}
	svc := newClassService(repo, nil, nil)
	adminID := uuid.New()
	teacherID := uuid.New()
	orgID := uuid.New()
	ctx := classCtx(adminID, &orgID, false, domain.PermClassesCreateAny)

	repo.On("Create", ctx, mock.MatchedBy(func(class *domain.Class) bool {
		return class.UserID == teacherID && class.OrganizationID == orgID
	})).Return(nil)

	created, err := svc.Create(ctx, domain.CreateClassDTO{Name: "Science", UserID: &teacherID})
	assert.NoError(t, err)
	assert.Equal(t, teacherID, created.UserID)
}

func TestClassListScopesByRole(t *testing.T) {
	params := domain.ListParams{Page: 2, PageSize: 10}
	userID := uuid.New()
	orgID := uuid.New()

	tests := []struct {
		name      string
		ctx       context.Context
		wantScope func(domain.ClassListScope) bool
	}{
		{
			name: "admin sees all",
			ctx:  classCtx(userID, nil, true),
			wantScope: func(scope domain.ClassListScope) bool {
				return scope.All && scope.OrganizationID == nil && scope.TeacherID == nil && scope.MemberUserID == nil
			},
		},
		{
			name: "org staff sees own org",
			ctx:  classCtx(userID, &orgID, false, domain.PermClassesViewAny),
			wantScope: func(scope domain.ClassListScope) bool {
				return scope.All && scope.OrganizationID != nil && *scope.OrganizationID == orgID
			},
		},
		{
			name: "regular user sees own teaching or enrolled classes",
			ctx:  classCtx(userID, &orgID, false),
			wantScope: func(scope domain.ClassListScope) bool {
				return !scope.All &&
					scope.TeacherID != nil && *scope.TeacherID == userID &&
					scope.MemberUserID != nil && *scope.MemberUserID == userID
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &classRepoSvcMock{}
			svc := newClassService(repo, nil, nil)
			repo.On("List", tt.ctx, mock.MatchedBy(tt.wantScope), params).
				Return([]domain.Class{{ID: uuid.New(), Name: "Class"}}, int64(1), nil)

			items, total, err := svc.List(tt.ctx, params)
			assert.NoError(t, err)
			assert.Len(t, items, 1)
			assert.Equal(t, int64(1), total)
		})
	}
}

func TestClassEnrollmentRulesAndCapacity(t *testing.T) {
	classID := uuid.New()
	teacherID := uuid.New()
	studentID := uuid.New()
	otherStudentID := uuid.New()

	tests := []struct {
		name       string
		callerID   uuid.UUID
		targetID   uuid.UUID
		class      *domain.Class
		count      int64
		wantErr    error
		wantCreate bool
	}{
		{
			name:       "student cannot enroll another user",
			callerID:   studentID,
			targetID:   otherStudentID,
			class:      &domain.Class{ID: classID, UserID: teacherID, TotalUsers: 0},
			wantErr:    domain.ErrForbidden,
			wantCreate: false,
		},
		{
			name:       "capacity full conflicts",
			callerID:   studentID,
			targetID:   studentID,
			class:      &domain.Class{ID: classID, UserID: teacherID, TotalUsers: 1},
			count:      1,
			wantErr:    domain.ErrConflict,
			wantCreate: false,
		},
		{
			name:       "student can self enroll when capacity available",
			callerID:   studentID,
			targetID:   studentID,
			class:      &domain.Class{ID: classID, UserID: teacherID, TotalUsers: 2},
			count:      1,
			wantCreate: true,
		},
		{
			name:       "teacher can enroll another user",
			callerID:   teacherID,
			targetID:   otherStudentID,
			class:      &domain.Class{ID: classID, UserID: teacherID, TotalUsers: 0},
			wantCreate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &classRepoSvcMock{}
			members := &classMemberRepoSvcMock{}
			svc := newClassService(repo, nil, members)
			ctx := classCtx(tt.callerID, nil, false)

			repo.On("FindByID", ctx, classID).Return(tt.class, nil)
			if tt.class.TotalUsers > 0 && !assert.ObjectsAreEqual(tt.wantErr, domain.ErrForbidden) {
				members.On("CountByClass", ctx, classID).Return(tt.count, nil)
			}
			if tt.wantCreate {
				members.On("Create", ctx, mock.MatchedBy(func(member *domain.ClassMember) bool {
					return member.ClassID == classID && member.UserID == tt.targetID
				})).Return(nil)
			}

			member, err := svc.Enroll(ctx, classID, domain.EnrollClassMemberDTO{UserID: tt.targetID})
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, member)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, member)
			}
			if !tt.wantCreate {
				members.AssertNotCalled(t, "Create")
			}
		})
	}
}

func TestClassListSessionsStripsIncludeDeletedForStudents(t *testing.T) {
	classID := uuid.New()
	userID := uuid.New()
	teacherID := uuid.New()
	ctx := classCtx(userID, nil, false)
	repo := &classRepoSvcMock{}
	members := &classMemberRepoSvcMock{}
	sessions := &classSessionRepoSvcMock{}
	svc := newClassService(repo, sessions, members)

	repo.On("FindByID", ctx, classID).Return(&domain.Class{ID: classID, UserID: teacherID}, nil)
	members.On("Exists", ctx, classID, userID).Return(true, nil)
	sessions.On("ListByClass", ctx, classID, mock.MatchedBy(func(q domain.ListClassSessionsQuery) bool {
		return !q.IncludeDeleted && q.ListParams.Page == 1
	})).Return([]domain.ClassSession{{ID: uuid.New(), ClassID: classID}}, int64(1), nil)

	items, total, err := svc.ListSessions(ctx, classID, domain.ListClassSessionsQuery{
		IncludeDeleted: true,
		ListParams:     domain.ListParams{Page: 1, PageSize: 10},
	})

	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, int64(1), total)
}

func TestClassSessionUpdateRequiresManagerAndAppliesPartialFields(t *testing.T) {
	classID := uuid.New()
	sessionID := uuid.New()
	teacherID := uuid.New()
	studentID := uuid.New()
	start := time.Now().Add(2 * time.Hour)
	newName := "Updated"
	ctx := classCtx(studentID, nil, false)
	repo := &classRepoSvcMock{}
	sessions := &classSessionRepoSvcMock{}
	svc := newClassService(repo, sessions, nil)

	sessions.On("FindByID", ctx, sessionID).Return(&domain.ClassSession{ID: sessionID, ClassID: classID, Name: "Old"}, nil).Once()
	repo.On("FindByID", ctx, classID).Return(&domain.Class{ID: classID, UserID: teacherID}, nil).Once()
	_, err := svc.UpdateSession(ctx, sessionID, domain.UpdateClassSessionDTO{Name: &newName})
	assert.ErrorIs(t, err, domain.ErrForbidden)

	teacherCtx := classCtx(teacherID, nil, false)
	sessions.On("FindByID", teacherCtx, sessionID).Return(&domain.ClassSession{ID: sessionID, ClassID: classID, Name: "Old"}, nil).Once()
	repo.On("FindByID", teacherCtx, classID).Return(&domain.Class{ID: classID, UserID: teacherID}, nil).Once()
	sessions.On("Update", teacherCtx, mock.MatchedBy(func(session *domain.ClassSession) bool {
		return session.ID == sessionID && session.Name == newName && session.StartTime.Equal(start)
	})).Return(nil).Once()

	updated, err := svc.UpdateSession(teacherCtx, sessionID, domain.UpdateClassSessionDTO{Name: &newName, StartTime: &start})
	assert.NoError(t, err)
	assert.Equal(t, newName, updated.Name)
}

func TestClassAdminMethodsRequireAdminAndDefaultPagination(t *testing.T) {
	repo := &classRepoSvcMock{}
	sessions := &classSessionRepoSvcMock{}
	svc := newClassService(repo, sessions, nil)
	nonAdmin := classCtx(uuid.New(), nil, false)
	admin := classCtx(uuid.New(), nil, true)
	classID := uuid.New()
	sessionID := uuid.New()

	_, _, err := svc.AdminList(nonAdmin, domain.AdminListClassesQuery{})
	assert.ErrorIs(t, err, domain.ErrForbidden)
	_, err = svc.AdminCreate(nonAdmin, domain.AdminCreateClassDTO{})
	assert.ErrorIs(t, err, domain.ErrForbidden)
	assert.ErrorIs(t, svc.AdminHardDelete(nonAdmin, classID), domain.ErrForbidden)
	_, _, err = svc.AdminListSessions(nonAdmin, domain.AdminListClassSessionsQuery{})
	assert.ErrorIs(t, err, domain.ErrForbidden)
	assert.ErrorIs(t, svc.AdminHardDeleteSession(nonAdmin, sessionID), domain.ErrForbidden)

	repo.On("AdminList", admin, mock.MatchedBy(func(q domain.AdminListClassesQuery) bool {
		return q.ListParams.Page == 1 && q.ListParams.PageSize == domain.DefaultPageSize
	})).Return([]domain.Class{}, int64(0), nil)
	_, _, err = svc.AdminList(admin, domain.AdminListClassesQuery{ListParams: domain.ListParams{Page: -3}})
	assert.NoError(t, err)

	orgID := uuid.New()
	teacherID := uuid.New()
	repo.On("Create", admin, mock.MatchedBy(func(class *domain.Class) bool {
		return class.OrganizationID == orgID && class.UserID == teacherID && class.Name == "Admin Class"
	})).Return(nil)
	created, err := svc.AdminCreate(admin, domain.AdminCreateClassDTO{OrganizationID: orgID, UserID: teacherID, Name: "Admin Class"})
	assert.NoError(t, err)
	assert.Equal(t, "Admin Class", created.Name)

	repo.On("HardDelete", admin, classID).Return(nil)
	assert.NoError(t, svc.AdminHardDelete(admin, classID))

	sessions.On("AdminList", admin, mock.MatchedBy(func(q domain.AdminListClassSessionsQuery) bool {
		return q.ListParams.Page == 1 && q.ListParams.PageSize == domain.DefaultPageSize
	})).Return([]domain.ClassSession{}, int64(0), nil)
	_, _, err = svc.AdminListSessions(admin, domain.AdminListClassSessionsQuery{})
	assert.NoError(t, err)

	sessions.On("HardDelete", admin, sessionID).Return(nil)
	assert.NoError(t, svc.AdminHardDeleteSession(admin, sessionID))
}

// chatCtx builds a caller context for a chat-enabled org (Pro tier has the chat
// feature) so the ProvisionConversation feature gate passes for non-admins.
func chatCtx(userID uuid.UUID, orgID uuid.UUID, perms ...domain.PermissionName) context.Context {
	p := make([]string, 0, len(perms))
	for _, perm := range perms {
		p = append(p, string(perm))
	}
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID:      userID,
		OrgID:       &orgID,
		Permissions: p,
		Ent:         domain.PlanCatalog[domain.PlanKey(domain.TierPro, 50)],
	})
}

func TestProvisionConversationCreatesAndLinksForOwningTeacher(t *testing.T) {
	repo := &classRepoSvcMock{}
	members := &classMemberRepoSvcMock{}
	chat := &chatProvisionerMock{}
	svc := newClassService(repo, nil, members, chat)

	teacherID := uuid.New()
	orgID := uuid.New()
	classID := uuid.New()
	studentA, studentB := uuid.New(), uuid.New()
	newConvID := uuid.New()
	ctx := chatCtx(teacherID, orgID) // no perms: authorized purely by ownership

	repo.On("FindByID", ctx, classID).Return(&domain.Class{
		ID: classID, OrganizationID: orgID, UserID: teacherID, Name: "Algebra",
	}, nil)
	members.On("ListAllByClass", ctx, classID).Return([]domain.ClassMember{
		{UserID: studentA}, {UserID: studentB},
	}, nil)
	chat.On("CreateForClass", ctx, mock.MatchedBy(func(in domain.ProvisionClassChatDTO) bool {
		return in.OrganizationID == orgID &&
			in.CreatorID == teacherID &&
			in.Type == domain.ConversationTypeGroup &&
			in.Name == "Algebra" && // defaults to class name when dto.Name empty
			len(in.MemberIDs) == 2
	})).Return(&domain.Conversation{ID: newConvID, Type: domain.ConversationTypeGroup, Name: "Algebra"}, nil)
	repo.On("Update", ctx, mock.MatchedBy(func(cl *domain.Class) bool {
		return cl.ConversationID != nil && *cl.ConversationID == newConvID
	})).Return(nil)

	conv, err := svc.ProvisionConversation(ctx, classID, domain.ProvisionClassConversationDTO{Type: domain.ConversationTypeGroup})
	assert.NoError(t, err)
	assert.Equal(t, newConvID, conv.ID)
	chat.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestProvisionConversationAddsNonTeacherManagerAsMember(t *testing.T) {
	repo := &classRepoSvcMock{}
	members := &classMemberRepoSvcMock{}
	chat := &chatProvisionerMock{}
	svc := newClassService(repo, nil, members, chat)

	teacherID := uuid.New()
	managerID := uuid.New() // neither the class teacher nor an enrolled student
	orgID := uuid.New()
	classID := uuid.New()
	student := uuid.New()
	newConvID := uuid.New()
	// Manager authorized by an org-wide manage perm, not ownership.
	ctx := chatCtx(managerID, orgID, domain.PermClassesUpdateAny)

	repo.On("FindByID", ctx, classID).Return(&domain.Class{
		ID: classID, OrganizationID: orgID, UserID: teacherID, Name: "Algebra",
	}, nil)
	members.On("ListAllByClass", ctx, classID).Return([]domain.ClassMember{{UserID: student}}, nil)
	chat.On("CreateForClass", ctx, mock.MatchedBy(func(in domain.ProvisionClassChatDTO) bool {
		// Teacher stays the creator/admin; the manager rides along in MemberIDs.
		return in.CreatorID == teacherID &&
			len(in.MemberIDs) == 2 &&
			in.MemberIDs[0] == student && in.MemberIDs[1] == managerID
	})).Return(&domain.Conversation{ID: newConvID, Type: domain.ConversationTypeGroup}, nil)
	repo.On("Update", ctx, mock.Anything).Return(nil)

	_, err := svc.ProvisionConversation(ctx, classID, domain.ProvisionClassConversationDTO{Type: domain.ConversationTypeGroup})
	assert.NoError(t, err)
	chat.AssertExpectations(t)
}

func TestProvisionConversationSyncsWhenAlreadyLinked(t *testing.T) {
	repo := &classRepoSvcMock{}
	members := &classMemberRepoSvcMock{}
	chat := &chatProvisionerMock{}
	svc := newClassService(repo, nil, members, chat)

	teacherID := uuid.New()
	orgID := uuid.New()
	classID := uuid.New()
	convID := uuid.New()
	student := uuid.New()
	ctx := chatCtx(teacherID, orgID)

	repo.On("FindByID", ctx, classID).Return(&domain.Class{
		ID: classID, OrganizationID: orgID, UserID: teacherID, Name: "Algebra", ConversationID: &convID,
	}, nil)
	members.On("ListAllByClass", ctx, classID).Return([]domain.ClassMember{{UserID: student}}, nil)
	chat.On("SyncClassMembers", ctx, convID, mock.MatchedBy(func(ids []uuid.UUID) bool {
		return len(ids) == 1 && ids[0] == student
	})).Return(&domain.Conversation{ID: convID}, nil)

	conv, err := svc.ProvisionConversation(ctx, classID, domain.ProvisionClassConversationDTO{Type: domain.ConversationTypeGroup})
	assert.NoError(t, err)
	assert.Equal(t, convID, conv.ID)
	// No create, no re-link on the sync path.
	chat.AssertNotCalled(t, "CreateForClass", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	chat.AssertExpectations(t)
}

func TestProvisionConversationRecreatesWhenLinkedConversationGone(t *testing.T) {
	repo := &classRepoSvcMock{}
	members := &classMemberRepoSvcMock{}
	chat := &chatProvisionerMock{}
	svc := newClassService(repo, nil, members, chat)

	teacherID := uuid.New()
	orgID := uuid.New()
	classID := uuid.New()
	staleConvID := uuid.New()
	newConvID := uuid.New()
	ctx := chatCtx(teacherID, orgID)

	repo.On("FindByID", ctx, classID).Return(&domain.Class{
		ID: classID, OrganizationID: orgID, UserID: teacherID, Name: "Algebra", ConversationID: &staleConvID,
	}, nil)
	members.On("ListAllByClass", ctx, classID).Return([]domain.ClassMember{}, nil)
	// Linked conversation was deleted: sync reports not-found, so we recreate.
	chat.On("SyncClassMembers", ctx, staleConvID, mock.Anything).Return(nil, domain.ErrNotFound)
	chat.On("CreateForClass", ctx, mock.MatchedBy(func(in domain.ProvisionClassChatDTO) bool {
		return in.CreatorID == teacherID && in.Type == domain.ConversationTypeChannel
	})).Return(&domain.Conversation{ID: newConvID, Type: domain.ConversationTypeChannel}, nil)
	repo.On("Update", ctx, mock.MatchedBy(func(cl *domain.Class) bool {
		return cl.ConversationID != nil && *cl.ConversationID == newConvID
	})).Return(nil)

	conv, err := svc.ProvisionConversation(ctx, classID, domain.ProvisionClassConversationDTO{Type: domain.ConversationTypeChannel})
	assert.NoError(t, err)
	assert.Equal(t, newConvID, conv.ID)
	chat.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestProvisionConversationForbiddenForNonManager(t *testing.T) {
	repo := &classRepoSvcMock{}
	chat := &chatProvisionerMock{}
	svc := newClassService(repo, nil, nil, chat)

	teacherID := uuid.New()
	studentID := uuid.New()
	orgID := uuid.New()
	classID := uuid.New()
	ctx := chatCtx(studentID, orgID) // not the owner, no manage perms

	repo.On("FindByID", ctx, classID).Return(&domain.Class{
		ID: classID, OrganizationID: orgID, UserID: teacherID, Name: "Algebra",
	}, nil)

	_, err := svc.ProvisionConversation(ctx, classID, domain.ProvisionClassConversationDTO{Type: domain.ConversationTypeGroup})
	assert.ErrorIs(t, err, domain.ErrForbidden)
	chat.AssertNotCalled(t, "CreateForClass", mock.Anything, mock.Anything)
}

func TestProvisionConversationRequiresChatFeature(t *testing.T) {
	repo := &classRepoSvcMock{}
	chat := &chatProvisionerMock{}
	svc := newClassService(repo, nil, nil, chat)

	teacherID := uuid.New()
	orgID := uuid.New()
	classID := uuid.New()
	// Free tier lacks the chat feature.
	ctx := domain.WithCaller(context.Background(), domain.Caller{
		UserID: teacherID, OrgID: &orgID,
		Ent: domain.PlanCatalog[domain.PlanKey(domain.TierFree, 50)],
	})

	repo.On("FindByID", ctx, classID).Return(&domain.Class{
		ID: classID, OrganizationID: orgID, UserID: teacherID, Name: "Algebra",
	}, nil)

	_, err := svc.ProvisionConversation(ctx, classID, domain.ProvisionClassConversationDTO{Type: domain.ConversationTypeGroup})
	assert.Error(t, err)
	assert.NotErrorIs(t, err, domain.ErrForbidden)
	chat.AssertNotCalled(t, "CreateForClass", mock.Anything, mock.Anything)
}
