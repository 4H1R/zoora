package domain

// PermissionName represents a permission string in the system.
type PermissionName string

const (
	PermUsersView              PermissionName = "users:view"
	PermUsersViewAny           PermissionName = "users:view_any"
	PermUsersCreate            PermissionName = "users:create"
	PermUsersUpdate            PermissionName = "users:update"
	PermUsersUpdateAny         PermissionName = "users:update_any"
	PermUsersDelete            PermissionName = "users:delete"
	PermUsersDeleteAny         PermissionName = "users:delete_any"
	PermUsersDisable           PermissionName = "users:disable"
	PermUsersDisableAny        PermissionName = "users:disable_any"
	PermRolesView              PermissionName = "roles:view"
	PermRolesCreate            PermissionName = "roles:create"
	PermRolesUpdate            PermissionName = "roles:update"
	PermRolesDelete            PermissionName = "roles:delete"
	PermOrganizationsUpdate    PermissionName = "organizations:update"
	PermClassesView            PermissionName = "classes:view"
	PermClassesViewAny         PermissionName = "classes:view_any"
	PermClassesCreate          PermissionName = "classes:create"
	PermClassesCreateAny       PermissionName = "classes:create_any"
	PermClassesUpdate          PermissionName = "classes:update"
	PermClassesUpdateAny       PermissionName = "classes:update_any"
	PermClassesDelete          PermissionName = "classes:delete"
	PermClassesDeleteAny       PermissionName = "classes:delete_any"
	PermClassesJoin            PermissionName = "classes:join"
	PermLiveSessionsView       PermissionName = "live_sessions:view"
	PermLiveSessionsViewAny    PermissionName = "live_sessions:view_any"
	PermLiveSessionsCreate     PermissionName = "live_sessions:create"
	PermLiveSessionsUpdate     PermissionName = "live_sessions:update"
	PermLiveSessionsUpdateAny  PermissionName = "live_sessions:update_any"
	PermLiveSessionsJoin       PermissionName = "live_sessions:join"
	PermLiveSessionsJoinAny    PermissionName = "live_sessions:join_any"
	PermLiveSessionsManage     PermissionName = "live_sessions:manage"
	PermLiveSessionsManageAny  PermissionName = "live_sessions:manage_any"
	PermQuestionBanksView      PermissionName = "question_banks:view"
	PermQuestionBanksViewAny   PermissionName = "question_banks:view_any"
	PermQuestionBanksCreate    PermissionName = "question_banks:create"
	PermQuestionBanksCreateAny PermissionName = "question_banks:create_any"
	PermQuestionBanksUpdate    PermissionName = "question_banks:update"
	PermQuestionBanksUpdateAny PermissionName = "question_banks:update_any"
	PermQuestionBanksDelete    PermissionName = "question_banks:delete"
	PermQuestionBanksDeleteAny PermissionName = "question_banks:delete_any"
	PermQuizzesView            PermissionName = "quizzes:view"
	PermQuizzesViewAny         PermissionName = "quizzes:view_any"
	PermQuizzesCreate          PermissionName = "quizzes:create"
	PermQuizzesUpdate          PermissionName = "quizzes:update"
	PermQuizzesUpdateAny       PermissionName = "quizzes:update_any"
	PermQuizzesDelete          PermissionName = "quizzes:delete"
	PermQuizzesDeleteAny       PermissionName = "quizzes:delete_any"
	PermQuizzesTake            PermissionName = "quizzes:take"
	PermPollsView              PermissionName = "polls:view"
	PermPollsCreate            PermissionName = "polls:create"
	PermPollsUpdate            PermissionName = "polls:update"
	PermPollsUpdateAny         PermissionName = "polls:update_any"
	PermPollsDelete            PermissionName = "polls:delete"
	PermChatsView              PermissionName = "chats:view"
	PermChatsCreate            PermissionName = "chats:create"
	PermChatsUpdate            PermissionName = "chats:update"
	PermChatsDelete            PermissionName = "chats:delete"
	PermChatsWrite             PermissionName = "chats:write"
	PermChatsManage            PermissionName = "chats:manage"
	PermMediaView              PermissionName = "media:view"
	PermMediaViewAny           PermissionName = "media:view_any"
	PermMediaCreate            PermissionName = "media:create"
	PermMediaDelete            PermissionName = "media:delete"
	PermMediaDeleteAny         PermissionName = "media:delete_any"
	PermPracticesView          PermissionName = "practices:view"
	PermPracticesViewAny       PermissionName = "practices:view_any"
	PermPracticesCreate        PermissionName = "practices:create"
	PermPracticesCreateAny     PermissionName = "practices:create_any"
	PermPracticesUpdate        PermissionName = "practices:update"
	PermPracticesUpdateAny     PermissionName = "practices:update_any"
	PermPracticesDelete        PermissionName = "practices:delete"
	PermPracticesDeleteAny     PermissionName = "practices:delete_any"
	PermPracticesSubmit        PermissionName = "practices:submit"
	PermPracticesGrade         PermissionName = "practices:grade"
	PermGradebookView          PermissionName = "gradebook:view"
	PermGradebookViewAny       PermissionName = "gradebook:view_any"
	PermGradebookCreate        PermissionName = "gradebook:create"
	PermGradebookUpdate        PermissionName = "gradebook:update"
	PermGradebookUpdateAny     PermissionName = "gradebook:update_any"
	PermGradebookDelete        PermissionName = "gradebook:delete"
	PermGradebookDeleteAny     PermissionName = "gradebook:delete_any"
	PermOfflinesView           PermissionName = "offlines:view"
	PermOfflinesViewAny        PermissionName = "offlines:view_any"
	PermOfflinesCreate         PermissionName = "offlines:create"
	PermOfflinesCreateAny      PermissionName = "offlines:create_any"
	PermOfflinesUpdate         PermissionName = "offlines:update"
	PermOfflinesUpdateAny      PermissionName = "offlines:update_any"
	PermOfflinesDelete         PermissionName = "offlines:delete"
	PermOfflinesDeleteAny      PermissionName = "offlines:delete_any"
	PermAttendanceView         PermissionName = "attendance:view"
	PermAttendanceViewAny      PermissionName = "attendance:view_any"
	PermAttendanceCreate       PermissionName = "attendance:create"
	PermAttendanceCreateAny    PermissionName = "attendance:create_any"
	PermAttendanceUpdate       PermissionName = "attendance:update"
	PermAttendanceUpdateAny    PermissionName = "attendance:update_any"
	PermAttendanceDelete       PermissionName = "attendance:delete"
	PermNotificationsSend      PermissionName = "notifications:send"
	PermNotificationsSendAny   PermissionName = "notifications:send_any"
)

// AllPermissions is the authoritative list of every permission in the system.
// Seeder reads this to populate the permissions table.
var AllPermissions = []PermissionName{
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
	PermQuizzesUpdate, PermQuizzesUpdateAny, PermQuizzesDelete, PermQuizzesDeleteAny, PermQuizzesTake,
	PermPollsView, PermPollsCreate,
	PermPollsUpdate, PermPollsUpdateAny, PermPollsDelete,
	PermChatsView, PermChatsCreate, PermChatsUpdate, PermChatsDelete,
	PermChatsWrite, PermChatsManage,
	PermMediaView, PermMediaViewAny, PermMediaCreate, PermMediaDelete, PermMediaDeleteAny,
	PermPracticesView, PermPracticesViewAny, PermPracticesCreate, PermPracticesCreateAny,
	PermPracticesUpdate, PermPracticesUpdateAny, PermPracticesDelete, PermPracticesDeleteAny,
	PermPracticesSubmit, PermPracticesGrade,
	PermGradebookView, PermGradebookViewAny, PermGradebookCreate,
	PermGradebookUpdate, PermGradebookUpdateAny, PermGradebookDelete, PermGradebookDeleteAny,
	PermOfflinesView, PermOfflinesViewAny, PermOfflinesCreate, PermOfflinesCreateAny,
	PermOfflinesUpdate, PermOfflinesUpdateAny, PermOfflinesDelete, PermOfflinesDeleteAny,
	PermAttendanceView, PermAttendanceViewAny, PermAttendanceCreate, PermAttendanceCreateAny,
	PermAttendanceUpdate, PermAttendanceUpdateAny, PermAttendanceDelete,
	PermNotificationsSend, PermNotificationsSendAny,
}
