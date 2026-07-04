package quizzes

import (
	"context"
	"math"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

type gpsPoint struct {
	UserID uuid.UUID
	Lat    float64
	Lng    float64
	Acc    *float64
}

func haversineMeters(lat1, lng1, lat2, lng2 float64) float64 {
	const r = 6371000.0
	toRad := func(d float64) float64 { return d * math.Pi / 180 }
	dLat := toRad(lat2 - lat1)
	dLng := toRad(lng2 - lng1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*math.Sin(dLng/2)*math.Sin(dLng/2)
	return r * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// clusterSameLocation returns, per user, the other users within
// SameLocationMeters. Points whose accuracy is missing or worse than
// SameLocationMaxAccuracy are excluded (too coarse / IP-level).
func clusterSameLocation(points []gpsPoint) map[uuid.UUID][]uuid.UUID {
	out := make(map[uuid.UUID][]uuid.UUID)
	good := make([]gpsPoint, 0, len(points))
	for _, p := range points {
		if p.Acc != nil && *p.Acc < domain.SameLocationMaxAccuracy {
			good = append(good, p)
		}
	}
	for i := 0; i < len(good); i++ {
		for j := i + 1; j < len(good); j++ {
			if haversineMeters(good[i].Lat, good[i].Lng, good[j].Lat, good[j].Lng) <= domain.SameLocationMeters {
				out[good[i].UserID] = append(out[good[i].UserID], good[j].UserID)
				out[good[j].UserID] = append(out[good[j].UserID], good[i].UserID)
			}
		}
	}
	return out
}

// AntiCheatReport builds the advisory review report for every submission of a
// quiz. Requires quiz-management rights: the route perm (PermQuizzesView) is
// org-wide, not quiz-scoped, so authz must be checked here like every other
// teacher-facing method (ListSubmissions, GradeSubmission).
func (s *service) AntiCheatReport(ctx context.Context, quizID uuid.UUID) ([]domain.SubmissionAntiCheatReport, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	quiz, err := s.repo.FindByID(ctx, quizID)
	if err != nil {
		return nil, err
	}
	if !canManageQuiz(caller, quiz) {
		return nil, domain.ErrForbidden
	}

	subs, _, err := s.submissions.ListByQuiz(ctx, quizID, domain.ListSubmissionsQuery{ListParams: domain.ListParams{Page: 1, PageSize: 10000}})
	if err != nil {
		return nil, err
	}

	// question min_seconds lookup
	qidSet := make(map[uuid.UUID]struct{})
	for i := range subs {
		for _, a := range subs[i].Answers {
			qidSet[a.QuestionID] = struct{}{}
		}
	}
	qids := make([]uuid.UUID, 0, len(qidSet))
	for id := range qidSet {
		qids = append(qids, id)
	}
	minByQ := make(map[uuid.UUID]int)
	if len(qids) > 0 {
		qs, err := s.questions.FindByIDs(ctx, qids)
		if err != nil {
			return nil, err
		}
		for _, q := range qs {
			minByQ[q.ID] = q.MinSeconds
		}
	}

	points := make([]gpsPoint, 0, len(subs))
	for i := range subs {
		if subs[i].GPSLat != nil && subs[i].GPSLng != nil {
			points = append(points, gpsPoint{UserID: subs[i].UserID, Lat: *subs[i].GPSLat, Lng: *subs[i].GPSLng, Acc: subs[i].GPSAccuracy})
		}
	}
	clusters := clusterSameLocation(points)

	out := make([]domain.SubmissionAntiCheatReport, 0, len(subs))
	for i := range subs {
		sub := subs[i]
		rep := domain.SubmissionAntiCheatReport{
			SubmissionID:        sub.ID,
			UserID:              sub.UserID,
			TabHiddenCount:      sub.TabHiddenCount,
			TabHiddenSeconds:    sub.TabHiddenSeconds,
			TabFlagged:          sub.TabHiddenCount > domain.TabHiddenWarnCount,
			GPSDenied:           sub.GPSDenied,
			SameLocationUserIDs: clusters[sub.UserID],
		}
		for _, a := range sub.Answers {
			if m := minByQ[a.QuestionID]; m > 0 && a.SpentSeconds < m {
				rep.FastAnswers = append(rep.FastAnswers, domain.FastAnswerFlag{
					QuestionID: a.QuestionID, SpentSeconds: a.SpentSeconds, MinSeconds: m,
				})
			}
		}
		out = append(out, rep)
	}
	return out, nil
}
