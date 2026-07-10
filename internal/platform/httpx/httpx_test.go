package httpx

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/domain"
)

func init() { gin.SetMode(gin.TestMode) }

func newCtx(t *testing.T, rawQuery string) *gin.Context {
	t.Helper()
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequest("GET", "/?"+rawQuery, nil)
	c.Request = req
	return c
}

func TestParseUUIDQuery_Present(t *testing.T) {
	want := uuid.New()
	c := newCtx(t, "class_id="+want.String())

	got, err := ParseUUIDQuery(c, "class_id")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, want, *got)
}

func TestParseUUIDQuery_Missing(t *testing.T) {
	c := newCtx(t, "")
	got, err := ParseUUIDQuery(c, "class_id")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestParseUUIDQuery_EmptyValue(t *testing.T) {
	c := newCtx(t, "class_id=")
	got, err := ParseUUIDQuery(c, "class_id")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestParseUUIDQuery_Malformed(t *testing.T) {
	c := newCtx(t, "class_id=not-a-uuid")
	got, err := ParseUUIDQuery(c, "class_id")
	require.Error(t, err)
	assert.Nil(t, got)
}

func TestBindUUIDQueries_SetsAllPresent(t *testing.T) {
	classID := uuid.New()
	sessionID := uuid.New()
	c := newCtx(t, "class_id="+classID.String()+"&class_session_id="+sessionID.String())

	var (
		gotClass   *uuid.UUID
		gotSession *uuid.UUID
		gotCreator *uuid.UUID
	)
	err := BindUUIDQueries(c, map[string]**uuid.UUID{
		"class_id":         &gotClass,
		"class_session_id": &gotSession,
		"creator_id":       &gotCreator,
	})
	require.NoError(t, err)
	require.NotNil(t, gotClass)
	require.NotNil(t, gotSession)
	assert.Equal(t, classID, *gotClass)
	assert.Equal(t, sessionID, *gotSession)
	assert.Nil(t, gotCreator)
}

func TestBindUUIDQueries_MalformedReturnsValidationError(t *testing.T) {
	c := newCtx(t, "class_id=bogus")
	var target *uuid.UUID
	err := BindUUIDQueries(c, map[string]**uuid.UUID{"class_id": &target})
	require.Error(t, err)

	var ve *domain.ValidationError
	require.True(t, errors.As(err, &ve), "expected *domain.ValidationError, got %T", err)
	_, ok := ve.Fields["class_id"]
	assert.True(t, ok, "fields should be keyed by query param name")
	assert.Nil(t, target)
}

func TestBindUUIDQueries_NoFieldsNoOp(t *testing.T) {
	c := newCtx(t, "class_id="+uuid.New().String())
	err := BindUUIDQueries(c, map[string]**uuid.UUID{})
	assert.NoError(t, err)
}

func TestUsernameValidator(t *testing.T) {
	ok := []string{"ali", "ali_r", "sara.k", "abc123", "a_b.c"}
	bad := []string{"ab", "Ali", "ali r", "ali-r", "toolong_username_exceeding_thirty_chars"}
	for _, s := range ok {
		if !usernameRe.MatchString(s) {
			t.Errorf("expected %q valid", s)
		}
	}
	for _, s := range bad {
		if usernameRe.MatchString(s) {
			t.Errorf("expected %q invalid", s)
		}
	}
}
