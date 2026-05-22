package factory

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func NewGradebookColumn(classID uuid.UUID, colType domain.GradebookColumnType, opts ...func(*domain.GradebookColumn)) *domain.GradebookColumn {
	id := nextID()
	c := &domain.GradebookColumn{
		ClassID:    classID,
		Title:      fmt.Sprintf("Column %d", id),
		Type:       colType,
		OrderIndex: int(id % 100),
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

func NewGradebookCell(columnID, studentID uuid.UUID, value string) *domain.GradebookCell {
	return &domain.GradebookCell{
		ColumnID:  columnID,
		StudentID: studentID,
		Value:     value,
	}
}
