package domain

const (
	PresetRoleStaff   = "Staff"
	PresetRoleTeacher = "Teacher"
)

var StaffPermissions = []PermissionName{
	// Users
	PermUsersView, PermUsersViewAny, PermUsersCreate,
	PermUsersUpdate, PermUsersUpdateAny, PermUsersDelete, PermUsersDeleteAny,
	// Roles
	PermRolesView, PermRolesCreate, PermRolesUpdate, PermRolesDelete,
	// Organizations
	PermOrganizationsUpdate,
	// Classes
	PermClassesView, PermClassesViewAny, PermClassesCreate, PermClassesCreateAny,
	PermClassesUpdate, PermClassesUpdateAny, PermClassesDelete, PermClassesDeleteAny, PermClassesJoin,
	// Live Sessions
	PermLiveSessionsView, PermLiveSessionsViewAny, PermLiveSessionsCreate, PermLiveSessionsCreateAny,
	PermLiveSessionsUpdate, PermLiveSessionsUpdateAny, PermLiveSessionsDelete, PermLiveSessionsDeleteAny,
	PermLiveSessionsJoin, PermLiveSessionsJoinAny, PermLiveSessionsManage, PermLiveSessionsManageAny,
	// Recordings
	PermRecordingsView, PermRecordingsViewAny, PermRecordingsCreate, PermRecordingsCreateAny,
	PermRecordingsUpdate, PermRecordingsUpdateAny, PermRecordingsDelete, PermRecordingsDeleteAny, PermRecordingsDownload,
	// Question Banks
	PermQuestionBanksView, PermQuestionBanksViewAny, PermQuestionBanksCreate, PermQuestionBanksCreateAny,
	PermQuestionBanksUpdate, PermQuestionBanksUpdateAny, PermQuestionBanksDelete, PermQuestionBanksDeleteAny,
	// Quizzes
	PermQuizzesView, PermQuizzesViewAny, PermQuizzesCreate, PermQuizzesCreateAny,
	PermQuizzesUpdate, PermQuizzesUpdateAny, PermQuizzesDelete, PermQuizzesDeleteAny,
	// Polls
	PermPollsView, PermPollsViewAny, PermPollsCreate, PermPollsCreateAny,
	PermPollsUpdate, PermPollsUpdateAny, PermPollsDelete, PermPollsDeleteAny,
	// Chats
	PermChatsView, PermChatsViewAny, PermChatsCreate, PermChatsCreateAny,
	PermChatsUpdate, PermChatsUpdateAny, PermChatsDelete, PermChatsDeleteAny,
	PermChatsWrite, PermChatsManage,
	// Media
	PermMediaView, PermMediaViewAny, PermMediaCreate, PermMediaCreateAny,
	PermMediaDelete, PermMediaDeleteAny,
	// Practices
	PermPracticesView, PermPracticesViewAny, PermPracticesCreate, PermPracticesCreateAny,
	PermPracticesUpdate, PermPracticesUpdateAny, PermPracticesDelete, PermPracticesDeleteAny,
	PermPracticesSubmit, PermPracticesGrade,
	// Attendance
	PermAttendanceView, PermAttendanceViewAny, PermAttendanceCreate, PermAttendanceCreateAny,
	PermAttendanceUpdate, PermAttendanceUpdateAny, PermAttendanceDelete, PermAttendanceDeleteAny,
}

var TeacherPermissions = []PermissionName{
	PermLiveSessionsCreate, PermLiveSessionsView, PermLiveSessionsManage, PermLiveSessionsJoin,
	PermRecordingsView, PermRecordingsDownload,
	PermClassesView, PermClassesCreate,
	PermClassesUpdate, PermClassesDelete, PermClassesJoin,
	PermUsersView, PermUsersViewAny,
	PermQuizzesCreate, PermQuizzesUpdate, PermQuizzesView,
	PermQuestionBanksCreate, PermQuestionBanksUpdate, PermQuestionBanksView,
	PermPollsCreate, PermPollsUpdate, PermPollsView,
	PermAttendanceView, PermAttendanceCreate, PermAttendanceUpdate, PermAttendanceDelete,
}
