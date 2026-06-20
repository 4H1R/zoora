package domain

const (
	PresetRoleManager = "Manager"
	PresetRoleTeacher = "Teacher"
	PresetRoleStudent = "Student"
)

var ManagerPermissions = []PermissionName{
	// Users
	PermUsersView, PermUsersViewAny, PermUsersCreate,
	PermUsersUpdate, PermUsersUpdateAny, PermUsersDelete, PermUsersDeleteAny,
	PermUsersDisable, PermUsersDisableAny,
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
	// Offlines
	PermOfflinesView, PermOfflinesViewAny, PermOfflinesCreate, PermOfflinesCreateAny,
	PermOfflinesUpdate, PermOfflinesUpdateAny, PermOfflinesDelete, PermOfflinesDeleteAny,
	// Attendance
	PermAttendanceView, PermAttendanceViewAny, PermAttendanceCreate, PermAttendanceCreateAny,
	PermAttendanceUpdate, PermAttendanceUpdateAny, PermAttendanceDelete, PermAttendanceDeleteAny,
	// Gradebook
	PermGradebookView, PermGradebookViewAny, PermGradebookCreate, PermGradebookCreateAny,
	PermGradebookUpdate, PermGradebookUpdateAny, PermGradebookDelete, PermGradebookDeleteAny,
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
	PermOfflinesView, PermOfflinesCreate, PermOfflinesUpdate, PermOfflinesDelete,
	PermAttendanceView, PermAttendanceCreate, PermAttendanceUpdate, PermAttendanceDelete,
}

// StudentPermissions is the default permission set for learners. It grants the
// ability to view their own classes/recordings, join live sessions, and take
// exams + see their own grades/attendance via the self-scoped endpoints.
var StudentPermissions = []PermissionName{
	PermClassesView, PermClassesJoin,
	PermLiveSessionsView, PermLiveSessionsJoin,
	PermRecordingsView, PermRecordingsDownload,
	PermOfflinesView,
	PermMediaView,
	PermQuizzesView, PermQuizzesTake,
	PermGradebookViewOwn,
	PermAttendanceViewOwn,
}
