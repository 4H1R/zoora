package database

import (
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestPostgresErrorClassifiers(t *testing.T) {
	unique := &pgconn.PgError{Code: pgUniqueViolation}
	fk := &pgconn.PgError{Code: pgForeignKeyViolation}
	notNull := &pgconn.PgError{Code: pgNotNullViolation}

	if !IsUniqueViolation(unique) {
		t.Fatal("IsUniqueViolation returned false for unique violation")
	}
	if IsUniqueViolation(fk) || IsUniqueViolation(notNull) || IsUniqueViolation(nil) {
		t.Fatal("IsUniqueViolation returned true for non-unique error")
	}

	if !IsForeignKeyViolation(fk) {
		t.Fatal("IsForeignKeyViolation returned false for foreign-key violation")
	}
	if IsForeignKeyViolation(unique) || IsForeignKeyViolation(notNull) || IsForeignKeyViolation(nil) {
		t.Fatal("IsForeignKeyViolation returned true for non-foreign-key error")
	}
}

func TestPostgresErrorClassifiersUnwrapWrappedErrors(t *testing.T) {
	err := fmt.Errorf("repo create: %w", &pgconn.PgError{Code: pgUniqueViolation})

	if !IsUniqueViolation(err) {
		t.Fatal("IsUniqueViolation returned false for wrapped PgError")
	}
	if IsForeignKeyViolation(err) {
		t.Fatal("IsForeignKeyViolation returned true for wrapped unique PgError")
	}
}
