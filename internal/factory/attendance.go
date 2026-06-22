package factory

import (
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func NewAttendance(orgID, classID, sessionID, userID uuid.UUID, opts ...func(*domain.Attendance)) *domain.Attendance {
	statuses := []domain.AttendanceStatus{
		domain.AttendanceStatusPresent,
		domain.AttendanceStatusAbsent,
		domain.AttendanceStatusLate,
		domain.AttendanceStatusExcused,
	}
	a := &domain.Attendance{
		OrganizationID: orgID,
		ClassID:        classID,
		ClassSessionID: sessionID,
		UserID:         userID,
		Status:         statuses[nextID()%uint64(len(statuses))],
		IsAutoMarked:   fake.Bool(),
		Remarks:        fakeSentence(6),
	}
	for _, o := range opts {
		o(a)
	}
	return a
}
