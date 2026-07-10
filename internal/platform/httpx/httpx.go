package httpx

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"regexp"
	"strconv"
	"strings"

	"github.com/4H1R/zoora/internal/domain"
)

// RegisterValidators wires custom binding tags onto Gin's default validator.
// Call once at startup.
func RegisterValidators() error {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return errors.New("httpx: gin validator engine is not *validator.Validate")
	}
	if err := v.RegisterValidation("strongpassword", strongPassword); err != nil {
		return err
	}
	return v.RegisterValidation("username", username)
}

func strongPassword(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	return len(s) >= 6
}

var usernameRe = regexp.MustCompile(`^[a-z0-9_.]{3,30}$`)

func username(fl validator.FieldLevel) bool {
	return usernameRe.MatchString(fl.Field().String())
}

func QueryInt(c *gin.Context, key string, defaultVal int) int {
	raw := c.Query(key)
	if raw == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return defaultVal
	}
	return v
}

// BoolQuery parses an optional tri-state bool query param. Returns nil when
// absent/empty or unparseable (treated as "no filter"), otherwise the value.
func BoolQuery(c *gin.Context, key string) *bool {
	raw := c.Query(key)
	if raw == "" {
		return nil
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return nil
	}
	return &v
}

// Pagination pulls offset/limit from query with sane defaults + caps.
func Pagination(c *gin.Context) (offset, limit int) {
	offset = QueryInt(c, "offset", 0)
	limit = min(QueryInt(c, "limit", 20), 100)
	return
}

// ParseUUIDQuery extracts and parses a UUID from query string. Returns (nil, nil)
// when the param is absent or empty. Gin's form binder cannot natively bind
// *uuid.UUID because uuid.UUID's underlying [16]byte triggers Gin's array
// length check (binding/form_mapping.go ~line 302). Handlers must mark such
// fields form:"-" and call this helper.
func ParseUUIDQuery(c *gin.Context, key string) (*uuid.UUID, error) {
	raw := c.Query(key)
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// BindUUIDQueries iterates the provided key→target pairs and parses each one
// from the request query string into *uuid.UUID. Returns the first parse
// failure as a domain.ValidationError keyed by the offending query parameter.
func BindUUIDQueries(c *gin.Context, fields map[string]**uuid.UUID) error {
	for key, target := range fields {
		id, err := ParseUUIDQuery(c, key)
		if err != nil {
			return domain.NewValidationError(map[string]string{key: err.Error()})
		}
		*target = id
	}
	return nil
}

// RequireUUIDParam parses URL param as UUID, stashes it in context under param name.
// Aborts 400 on parse failure.
func RequireUUIDParam(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param(name))
		if err != nil {
			domain.ErrorResponse(c, errors.Join(domain.ErrValidation, err))
			c.Abort()
			return
		}
		c.Set("uuid:"+name, id)
		c.Next()
	}
}

// UUIDParam retrieves a parsed UUID stashed by RequireUUIDParam.
func UUIDParam(c *gin.Context, name string) uuid.UUID {
	v, _ := c.Get("uuid:" + name)
	return v.(uuid.UUID)
}

// BindJSON parses and validates JSON body; on failure, sends 400 and returns false.
// validator.ValidationErrors translate into domain.ValidationError per-field map.
//
// Deprecated: legacy handlers use this; new code should use Bind which returns
// the error for the global error middleware to handle.
func BindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		domain.ErrorResponse(c, toValidationError(err))
		return false
	}
	return true
}

// Bind parses and validates JSON body and returns the error (nil on success).
// Callers should attach the error via c.Error and return; the global error
// middleware maps it to an HTTP response.
func Bind(c *gin.Context, dst any) error {
	if err := c.ShouldBindJSON(dst); err != nil {
		return toValidationError(err)
	}
	return nil
}

func toValidationError(err error) error {
	var vErrs validator.ValidationErrors
	if !errors.As(err, &vErrs) {
		return errors.Join(domain.ErrValidation, err)
	}
	fields := make(map[string]string, len(vErrs))
	for _, fe := range vErrs {
		fields[strings.ToLower(fe.Field())] = describeTag(fe)
	}
	return domain.NewValidationError(fields)
}

func describeTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "required"
	case "min":
		return "minimum " + fe.Param()
	case "max":
		return "maximum " + fe.Param()
	case "len":
		return "length must be " + fe.Param()
	case "email":
		return "must be an email"
	case "e164":
		return "must be E.164 phone format"
	case "strongpassword":
		return "must be at least 6 characters"
	case "username":
		return "must be 3-30 chars: lowercase letters, digits, dot or underscore"
	default:
		return fe.Tag()
	}
}
