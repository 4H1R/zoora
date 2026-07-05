package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QuizRuleType string

const (
	QuizRuleTypeManual QuizRuleType = "manual"
	QuizRuleTypeRandom QuizRuleType = "random"
)

func (t QuizRuleType) Valid() bool {
	switch t {
	case QuizRuleTypeManual, QuizRuleTypeRandom:
		return true
	}
	return false
}

type Quiz struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	OrganizationID   uuid.UUID `gorm:"type:uuid;not null;index" json:"organization_id"`
	UserID           uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User             *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ClassID          uuid.UUID `gorm:"type:uuid;not null;index" json:"class_id"`
	Class            *Class    `gorm:"foreignKey:ClassID" json:"class,omitempty"`
	Title            string    `gorm:"not null" json:"title"`
	Description      string    `json:"description"`
	DurationMinutes  int       `gorm:"not null" json:"duration_minutes"`
	TotalScore       float64   `gorm:"not null;default:0" json:"total_score"`
	NoBackNavigation bool      `gorm:"not null;default:false" json:"no_back_navigation"`
	ShuffleQuestions bool      `gorm:"not null;default:false" json:"shuffle_questions"`

	// Anti-cheat config toggles. Class A (enforced): ShuffleOptions. Class B
	// (advisory signals): TrackTabSwitches, RequireGPS. Class C (frontend-only
	// deterrents): DisableCopyPaste, DisableRightClickShortcuts.
	ShuffleOptions             bool `gorm:"not null;default:false" json:"shuffle_options"`
	TrackTabSwitches           bool `gorm:"not null;default:false" json:"track_tab_switches"`
	RequireGPS                 bool `gorm:"not null;default:false" json:"require_gps"`
	DisableCopyPaste           bool `gorm:"not null;default:false" json:"disable_copy_paste"`
	DisableRightClickShortcuts bool `gorm:"not null;default:false" json:"disable_right_click_shortcuts"`

	// Quiz-wide negative-marking override (Layer 2b). Fills gaps for questions
	// (manual and random) lacking their own setting.
	NegativeMarkMode NegativeMarkMode `gorm:"type:varchar(20);not null;default:'none'" json:"negative_mark_mode"`
	NegativeValue    float64          `gorm:"not null;default:0" json:"negative_value"`
	WrongsPerPoint   int              `gorm:"not null;default:0" json:"wrongs_per_point"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// QuizQuestionNegativeOverride is a per-question negative-marking override
// (Layer 2a) stored on a QuizRule.
type QuizQuestionNegativeOverride struct {
	QuestionID     uuid.UUID        `json:"question_id"`
	Mode           NegativeMarkMode `json:"mode"`
	NegativeValue  float64          `json:"negative_value"`
	WrongsPerPoint int              `json:"wrongs_per_point"`
}

type QuizRule struct {
	ID          uuid.UUID     `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	QuizID      uuid.UUID     `gorm:"type:uuid;not null;index" json:"quiz_id"`
	Quiz        *Quiz         `gorm:"foreignKey:QuizID" json:"quiz,omitempty"`
	Type        QuizRuleType  `gorm:"type:varchar(20);not null" json:"type"`
	BankID      *uuid.UUID    `gorm:"type:uuid" json:"bank_id,omitempty"`
	Bank        *QuestionBank `gorm:"foreignKey:BankID" json:"bank,omitempty"`
	QuestionIDs []uuid.UUID   `gorm:"type:jsonb;serializer:json" json:"question_ids,omitempty"`
	Count       int           `gorm:"not null;default:0" json:"count"`
	IsDynamic   bool          `gorm:"not null;default:false" json:"is_dynamic"`

	// NegativeOverrides holds per-question negative-marking overrides (Layer 2a).
	NegativeOverrides []QuizQuestionNegativeOverride `gorm:"type:jsonb;serializer:json" json:"negative_overrides"`

	// NegativeDefaultMode is the rule-wide negative-marking default (Layer 2-bank)
	// applied to every choice question this rule contributes — manual and random
	// alike. nil keeps each question's own default; "none" forces no penalty.
	NegativeDefaultMode *NegativeMarkMode `gorm:"type:varchar(20)" json:"negative_default_mode,omitempty"`

	// NegativeDefaultValue (per_wrong) and NegativeDefaultWrongsPerPoint
	// (accumulative) are the optional explicit numbers for the rule-wide default.
	// nil means "derive per question from its option count at grade time"; a set
	// value applies as-is to every choice question this rule contributes.
	NegativeDefaultValue          *float64 `gorm:"type:double precision" json:"negative_default_value,omitempty"`
	NegativeDefaultWrongsPerPoint *int     `gorm:"type:int" json:"negative_default_wrongs_per_point,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NegativeDefaultConfig builds the rule-wide default as a NegativeMarkConfig, or
// nil when the rule keeps each question's own default. A zero numeric field
// signals "derive from the question's option count" to ResolveNegativeMark.
func (r QuizRule) NegativeDefaultConfig() *NegativeMarkConfig {
	if r.NegativeDefaultMode == nil {
		return nil
	}
	cfg := &NegativeMarkConfig{Mode: *r.NegativeDefaultMode}
	if r.NegativeDefaultValue != nil {
		cfg.NegativeValue = *r.NegativeDefaultValue
	}
	if r.NegativeDefaultWrongsPerPoint != nil {
		cfg.WrongsPerPoint = *r.NegativeDefaultWrongsPerPoint
	}
	return cfg
}

type QuizRoom struct {
	ID             uuid.UUID     `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	QuizID         uuid.UUID     `gorm:"type:uuid;not null;index" json:"quiz_id"`
	Quiz           *Quiz         `gorm:"foreignKey:QuizID" json:"quiz,omitempty"`
	ClassSessionID uuid.UUID     `gorm:"type:uuid;not null;index" json:"class_session_id"`
	ClassSession   *ClassSession `gorm:"foreignKey:ClassSessionID" json:"class_session,omitempty"`
	StartedAt      *time.Time    `json:"started_at"`
	EndedAt        *time.Time    `json:"ended_at"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

type CreateQuizDTO struct {
	ClassID          uuid.UUID `json:"class_id" binding:"required"`
	Title            string    `json:"title" binding:"required,min=2"`
	Description      string    `json:"description"`
	DurationMinutes  int       `json:"duration_minutes" binding:"required,gt=0"`
	NoBackNavigation bool      `json:"no_back_navigation"`
	ShuffleQuestions bool      `json:"shuffle_questions"`

	ShuffleOptions             bool `json:"shuffle_options"`
	TrackTabSwitches           bool `json:"track_tab_switches"`
	RequireGPS                 bool `json:"require_gps"`
	DisableCopyPaste           bool `json:"disable_copy_paste"`
	DisableRightClickShortcuts bool `json:"disable_right_click_shortcuts"`

	NegativeMarkMode NegativeMarkMode `json:"negative_mark_mode"`
	NegativeValue    float64          `json:"negative_value"`
	WrongsPerPoint   int              `json:"wrongs_per_point"`
}

type UpdateQuizDTO struct {
	Title            *string `json:"title" binding:"omitempty,min=2"`
	Description      *string `json:"description"`
	DurationMinutes  *int    `json:"duration_minutes" binding:"omitempty,gt=0"`
	NoBackNavigation *bool   `json:"no_back_navigation"`
	ShuffleQuestions *bool   `json:"shuffle_questions"`

	ShuffleOptions             *bool `json:"shuffle_options"`
	TrackTabSwitches           *bool `json:"track_tab_switches"`
	RequireGPS                 *bool `json:"require_gps"`
	DisableCopyPaste           *bool `json:"disable_copy_paste"`
	DisableRightClickShortcuts *bool `json:"disable_right_click_shortcuts"`

	NegativeMarkMode *NegativeMarkMode `json:"negative_mark_mode"`
	NegativeValue    *float64          `json:"negative_value"`
	WrongsPerPoint   *int              `json:"wrongs_per_point"`
}

type CreateQuizRuleDTO struct {
	Type                QuizRuleType                   `json:"type" binding:"required,oneof=manual random"`
	BankID              *uuid.UUID                     `json:"bank_id"`
	QuestionIDs         []uuid.UUID                    `json:"question_ids"`
	Count               int                            `json:"count" binding:"gte=0"`
	IsDynamic           bool                           `json:"is_dynamic"`
	NegativeOverrides   []QuizQuestionNegativeOverride `json:"negative_overrides"`
	NegativeDefaultMode *NegativeMarkMode              `json:"negative_default_mode"`
	// Optional explicit numbers for the rule-wide default; nil derives from each
	// question's option count. per_wrong uses Value, accumulative uses WrongsPerPoint.
	NegativeDefaultValue          *float64 `json:"negative_default_value"`
	NegativeDefaultWrongsPerPoint *int     `json:"negative_default_wrongs_per_point"`
}

type UpdateQuizRuleDTO struct {
	Type                *QuizRuleType                  `json:"type" binding:"omitempty,oneof=manual random"`
	BankID              *uuid.UUID                     `json:"bank_id"`
	QuestionIDs         []uuid.UUID                    `json:"question_ids"`
	Count               *int                           `json:"count" binding:"omitempty,gte=0"`
	IsDynamic           *bool                          `json:"is_dynamic"`
	NegativeOverrides   []QuizQuestionNegativeOverride `json:"negative_overrides"`
	NegativeDefaultMode *NegativeMarkMode              `json:"negative_default_mode"`
	// Optional explicit numbers for the rule-wide default; nil derives from each
	// question's option count. per_wrong uses Value, accumulative uses WrongsPerPoint.
	NegativeDefaultValue          *float64 `json:"negative_default_value"`
	NegativeDefaultWrongsPerPoint *int     `json:"negative_default_wrongs_per_point"`
}

type CreateQuizRoomDTO struct {
	ClassSessionID uuid.UUID  `json:"class_session_id" binding:"required"`
	StartedAt      *time.Time `json:"started_at" binding:"required"`
	EndedAt        *time.Time `json:"ended_at" binding:"required"`
}

func (d CreateQuizRoomDTO) Validate() error {
	if d.StartedAt == nil || d.EndedAt == nil {
		return NewValidationError(map[string]string{"window": "started_at and ended_at are required"})
	}
	if !d.EndedAt.After(*d.StartedAt) {
		return NewValidationError(map[string]string{"window": "ended_at must be after started_at"})
	}
	return nil
}

type ListQuizzesQuery struct {
	ClassID        *uuid.UUID `form:"-"`
	ClassSessionID *uuid.UUID `form:"-"`
	IncludeDeleted bool       `form:"include_deleted"`
	ListParams     ListParams `form:"-"`
}

type AdminListQuizzesQuery struct {
	ClassID        *uuid.UUID `form:"-"`
	ClassSessionID *uuid.UUID `form:"-"`
	UserID         *uuid.UUID `form:"-"`
	IncludeDeleted bool       `form:"include_deleted"`
	ListParams     ListParams `form:"-"`
}

type ListQuizRulesQuery struct {
	ListParams ListParams `form:"-"`
}

type ListQuizRoomsQuery struct {
	QuizID     *uuid.UUID `form:"-"`
	ListParams ListParams `form:"-"`
}

type SubmissionStatus string

const (
	SubmissionStatusInProgress SubmissionStatus = "in_progress"
	SubmissionStatusSubmitted  SubmissionStatus = "submitted"
	SubmissionStatusGraded     SubmissionStatus = "graded"
)

func (s SubmissionStatus) Valid() bool {
	switch s {
	case SubmissionStatusInProgress, SubmissionStatusSubmitted, SubmissionStatusGraded:
		return true
	}
	return false
}

type SubmissionAnswer struct {
	QuestionID        uuid.UUID `json:"question_id"`
	SelectedOptionIDs []string  `json:"selected_option_ids,omitempty"`
	Value             string    `json:"value,omitempty"`
	EarnedScore       float64   `json:"earned_score"`
	SpentSeconds      int       `json:"spent_seconds"`

	// Advisory auto-grading signals for descriptive answers, computed at
	// finalize from the question's rubric options and model answer. They never
	// affect EarnedScore/TotalScore and are stripped from student-facing reads
	// — the teacher confirms or overrides them during manual grading.
	SuggestedScore  *float64 `json:"suggested_score,omitempty"`
	MatchedConcepts []string `json:"matched_concepts,omitempty"`
	SimilarityPct   *float64 `json:"similarity_pct,omitempty"`
}

// SubmissionQuestion is one frozen question in a student's submission: the
// question id plus the shuffled display order of its option ids. Built once at
// StartSubmission so reloads/resume are deterministic.
type SubmissionQuestion struct {
	QuestionID    uuid.UUID `json:"question_id"`
	OptionIDOrder []string  `json:"option_id_order,omitempty"`
}

type QuizSubmission struct {
	ID      uuid.UUID          `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	QuizID  uuid.UUID          `gorm:"type:uuid;not null;index;uniqueIndex:idx_quiz_submissions_quiz_user" json:"quiz_id"`
	Quiz    *Quiz              `gorm:"foreignKey:QuizID" json:"quiz,omitempty"`
	UserID  uuid.UUID          `gorm:"type:uuid;not null;index;uniqueIndex:idx_quiz_submissions_quiz_user" json:"user_id"`
	User    *User              `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Status  SubmissionStatus   `gorm:"type:varchar(20);not null;default:'in_progress'" json:"status"`
	Answers []SubmissionAnswer `gorm:"type:jsonb;serializer:json" json:"answers"`

	// QuestionSet is the frozen, per-student ordered question list (with option
	// order) resolved at StartSubmission. QuizRoomID records the room the
	// submission was started in, used to cap the deadline. The GPS/Tab fields are
	// advisory anti-cheat signals.
	QuestionSet      []SubmissionQuestion `gorm:"type:jsonb;serializer:json" json:"question_set"`
	QuizRoomID       *uuid.UUID           `gorm:"type:uuid" json:"quiz_room_id,omitempty"`
	TabHiddenCount   int                  `gorm:"not null;default:0" json:"tab_hidden_count"`
	TabHiddenSeconds int                  `gorm:"not null;default:0" json:"tab_hidden_seconds"`
	GPSLat           *float64             `gorm:"column:gps_lat" json:"gps_lat,omitempty"`
	GPSLng           *float64             `gorm:"column:gps_lng" json:"gps_lng,omitempty"`
	GPSAccuracy      *float64             `gorm:"column:gps_accuracy" json:"gps_accuracy,omitempty"`
	GPSDenied        bool                 `gorm:"not null;default:false" json:"gps_denied"`

	TotalScore  float64    `gorm:"not null;default:0" json:"total_score"`
	StartedAt   time.Time  `gorm:"not null" json:"started_at"`
	SubmittedAt *time.Time `json:"submitted_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// IsRoomOpen returns true when the quiz room window contains now.
// StartedAt/EndedAt define the scheduled availability window.
// A nil EndedAt is treated as open-ended (manual close).
func (r *QuizRoom) IsRoomOpen() bool {
	return r.IsRoomOpenAt(time.Now())
}

func (r *QuizRoom) IsRoomOpenAt(t time.Time) bool {
	if r.StartedAt == nil {
		return false
	}
	if t.Before(*r.StartedAt) {
		return false
	}
	if r.EndedAt != nil && !t.Before(*r.EndedAt) {
		return false
	}
	return true
}

type StartQuizSubmissionDTO struct {
	QuizRoomID  uuid.UUID `json:"quiz_room_id" binding:"required"`
	GPSLat      *float64  `json:"gps_lat"`
	GPSLng      *float64  `json:"gps_lng"`
	GPSAccuracy *float64  `json:"gps_accuracy"`
	GPSDenied   bool      `json:"gps_denied"`
}

type SaveAnswerDTO struct {
	QuestionID        uuid.UUID `json:"question_id" binding:"required"`
	SelectedOptionIDs []string  `json:"selected_option_ids"`
	Value             string    `json:"value"`
	SpentSeconds      int       `json:"spent_seconds" binding:"gte=0"`
	TabHiddenCount    int       `json:"tab_hidden_count" binding:"gte=0"`
	TabHiddenSeconds  int       `json:"tab_hidden_seconds" binding:"gte=0"`
}

type SubmitAnswerDTO struct {
	QuestionID        uuid.UUID `json:"question_id" binding:"required"`
	SelectedOptionIDs []string  `json:"selected_option_ids"`
	Value             string    `json:"value"`
	SpentSeconds      int       `json:"spent_seconds" binding:"gte=0"`
}

type SubmitQuizDTO struct {
	// Answers is optional: with incremental save the final submit may carry zero
	// new answers ("finalize what's saved"). Merged into the saved answers.
	Answers          []SubmitAnswerDTO `json:"answers" binding:"omitempty,dive"`
	TabHiddenCount   int               `json:"tab_hidden_count" binding:"gte=0"`
	TabHiddenSeconds int               `json:"tab_hidden_seconds" binding:"gte=0"`
}

type GradeAnswerDTO struct {
	QuestionID  uuid.UUID `json:"question_id" binding:"required"`
	EarnedScore float64   `json:"earned_score" binding:"gte=0"`
}

type GradeSubmissionDTO struct {
	Grades []GradeAnswerDTO `json:"grades" binding:"required,dive"`
}

type ListSubmissionsQuery struct {
	UserID     *uuid.UUID `form:"-"`
	Status     *string    `form:"status"`
	ListParams ListParams `form:"-"`
}

// SubmissionGracePeriod is the extra time allowed beyond duration_minutes
// before a submission is rejected outright.
const SubmissionGracePeriod = 30 // seconds

type QuizListScope struct {
	All            bool
	OrganizationID *uuid.UUID
	OwnerID        *uuid.UUID
	MemberUserID   *uuid.UUID
	ClassID        *uuid.UUID
	ClassSessionID *uuid.UUID
	IncludeDeleted bool
}

// MyExamState is the derived status of an exam for a specific student.
type MyExamState string

const (
	// MyExamStateUpcoming: a room is scheduled but not open yet, no submission.
	MyExamStateUpcoming MyExamState = "upcoming"
	// MyExamStateOpen: a room is open now and the student has not submitted.
	MyExamStateOpen MyExamState = "open"
	// MyExamStateSubmitted: submitted, awaiting grading.
	MyExamStateSubmitted MyExamState = "submitted"
	// MyExamStateGraded: graded; Score is populated.
	MyExamStateGraded MyExamState = "graded"
)

// MyExamRoom is the room a student should use to take/continue an exam.
type MyExamRoom struct {
	ID             uuid.UUID  `json:"id"`
	ClassSessionID uuid.UUID  `json:"class_session_id"`
	StartedAt      *time.Time `json:"started_at"`
	EndedAt        *time.Time `json:"ended_at"`
	IsOpen         bool       `json:"is_open"`
}

// MyExam is one exam as seen by a student, with availability + their own result.
type MyExam struct {
	QuizID          uuid.UUID   `json:"quiz_id"`
	Title           string      `json:"title"`
	ClassID         uuid.UUID   `json:"class_id"`
	ClassName       string      `json:"class_name"`
	DurationMinutes int         `json:"duration_minutes"`
	TotalScore      float64     `json:"total_score"`
	State           MyExamState `json:"state"`
	// Room is the open room if one exists, else the next upcoming room, else nil.
	Room *MyExamRoom `json:"room,omitempty"`
	// Score is populated only when State == graded.
	Score *float64 `json:"score,omitempty"`
	// SubmittedAt is populated when the student has submitted.
	SubmittedAt *time.Time `json:"submitted_at,omitempty"`
}

type QuizRepository interface {
	Create(ctx context.Context, quiz *Quiz) error
	FindByID(ctx context.Context, id uuid.UUID) (*Quiz, error)
	Update(ctx context.Context, quiz *Quiz) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, scope QuizListScope, p ListParams) ([]Quiz, int64, error)
	// ListByMemberWithRooms returns quizzes whose class the given user is a
	// member of, with Class preloaded, ordered by created_at desc.
	ListByMemberWithRooms(ctx context.Context, userID uuid.UUID, p ListParams) ([]Quiz, int64, error)

	HardDelete(ctx context.Context, id uuid.UUID) error
	FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*Quiz, error)
	AdminList(ctx context.Context, q AdminListQuizzesQuery) ([]Quiz, int64, error)
}

type QuizRuleRepository interface {
	Create(ctx context.Context, rule *QuizRule) error
	FindByID(ctx context.Context, id uuid.UUID) (*QuizRule, error)
	Update(ctx context.Context, rule *QuizRule) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByQuiz(ctx context.Context, quizID uuid.UUID, p ListParams) ([]QuizRule, int64, error)
}

type QuizRoomRepository interface {
	Create(ctx context.Context, room *QuizRoom) error
	FindByID(ctx context.Context, id uuid.UUID) (*QuizRoom, error)
	Update(ctx context.Context, room *QuizRoom) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByQuiz(ctx context.Context, quizID uuid.UUID, p ListParams) ([]QuizRoom, int64, error)
	ListBySessionID(ctx context.Context, sessionID uuid.UUID) ([]QuizRoom, error)
	FindOpenByQuizID(ctx context.Context, quizID uuid.UUID) (*QuizRoom, error)
}

type QuizSubmissionRepository interface {
	Create(ctx context.Context, sub *QuizSubmission) error
	FindByID(ctx context.Context, id uuid.UUID) (*QuizSubmission, error)
	Update(ctx context.Context, sub *QuizSubmission) error
	FindByQuizAndUser(ctx context.Context, quizID, userID uuid.UUID) (*QuizSubmission, error)
	ListByQuiz(ctx context.Context, quizID uuid.UUID, q ListSubmissionsQuery) ([]QuizSubmission, int64, error)
}

// Anti-cheat review thresholds. Fixed constants: raw values are always stored
// and shown; a threshold only decides which values get flagged.
const (
	TabHiddenWarnCount      = 3   // flag if a student left the tab more than this
	SameLocationMeters      = 50  // two students within this distance are flagged
	SameLocationMaxAccuracy = 100 // ignore coarse coords (accuracy worse than this)
)

// FastAnswerFlag marks an answer submitted faster than the question's declared
// min_seconds. Advisory only.
type FastAnswerFlag struct {
	QuestionID   uuid.UUID `json:"question_id"`
	SpentSeconds int       `json:"spent_seconds"`
	MinSeconds   int       `json:"min_seconds"`
}

// SubmissionAntiCheatReport is the advisory, teacher-facing anti-cheat summary
// for one submission. Never asserts guilt — the teacher reviews and decides.
type SubmissionAntiCheatReport struct {
	SubmissionID     uuid.UUID        `json:"submission_id"`
	UserID           uuid.UUID        `json:"user_id"`
	TabHiddenCount   int              `json:"tab_hidden_count"`
	TabHiddenSeconds int              `json:"tab_hidden_seconds"`
	TabFlagged       bool             `json:"tab_flagged"`
	GPSDenied        bool             `json:"gps_denied"`
	FastAnswers      []FastAnswerFlag `json:"fast_answers"`
	// SameLocationUserIDs lists other students in this quiz whose recorded GPS is
	// within SameLocationMeters (both accuracy < SameLocationMaxAccuracy).
	SameLocationUserIDs []uuid.UUID `json:"same_location_user_ids"`
}

type QuizService interface {
	Create(ctx context.Context, dto CreateQuizDTO) (*Quiz, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Quiz, error)
	Update(ctx context.Context, id uuid.UUID, dto UpdateQuizDTO) (*Quiz, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, q ListQuizzesQuery) ([]Quiz, int64, error)
	ListMine(ctx context.Context, p ListParams) ([]MyExam, int64, error)

	CreateRule(ctx context.Context, quizID uuid.UUID, dto CreateQuizRuleDTO) (*QuizRule, error)
	GetRule(ctx context.Context, id uuid.UUID) (*QuizRule, error)
	UpdateRule(ctx context.Context, id uuid.UUID, dto UpdateQuizRuleDTO) (*QuizRule, error)
	DeleteRule(ctx context.Context, id uuid.UUID) error
	ListRules(ctx context.Context, quizID uuid.UUID, q ListQuizRulesQuery) ([]QuizRule, int64, error)

	CreateRoom(ctx context.Context, quizID uuid.UUID, dto CreateQuizRoomDTO) (*QuizRoom, error)
	GetRoom(ctx context.Context, id uuid.UUID) (*QuizRoom, error)
	StartRoom(ctx context.Context, id uuid.UUID) (*QuizRoom, error)
	EndRoom(ctx context.Context, id uuid.UUID) (*QuizRoom, error)
	ListRooms(ctx context.Context, quizID uuid.UUID, q ListQuizRoomsQuery) ([]QuizRoom, int64, error)

	ListQuestionsForTaking(ctx context.Context, quizID uuid.UUID) ([]Question, error)
	StartSubmission(ctx context.Context, quizID uuid.UUID, dto StartQuizSubmissionDTO) (*QuizSubmission, error)
	SaveAnswer(ctx context.Context, submissionID uuid.UUID, dto SaveAnswerDTO) error
	SubmitQuiz(ctx context.Context, submissionID uuid.UUID, dto SubmitQuizDTO) (*QuizSubmission, error)
	GetSubmission(ctx context.Context, id uuid.UUID) (*QuizSubmission, error)
	ListSubmissions(ctx context.Context, quizID uuid.UUID, q ListSubmissionsQuery) ([]QuizSubmission, int64, error)
	GradeSubmission(ctx context.Context, id uuid.UUID, dto GradeSubmissionDTO) (*QuizSubmission, error)
	AntiCheatReport(ctx context.Context, quizID uuid.UUID) ([]SubmissionAntiCheatReport, error)

	AdminList(ctx context.Context, q AdminListQuizzesQuery) ([]Quiz, int64, error)
	AdminHardDelete(ctx context.Context, id uuid.UUID) error
}
