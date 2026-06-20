package domain

// PermissionName represents a permission string in the system.
type PermissionName string

const (
	// Users
	PermUsersView      PermissionName = "users:view"
	PermUsersViewAny   PermissionName = "users:view_any"
	PermUsersCreate    PermissionName = "users:create"
	PermUsersUpdate    PermissionName = "users:update"
	PermUsersUpdateAny PermissionName = "users:update_any"
	PermUsersDelete    PermissionName = "users:delete"
	PermUsersDeleteAny PermissionName = "users:delete_any"
	// Roles
	PermRolesView   PermissionName = "roles:view"
	PermRolesCreate PermissionName = "roles:create"
	PermRolesUpdate PermissionName = "roles:update"
	PermRolesDelete PermissionName = "roles:delete"
	// Organizations
	PermOrganizationsUpdate PermissionName = "organizations:update"
	// Classes
	PermClassesView      PermissionName = "classes:view"
	PermClassesViewAny   PermissionName = "classes:view_any"
	PermClassesCreate    PermissionName = "classes:create"
	PermClassesCreateAny PermissionName = "classes:create_any"
	PermClassesUpdate    PermissionName = "classes:update"
	PermClassesUpdateAny PermissionName = "classes:update_any"
	PermClassesDelete    PermissionName = "classes:delete"
	PermClassesDeleteAny PermissionName = "classes:delete_any"
	PermClassesJoin      PermissionName = "classes:join"
	// Live Sessions
	PermLiveSessionsView      PermissionName = "livesessions:view"
	PermLiveSessionsViewAny   PermissionName = "livesessions:view_any"
	PermLiveSessionsCreate    PermissionName = "livesessions:create"
	PermLiveSessionsCreateAny PermissionName = "livesessions:create_any"
	PermLiveSessionsUpdate    PermissionName = "livesessions:update"
	PermLiveSessionsUpdateAny PermissionName = "livesessions:update_any"
	PermLiveSessionsDelete    PermissionName = "livesessions:delete"
	PermLiveSessionsDeleteAny PermissionName = "livesessions:delete_any"
	PermLiveSessionsJoin      PermissionName = "livesessions:join"
	PermLiveSessionsJoinAny   PermissionName = "livesessions:join_any"
	PermLiveSessionsManage    PermissionName = "livesessions:manage"
	PermLiveSessionsManageAny PermissionName = "livesessions:manage_any"
	// Recordings
	PermRecordingsView      PermissionName = "recordings:view"
	PermRecordingsViewAny   PermissionName = "recordings:view_any"
	PermRecordingsCreate    PermissionName = "recordings:create"
	PermRecordingsCreateAny PermissionName = "recordings:create_any"
	PermRecordingsUpdate    PermissionName = "recordings:update"
	PermRecordingsUpdateAny PermissionName = "recordings:update_any"
	PermRecordingsDelete    PermissionName = "recordings:delete"
	PermRecordingsDeleteAny PermissionName = "recordings:delete_any"
	PermRecordingsDownload  PermissionName = "recordings:download"
	// Question Banks
	PermQuestionBanksView      PermissionName = "question_banks:view"
	PermQuestionBanksViewAny   PermissionName = "question_banks:view_any"
	PermQuestionBanksCreate    PermissionName = "question_banks:create"
	PermQuestionBanksCreateAny PermissionName = "question_banks:create_any"
	PermQuestionBanksUpdate    PermissionName = "question_banks:update"
	PermQuestionBanksUpdateAny PermissionName = "question_banks:update_any"
	PermQuestionBanksDelete    PermissionName = "question_banks:delete"
	PermQuestionBanksDeleteAny PermissionName = "question_banks:delete_any"
	// Quizzes
	PermQuizzesView      PermissionName = "quizzes:view"
	PermQuizzesViewAny   PermissionName = "quizzes:view_any"
	PermQuizzesCreate    PermissionName = "quizzes:create"
	PermQuizzesCreateAny PermissionName = "quizzes:create_any"
	PermQuizzesUpdate    PermissionName = "quizzes:update"
	PermQuizzesUpdateAny PermissionName = "quizzes:update_any"
	PermQuizzesDelete    PermissionName = "quizzes:delete"
	PermQuizzesDeleteAny PermissionName = "quizzes:delete_any"
	PermQuizzesTake      PermissionName = "quizzes:take"
	// Polls
	PermPollsView      PermissionName = "polls:view"
	PermPollsViewAny   PermissionName = "polls:view_any"
	PermPollsCreate    PermissionName = "polls:create"
	PermPollsCreateAny PermissionName = "polls:create_any"
	PermPollsUpdate    PermissionName = "polls:update"
	PermPollsUpdateAny PermissionName = "polls:update_any"
	PermPollsDelete    PermissionName = "polls:delete"
	PermPollsDeleteAny PermissionName = "polls:delete_any"
	// Chats
	PermChatsView      PermissionName = "chats:view"
	PermChatsViewAny   PermissionName = "chats:view_any"
	PermChatsCreate    PermissionName = "chats:create"
	PermChatsCreateAny PermissionName = "chats:create_any"
	PermChatsUpdate    PermissionName = "chats:update"
	PermChatsUpdateAny PermissionName = "chats:update_any"
	PermChatsDelete    PermissionName = "chats:delete"
	PermChatsDeleteAny PermissionName = "chats:delete_any"
	PermChatsWrite     PermissionName = "chats:write"
	PermChatsManage    PermissionName = "chats:manage"
	// Media
	PermMediaView      PermissionName = "media:view"
	PermMediaViewAny   PermissionName = "media:view_any"
	PermMediaCreate    PermissionName = "media:create"
	PermMediaCreateAny PermissionName = "media:create_any"
	PermMediaDelete    PermissionName = "media:delete"
	PermMediaDeleteAny PermissionName = "media:delete_any"
	// Practices
	PermPracticesView      PermissionName = "practices:view"
	PermPracticesViewAny   PermissionName = "practices:view_any"
	PermPracticesCreate    PermissionName = "practices:create"
	PermPracticesCreateAny PermissionName = "practices:create_any"
	PermPracticesUpdate    PermissionName = "practices:update"
	PermPracticesUpdateAny PermissionName = "practices:update_any"
	PermPracticesDelete    PermissionName = "practices:delete"
	PermPracticesDeleteAny PermissionName = "practices:delete_any"
	PermPracticesSubmit    PermissionName = "practices:submit"
	PermPracticesGrade     PermissionName = "practices:grade"
	// Gradebook
	PermGradebookView      PermissionName = "gradebook:view"
	PermGradebookViewAny   PermissionName = "gradebook:view_any"
	PermGradebookCreate    PermissionName = "gradebook:create"
	PermGradebookCreateAny PermissionName = "gradebook:create_any"
	PermGradebookUpdate    PermissionName = "gradebook:update"
	PermGradebookUpdateAny PermissionName = "gradebook:update_any"
	PermGradebookDelete    PermissionName = "gradebook:delete"
	PermGradebookDeleteAny PermissionName = "gradebook:delete_any"
	PermGradebookViewOwn   PermissionName = "gradebook:view_own"
	// Offlines
	PermOfflinesView      PermissionName = "offlines:view"
	PermOfflinesViewAny   PermissionName = "offlines:view_any"
	PermOfflinesCreate    PermissionName = "offlines:create"
	PermOfflinesCreateAny PermissionName = "offlines:create_any"
	PermOfflinesUpdate    PermissionName = "offlines:update"
	PermOfflinesUpdateAny PermissionName = "offlines:update_any"
	PermOfflinesDelete    PermissionName = "offlines:delete"
	PermOfflinesDeleteAny PermissionName = "offlines:delete_any"
	// Attendance
	PermAttendanceView      PermissionName = "attendance:view"
	PermAttendanceViewAny   PermissionName = "attendance:view_any"
	PermAttendanceCreate    PermissionName = "attendance:create"
	PermAttendanceCreateAny PermissionName = "attendance:create_any"
	PermAttendanceUpdate    PermissionName = "attendance:update"
	PermAttendanceUpdateAny PermissionName = "attendance:update_any"
	PermAttendanceDelete    PermissionName = "attendance:delete"
	PermAttendanceDeleteAny PermissionName = "attendance:delete_any"
	PermAttendanceViewOwn   PermissionName = "attendance:view_own"
)

// AllPermissions is the authoritative list of every permission in the system.
// Seeder reads this to populate the permissions table.
var AllPermissions = []PermissionName{
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
	PermQuizzesUpdate, PermQuizzesUpdateAny, PermQuizzesDelete, PermQuizzesDeleteAny, PermQuizzesTake,
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
	// Gradebook
	PermGradebookView, PermGradebookViewAny, PermGradebookCreate, PermGradebookCreateAny,
	PermGradebookUpdate, PermGradebookUpdateAny, PermGradebookDelete, PermGradebookDeleteAny, PermGradebookViewOwn,
	// Offlines
	PermOfflinesView, PermOfflinesViewAny, PermOfflinesCreate, PermOfflinesCreateAny,
	PermOfflinesUpdate, PermOfflinesUpdateAny, PermOfflinesDelete, PermOfflinesDeleteAny,
	// Attendance
	PermAttendanceView, PermAttendanceViewAny, PermAttendanceCreate, PermAttendanceCreateAny,
	PermAttendanceUpdate, PermAttendanceUpdateAny, PermAttendanceDelete, PermAttendanceDeleteAny, PermAttendanceViewOwn,
}
