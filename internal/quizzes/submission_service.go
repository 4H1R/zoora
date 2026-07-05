package quizzes

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// ListQuestionsForTaking returns the ordered question list for a caller. Quiz
// managers get a fresh rule-composed preview; enrolled students get their
// frozen, per-submission set with option order applied and answer keys stripped
// (option scores, short_answer correct values) so the caller can never derive
// the grading rubric.
func (s *service) ListQuestionsForTaking(ctx context.Context, quizID uuid.UUID) ([]domain.Question, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	quiz, err := s.repo.FindByID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	// Managers preview the composed rule output (no submission, fresh random picks).
	if canManageQuiz(caller, quiz) {
		return s.listQuestionsComposed(ctx, quiz)
	}

	// Students get their frozen, per-submission set.
	enrolled, err := s.members.Exists(ctx, quiz.ClassID, caller.UserID)
	if err != nil {
		return nil, err
	}
	if !enrolled {
		return nil, domain.ErrForbidden
	}
	// ErrNotFound here means the student hasn't started — client must POST /submissions first.
	sub, err := s.submissions.FindByQuizAndUser(ctx, quizID, caller.UserID)
	if err != nil {
		return nil, err
	}

	cfgFor, err := s.negativeCfgResolver(ctx, quiz)
	if err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, len(sub.QuestionSet))
	for i, sq := range sub.QuestionSet {
		ids[i] = sq.QuestionID
	}
	loaded, err := s.questions.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	byID := make(map[uuid.UUID]domain.Question, len(loaded))
	for i := range loaded {
		byID[loaded[i].ID] = loaded[i]
	}

	// Resume: with no back-navigation, skip already-answered questions.
	answered := make(map[uuid.UUID]struct{})
	if quiz.NoBackNavigation {
		for _, a := range sub.Answers {
			answered[a.QuestionID] = struct{}{}
		}
	}

	out := make([]domain.Question, 0, len(sub.QuestionSet))
	for _, sq := range sub.QuestionSet {
		q, ok := byID[sq.QuestionID]
		if !ok {
			continue
		}
		if _, done := answered[sq.QuestionID]; done {
			continue
		}
		sanitized := sanitizeQuestionForTaking(q)
		if sanitized.Type == domain.QuestionTypeChoice {
			if cfg := cfgFor(&q); cfg.Mode != domain.NegativeMarkNone {
				sanitized.NegativeConfig = &cfg
			}
			if len(sq.OptionIDOrder) > 0 {
				sanitized.Options = reorderOptions(sanitized.Options, sq.OptionIDOrder)
			}
		}
		out = append(out, sanitized)
	}
	return out, nil
}

// listQuestionsComposed is the manager preview path: compose rules + bank
// questions server-side, strip answer keys, attach negative-mark config.
func (s *service) listQuestionsComposed(ctx context.Context, quiz *domain.Quiz) ([]domain.Question, error) {
	rules, _, err := s.rules.ListByQuiz(ctx, quiz.ID, domain.ListParams{Page: 1, PageSize: 10000})
	if err != nil {
		return nil, err
	}

	quizWide := domain.NegativeMarkConfig{Mode: quiz.NegativeMarkMode, NegativeValue: quiz.NegativeValue, WrongsPerPoint: quiz.WrongsPerPoint}

	seen := make(map[uuid.UUID]struct{})
	out := make([]domain.Question, 0)
	for _, r := range rules {
		ovByQ := make(map[uuid.UUID]domain.QuizQuestionNegativeOverride, len(r.NegativeOverrides))
		for _, ov := range r.NegativeOverrides {
			ovByQ[ov.QuestionID] = ov
		}
		picked, err := s.pickQuestionsForRule(ctx, r)
		if err != nil {
			return nil, err
		}
		for i := range picked {
			if _, dup := seen[picked[i].ID]; dup {
				continue
			}
			seen[picked[i].ID] = struct{}{}
			sanitized := sanitizeQuestionForTaking(picked[i])
			if sanitized.Type == domain.QuestionTypeChoice {
				var ovPtr *domain.QuizQuestionNegativeOverride
				if ov, ok := ovByQ[picked[i].ID]; ok {
					ovPtr = &ov
				}
				cfg := domain.ResolveNegativeMark(picked[i], ovPtr, r.NegativeDefaultConfig(), quizWide)
				if cfg.Mode != domain.NegativeMarkNone {
					sanitized.NegativeConfig = &cfg
				}
			}
			out = append(out, sanitized)
		}
	}
	return out, nil
}

// negativeCfgResolver loads the quiz rules and returns the per-question
// negative-marking resolver. Shared by grading (SubmitQuiz / lazy finalize) and
// the student take-view so both resolve the same layered config (per-question
// override → rule default → bank default → quiz-wide).
func (s *service) negativeCfgResolver(ctx context.Context, quiz *domain.Quiz) (func(q *domain.Question) domain.NegativeMarkConfig, error) {
	rules, _, err := s.rules.ListByQuiz(ctx, quiz.ID, domain.ListParams{Page: 1, PageSize: 10000})
	if err != nil {
		return nil, err
	}
	quizWide := domain.NegativeMarkConfig{Mode: quiz.NegativeMarkMode, NegativeValue: quiz.NegativeValue, WrongsPerPoint: quiz.WrongsPerPoint}
	overrideByQ := make(map[uuid.UUID]domain.QuizQuestionNegativeOverride)
	// Rule-wide defaults resolved by question: manual rules attach their default
	// to explicit QuestionIDs; random rules attach theirs to every question in
	// the referenced bank (question IDs are unknown until grade time).
	defaultByQ := make(map[uuid.UUID]*domain.NegativeMarkConfig)
	defaultByBank := make(map[uuid.UUID]*domain.NegativeMarkConfig)
	for _, r := range rules {
		for _, ov := range r.NegativeOverrides {
			overrideByQ[ov.QuestionID] = ov
		}
		rd := r.NegativeDefaultConfig()
		if rd == nil {
			continue
		}
		switch r.Type {
		case domain.QuizRuleTypeManual:
			for _, qid := range r.QuestionIDs {
				defaultByQ[qid] = rd
			}
		case domain.QuizRuleTypeRandom:
			if r.BankID != nil {
				defaultByBank[*r.BankID] = rd
			}
		}
	}
	cfgFor := func(q *domain.Question) domain.NegativeMarkConfig {
		var ovPtr *domain.QuizQuestionNegativeOverride
		if ov, ok := overrideByQ[q.ID]; ok {
			ovPtr = &ov
		}
		ruleDefault := defaultByQ[q.ID]
		if ruleDefault == nil {
			ruleDefault = defaultByBank[q.BankID]
		}
		return domain.ResolveNegativeMark(*q, ovPtr, ruleDefault, quizWide)
	}
	return cfgFor, nil
}

// buildQuestionSet resolves the quiz rules into a concrete, ordered, per-student
// question set. Question order and option order are shuffled deterministically
// from the submission id when the quiz enables the respective toggle. The set is
// frozen on the submission so reloads and resume are stable.
func (s *service) buildQuestionSet(ctx context.Context, quiz *domain.Quiz, submissionID uuid.UUID) ([]domain.SubmissionQuestion, error) {
	rules, _, err := s.rules.ListByQuiz(ctx, quiz.ID, domain.ListParams{Page: 1, PageSize: 10000})
	if err != nil {
		return nil, err
	}
	seen := make(map[uuid.UUID]struct{})
	set := make([]domain.SubmissionQuestion, 0)
	for _, r := range rules {
		picked, err := s.pickQuestionsForRule(ctx, r)
		if err != nil {
			return nil, err
		}
		for i := range picked {
			q := picked[i]
			if _, dup := seen[q.ID]; dup {
				continue
			}
			seen[q.ID] = struct{}{}
			var order []string
			if q.Type == domain.QuestionTypeChoice {
				ids := make([]string, len(q.Options))
				for j, o := range q.Options {
					ids[j] = o.ID
				}
				if quiz.ShuffleOptions {
					// Salt with the question id: one salt for all questions would give
					// every same-option-count question the identical permutation.
					ids = shuffleStrings(submissionID, "opts:"+q.ID.String(), ids)
				}
				order = ids
			}
			set = append(set, domain.SubmissionQuestion{QuestionID: q.ID, OptionIDOrder: order})
		}
	}
	if quiz.ShuffleQuestions {
		r := rand.New(rand.NewSource(seedFrom(submissionID, "questions"))) //nolint:gosec // non-crypto
		r.Shuffle(len(set), func(i, j int) { set[i], set[j] = set[j], set[i] })
	}
	return set, nil
}

// reorderOptions returns opts ordered by the given id order; ids not present are
// appended in their original order (defensive against option edits mid-exam).
func reorderOptions(opts []domain.QuestionOption, order []string) []domain.QuestionOption {
	byID := make(map[string]domain.QuestionOption, len(opts))
	for _, o := range opts {
		byID[o.ID] = o
	}
	out := make([]domain.QuestionOption, 0, len(opts))
	used := make(map[string]struct{}, len(opts))
	for _, id := range order {
		if o, ok := byID[id]; ok {
			out = append(out, o)
			used[id] = struct{}{}
		}
	}
	for _, o := range opts {
		if _, ok := used[o.ID]; !ok {
			out = append(out, o)
		}
	}
	return out
}

func (s *service) pickQuestionsForRule(ctx context.Context, r domain.QuizRule) ([]domain.Question, error) {
	switch r.Type {
	case domain.QuizRuleTypeManual:
		if len(r.QuestionIDs) == 0 {
			return nil, nil
		}
		qs, err := s.questions.FindByIDs(ctx, r.QuestionIDs)
		if err != nil {
			return nil, err
		}
		byID := make(map[uuid.UUID]domain.Question, len(qs))
		for i := range qs {
			byID[qs[i].ID] = qs[i]
		}
		ordered := make([]domain.Question, 0, len(r.QuestionIDs))
		for _, id := range r.QuestionIDs {
			if q, ok := byID[id]; ok {
				ordered = append(ordered, q)
			}
		}
		return ordered, nil
	case domain.QuizRuleTypeRandom:
		if r.BankID == nil || r.Count == 0 {
			return nil, nil
		}
		return s.questions.RandomByBank(ctx, *r.BankID, r.Count)
	}
	return nil, nil
}

// sanitizeQuestionForTaking removes answer-key data from a question. Choice
// options keep id+value but lose score; short_answer/descriptive lose options
// entirely because their options hold correct answers.
func sanitizeQuestionForTaking(q domain.Question) domain.Question {
	switch q.Type {
	case domain.QuestionTypeChoice:
		multi := q.IsMultiSelect()
		opts := make([]domain.QuestionOption, len(q.Options))
		for i, o := range q.Options {
			opts[i] = domain.QuestionOption{ID: o.ID, Value: o.Value, ImageMediaID: o.ImageMediaID}
		}
		q.Options = opts
		q.IsMultiSelectFlag = &multi
	case domain.QuestionTypeShortAnswer, domain.QuestionTypeDescriptive:
		q.Options = []domain.QuestionOption{}
	}
	q.ModelAnswer = ""
	return q
}

func (s *service) StartSubmission(ctx context.Context, quizID uuid.UUID, dto domain.StartQuizSubmissionDTO) (*domain.QuizSubmission, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}

	quiz, err := s.repo.FindByID(ctx, quizID)
	if err != nil {
		return nil, err
	}

	enrolled, err := s.members.Exists(ctx, quiz.ClassID, caller.UserID)
	if err != nil {
		return nil, err
	}
	if !enrolled {
		return nil, domain.ErrForbidden
	}

	room, err := s.rooms.FindByID(ctx, dto.QuizRoomID)
	if err != nil {
		return nil, err
	}
	if room.QuizID != quizID {
		return nil, domain.ErrNotFound
	}
	if !room.IsRoomOpen() {
		return nil, domain.NewValidationError(map[string]string{"quiz_room": "room is not open"})
	}

	existing, err := s.submissions.FindByQuizAndUser(ctx, quizID, caller.UserID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrConflict
	}

	// A client that simply omits the GPS fields must not pass as "granted, no
	// data" — the cheapest bypass. Require coords or an explicit denial.
	if quiz.RequireGPS && !dto.GPSDenied && (dto.GPSLat == nil || dto.GPSLng == nil) {
		return nil, domain.NewValidationError(map[string]string{"gps": "gps coordinates or gps_denied are required for this quiz"})
	}

	// Generate the id first: buildQuestionSet seeds its shuffle on it. Persist the
	// room used at start so finalizeIfExpired can cap the deadline correctly.
	sub := &domain.QuizSubmission{
		ID:          uuid.Must(uuid.NewV7()),
		QuizID:      quizID,
		UserID:      caller.UserID,
		QuizRoomID:  &room.ID,
		Status:      domain.SubmissionStatusInProgress,
		Answers:     []domain.SubmissionAnswer{},
		StartedAt:   time.Now(),
		GPSLat:      dto.GPSLat,
		GPSLng:      dto.GPSLng,
		GPSAccuracy: dto.GPSAccuracy,
		GPSDenied:   dto.GPSDenied,
	}
	set, err := s.buildQuestionSet(ctx, quiz, sub.ID)
	if err != nil {
		return nil, err
	}
	sub.QuestionSet = set
	if err := s.submissions.Create(ctx, sub); err != nil {
		return nil, err
	}

	s.logger.Info("quiz submission started",
		"submission_id", sub.ID.String(),
		"quiz_id", quizID.String(),
		"user_id", caller.UserID.String(),
	)
	return sub, nil
}

// SaveAnswer upserts a single answer into an in-progress submission without
// grading it (grading happens only at submit/finalize so a client cannot probe
// correctness). It also records the latest client-reported tab-visibility
// counters. Answers for questions outside the frozen set are rejected.
//
// Known, accepted v1 race: this is a read-modify-write on the whole answers
// jsonb column; two concurrent saves can drop one answer. Autosave is sequential
// in the client, so no locking in v1.
func (s *service) SaveAnswer(ctx context.Context, submissionID uuid.UUID, dto domain.SaveAnswerDTO) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	sub, err := s.submissions.FindByID(ctx, submissionID)
	if err != nil {
		return err
	}
	if sub.UserID != caller.UserID {
		return domain.ErrForbidden
	}
	sub, _, err = s.finalizeIfExpired(ctx, sub)
	if err != nil {
		return err
	}
	if sub.Status != domain.SubmissionStatusInProgress {
		return domain.ErrConflict
	}
	inSet := false
	for _, sq := range sub.QuestionSet {
		if sq.QuestionID == dto.QuestionID {
			inSet = true
			break
		}
	}
	if !inSet {
		return domain.NewValidationError(map[string]string{"question_id": "not part of this submission"})
	}

	found := false
	for i := range sub.Answers {
		if sub.Answers[i].QuestionID == dto.QuestionID {
			sub.Answers[i].SelectedOptionIDs = dto.SelectedOptionIDs
			sub.Answers[i].Value = dto.Value
			sub.Answers[i].SpentSeconds = dto.SpentSeconds
			found = true
			break
		}
	}
	if !found {
		sub.Answers = append(sub.Answers, domain.SubmissionAnswer{
			QuestionID:        dto.QuestionID,
			SelectedOptionIDs: dto.SelectedOptionIDs,
			Value:             dto.Value,
			SpentSeconds:      dto.SpentSeconds,
		})
	}
	if dto.TabHiddenCount > sub.TabHiddenCount {
		sub.TabHiddenCount = dto.TabHiddenCount
	}
	if dto.TabHiddenSeconds > sub.TabHiddenSeconds {
		sub.TabHiddenSeconds = dto.TabHiddenSeconds
	}
	return s.submissions.Update(ctx, sub)
}

func (s *service) SubmitQuiz(ctx context.Context, submissionID uuid.UUID, dto domain.SubmitQuizDTO) (*domain.QuizSubmission, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}

	sub, err := s.submissions.FindByID(ctx, submissionID)
	if err != nil {
		return nil, err
	}
	if sub.UserID != caller.UserID {
		return nil, domain.ErrForbidden
	}

	// Past the deadline the submission is finalized from the saved answers and
	// the DTO is ignored — this is the server-enforced auto-submit.
	sub, finalized, err := s.finalizeIfExpired(ctx, sub)
	if err != nil {
		return nil, err
	}
	if finalized {
		stripSuggestions(sub)
		return sub, nil
	}
	if sub.Status != domain.SubmissionStatusInProgress {
		return nil, domain.ErrConflict
	}

	quiz, err := s.repo.FindByID(ctx, sub.QuizID)
	if err != nil {
		return nil, err
	}

	// Within the deadline a room must still be open (existing behavior).
	if _, err := s.rooms.FindOpenByQuizID(ctx, sub.QuizID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.NewValidationError(map[string]string{"quiz_room": "room is not open"})
		}
		return nil, err
	}

	// Merge new answers into the saved ones (same upsert as SaveAnswer),
	// dropping anything outside the frozen question set.
	inSet := make(map[uuid.UUID]struct{}, len(sub.QuestionSet))
	for _, sq := range sub.QuestionSet {
		inSet[sq.QuestionID] = struct{}{}
	}
	for _, a := range dto.Answers {
		if _, ok := inSet[a.QuestionID]; !ok {
			continue
		}
		found := false
		for i := range sub.Answers {
			if sub.Answers[i].QuestionID == a.QuestionID {
				sub.Answers[i].SelectedOptionIDs = a.SelectedOptionIDs
				sub.Answers[i].Value = a.Value
				sub.Answers[i].SpentSeconds = a.SpentSeconds
				found = true
				break
			}
		}
		if !found {
			sub.Answers = append(sub.Answers, domain.SubmissionAnswer{
				QuestionID:        a.QuestionID,
				SelectedOptionIDs: a.SelectedOptionIDs,
				Value:             a.Value,
				SpentSeconds:      a.SpentSeconds,
			})
		}
	}
	if dto.TabHiddenCount > sub.TabHiddenCount {
		sub.TabHiddenCount = dto.TabHiddenCount
	}
	if dto.TabHiddenSeconds > sub.TabHiddenSeconds {
		sub.TabHiddenSeconds = dto.TabHiddenSeconds
	}

	if err := s.gradeAndFinalize(ctx, sub, quiz); err != nil {
		return nil, err
	}

	s.logger.Info("quiz submitted",
		"submission_id", sub.ID.String(),
		"quiz_id", sub.QuizID.String(),
		"user_id", caller.UserID.String(),
		"total_score", sub.TotalScore,
	)
	// Students never see the advisory grading signals on their own submission.
	stripSuggestions(sub)
	return sub, nil
}

// gradeAndFinalize grades every saved answer on sub, sets totals/status, and
// persists. Shared by SubmitQuiz and lazy deadline finalize so both paths grade
// identically from the incrementally saved answers.
func (s *service) gradeAndFinalize(ctx context.Context, sub *domain.QuizSubmission, quiz *domain.Quiz) error {
	cfgFor, err := s.negativeCfgResolver(ctx, quiz)
	if err != nil {
		return err
	}
	ids := make([]uuid.UUID, 0, len(sub.Answers))
	for _, a := range sub.Answers {
		ids = append(ids, a.QuestionID)
	}
	qMap := make(map[uuid.UUID]*domain.Question, len(ids))
	if len(ids) > 0 {
		questions, err := s.questions.FindByIDs(ctx, ids)
		if err != nil {
			return err
		}
		for i := range questions {
			qMap[questions[i].ID] = &questions[i]
		}
	}
	var total float64
	for i := range sub.Answers {
		if q, ok := qMap[sub.Answers[i].QuestionID]; ok {
			sub.Answers[i].EarnedScore = gradeAnswer(q, sub.Answers[i], cfgFor(q))
			if q.Type == domain.QuestionTypeDescriptive {
				sub.Answers[i].SuggestedScore, sub.Answers[i].MatchedConcepts, sub.Answers[i].SimilarityPct =
					suggestDescriptive(q, sub.Answers[i].Value)
			}
		}
		total += sub.Answers[i].EarnedScore
	}
	now := time.Now()
	sub.TotalScore = total
	sub.SubmittedAt = &now
	sub.Status = domain.SubmissionStatusSubmitted
	return s.submissions.Update(ctx, sub)
}

// finalizeIfExpired finalizes an in-progress submission whose deadline has
// passed, grading it from whatever answers were saved incrementally. The bool
// reports whether finalization happened on this call. No-op if already finalized
// or still within the deadline (+grace). The deadline uses the room the
// submission was started in (sub.QuizRoomID) — not FindOpenByQuizID, which
// returns ErrNotFound once the room closes and can return a different room when
// a quiz has several.
func (s *service) finalizeIfExpired(ctx context.Context, sub *domain.QuizSubmission) (*domain.QuizSubmission, bool, error) {
	if sub.Status != domain.SubmissionStatusInProgress {
		return sub, false, nil
	}
	quiz, err := s.repo.FindByID(ctx, sub.QuizID)
	if err != nil {
		return nil, false, err
	}
	deadline := sub.StartedAt.Add(time.Duration(quiz.DurationMinutes) * time.Minute)
	if sub.QuizRoomID != nil {
		if room, err := s.rooms.FindByID(ctx, *sub.QuizRoomID); err == nil && room.EndedAt != nil && room.EndedAt.Before(deadline) {
			deadline = *room.EndedAt
		}
	}
	grace := time.Duration(domain.SubmissionGracePeriod) * time.Second
	if time.Now().Before(deadline.Add(grace)) {
		return sub, false, nil
	}
	if err := s.gradeAndFinalize(ctx, sub, quiz); err != nil {
		return nil, false, err
	}
	return sub, true, nil
}

func (s *service) GetSubmission(ctx context.Context, id uuid.UUID) (*domain.QuizSubmission, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}

	sub, err := s.submissions.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// Finalize before the authz check: the side effect is the correct server
	// action regardless of who triggers the read, and the student's own
	// resume/poll path (the common read) early-returns just below.
	sub, _, err = s.finalizeIfExpired(ctx, sub)
	if err != nil {
		return nil, err
	}

	if sub.UserID == caller.UserID {
		stripSuggestions(sub)
		return sub, nil
	}

	quiz, err := s.repo.FindByID(ctx, sub.QuizID)
	if err != nil {
		return nil, err
	}
	if !canManageQuiz(caller, quiz) {
		return nil, domain.ErrForbidden
	}
	return sub, nil
}

func (s *service) ListSubmissions(ctx context.Context, quizID uuid.UUID, q domain.ListSubmissionsQuery) ([]domain.QuizSubmission, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}

	quiz, err := s.repo.FindByID(ctx, quizID)
	if err != nil {
		return nil, 0, err
	}

	manager := canManageQuiz(caller, quiz)
	if !manager {
		userID := caller.UserID
		q.UserID = &userID
	}

	subs, total, err := s.submissions.ListByQuiz(ctx, quizID, q)
	if err != nil {
		return nil, 0, err
	}
	if !manager {
		for i := range subs {
			stripSuggestions(&subs[i])
		}
	}
	return subs, total, nil
}

func (s *service) GradeSubmission(ctx context.Context, id uuid.UUID, dto domain.GradeSubmissionDTO) (*domain.QuizSubmission, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}

	sub, err := s.submissions.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// Auto-finalize a never-submitted-but-expired student so the teacher can grade
	// instead of being blocked.
	sub, _, err = s.finalizeIfExpired(ctx, sub)
	if err != nil {
		return nil, err
	}
	if sub.Status == domain.SubmissionStatusInProgress {
		return nil, domain.NewValidationError(map[string]string{"status": "submission not yet submitted"})
	}

	quiz, err := s.repo.FindByID(ctx, sub.QuizID)
	if err != nil {
		return nil, err
	}
	if !canManageQuiz(caller, quiz) {
		return nil, domain.ErrForbidden
	}

	gradeMap := make(map[uuid.UUID]float64, len(dto.Grades))
	for _, g := range dto.Grades {
		gradeMap[g.QuestionID] = g.EarnedScore
	}

	var totalScore float64
	for i, a := range sub.Answers {
		if score, ok := gradeMap[a.QuestionID]; ok {
			sub.Answers[i].EarnedScore = score
		}
		totalScore += sub.Answers[i].EarnedScore
	}

	sub.TotalScore = totalScore
	sub.Status = domain.SubmissionStatusGraded

	if err := s.submissions.Update(ctx, sub); err != nil {
		return nil, err
	}

	s.logger.Info("quiz submission graded",
		"submission_id", sub.ID.String(),
		"graded_by", caller.UserID.String(),
		"total_score", sub.TotalScore,
	)
	return sub, nil
}

func gradeAnswer(q *domain.Question, a domain.SubmissionAnswer, cfg domain.NegativeMarkConfig) float64 {
	switch q.Type {
	case domain.QuestionTypeChoice:
		return gradeChoice(q.Options, a.SelectedOptionIDs, cfg)
	case domain.QuestionTypeShortAnswer:
		return gradeShortAnswer(q.Options, a.Value)
	case domain.QuestionTypeDescriptive:
		return 0
	}
	return 0
}

// gradeChoice computes earned = positive - penalty. Only the SIGN of an
// option's score matters: Score>0 is correct, Score<=0 is a distractor.
// Penalty is driven by the resolved negative-mark config. Result may be negative.
func gradeChoice(options []domain.QuestionOption, selectedIDs []string, cfg domain.NegativeMarkConfig) float64 {
	optMap := make(map[string]float64, len(options))
	for _, o := range options {
		optMap[o.ID] = o.Score
	}
	var positive float64
	wrongCount := 0
	for _, id := range selectedIDs {
		score, ok := optMap[id]
		if !ok {
			continue
		}
		if score > 0 {
			positive += score
		} else {
			wrongCount++
		}
	}
	return positive - cfg.Penalty(wrongCount)
}

// gradeShortAnswer matches the student's answer against every accepted
// option value and synonym. Pass 1 compares normalized text; pass 2 retries
// spacing-insensitively (ZWNJ/space/attached forms compare equal) and is
// skipped for purely numeric accepted answers so "1 5" never matches "15".
func gradeShortAnswer(options []domain.QuestionOption, value string) float64 {
	normalized := normalizeText(value)
	compact := normalizeCompact(value)
	for _, o := range options {
		if o.Score <= 0 {
			continue
		}
		for _, accepted := range append([]string{o.Value}, o.Synonyms...) {
			want := normalizeText(accepted)
			if want == "" {
				continue
			}
			if want == normalized {
				return o.Score
			}
			if !isNumericAnswer(accepted) && normalizeCompact(accepted) == compact {
				return o.Score
			}
		}
	}
	return 0
}
