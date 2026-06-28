package domain

const (
	PresetRoleManager = "Manager"
	PresetRoleTeacher = "Teacher"
	PresetRoleStudent = "Student"
)

var ManagerPermissions = []PermissionName{
	PermUsersView, PermUsersViewAny, PermUsersCreate,
	PermUsersUpdate, PermUsersUpdateAny, PermUsersDelete, PermUsersDeleteAny,
	PermUsersDisable, PermUsersDisableAny,
	PermRolesView, PermRolesCreate, PermRolesUpdate, PermRolesDelete,
	PermOrganizationsUpdate,
	PermClassesView, PermClassesViewAny, PermClassesCreate, PermClassesCreateAny,
	PermClassesUpdate, PermClassesUpdateAny, PermClassesDelete, PermClassesDeleteAny, PermClassesJoin,
	PermLiveSessionsView, PermLiveSessionsViewAny, PermLiveSessionsCreate,
	PermLiveSessionsUpdate, PermLiveSessionsUpdateAny,
	PermLiveSessionsJoin, PermLiveSessionsJoinAny, PermLiveSessionsManage, PermLiveSessionsManageAny,
	PermQuestionBanksView, PermQuestionBanksViewAny, PermQuestionBanksCreate, PermQuestionBanksCreateAny,
	PermQuestionBanksUpdate, PermQuestionBanksUpdateAny, PermQuestionBanksDelete, PermQuestionBanksDeleteAny,
	PermQuizzesView, PermQuizzesViewAny, PermQuizzesCreate,
	PermQuizzesUpdate, PermQuizzesUpdateAny, PermQuizzesDelete, PermQuizzesDeleteAny,
	PermPollsView, PermPollsCreate,
	PermPollsUpdate, PermPollsUpdateAny, PermPollsDelete,
	PermChatsView, PermChatsCreate, PermChatsUpdate, PermChatsDelete,
	PermChatsWrite, PermChatsManage,
	PermMediaView, PermMediaCreate, PermMediaDelete, PermMediaDeleteAny,
	PermPracticesView, PermPracticesViewAny, PermPracticesCreate, PermPracticesCreateAny,
	PermPracticesUpdate, PermPracticesUpdateAny, PermPracticesDelete, PermPracticesDeleteAny,
	PermPracticesSubmit, PermPracticesGrade,
	PermOfflinesView, PermOfflinesViewAny, PermOfflinesCreate, PermOfflinesCreateAny,
	PermOfflinesUpdate, PermOfflinesUpdateAny, PermOfflinesDelete, PermOfflinesDeleteAny,
	PermAttendanceView, PermAttendanceViewAny, PermAttendanceCreate, PermAttendanceCreateAny,
	PermAttendanceUpdate, PermAttendanceUpdateAny, PermAttendanceDelete,
	PermGradebookView, PermGradebookViewAny, PermGradebookCreate,
	PermGradebookUpdate, PermGradebookUpdateAny, PermGradebookDelete, PermGradebookDeleteAny,
}

var TeacherPermissions = []PermissionName{
	PermLiveSessionsCreate, PermLiveSessionsView, PermLiveSessionsManage, PermLiveSessionsJoin,
	PermClassesView, PermClassesCreate,
	PermClassesUpdate, PermClassesDelete, PermClassesJoin,
	PermUsersView, PermUsersViewAny,
	PermQuizzesCreate, PermQuizzesUpdate, PermQuizzesView,
	PermQuestionBanksCreate, PermQuestionBanksUpdate, PermQuestionBanksView,
	PermPollsCreate, PermPollsUpdate, PermPollsView,
	PermChatsView, PermChatsWrite, PermChatsManage,
	PermMediaView, PermMediaCreate,
	PermOfflinesView, PermOfflinesCreate, PermOfflinesUpdate, PermOfflinesDelete,
	PermPracticesView, PermPracticesCreate, PermPracticesUpdate, PermPracticesDelete, PermPracticesGrade,
	PermAttendanceView, PermAttendanceCreate, PermAttendanceUpdate, PermAttendanceDelete,
	PermGradebookView, PermGradebookCreate, PermGradebookUpdate, PermGradebookDelete,
}

// StudentPermissions is the default permission set for learners. It grants the
// ability to view their own classes, join live sessions, take exams, and see
// their own grades/attendance — relation-scoped to "own" by the authz resolver.
var StudentPermissions = []PermissionName{
	PermClassesView, PermClassesJoin,
	PermLiveSessionsView, PermLiveSessionsJoin,
	PermOfflinesView,
	PermMediaView,
	PermPollsView,
	PermChatsView, PermChatsWrite,
	PermQuizzesView, PermQuizzesTake,
	PermPracticesView, PermPracticesSubmit,
	PermGradebookView,
	PermAttendanceView,
}
