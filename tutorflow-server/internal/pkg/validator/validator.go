package validator

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/tutorflow/tutorflow-server/internal/domain"
)

var (
	validate  *validator.Validate
	slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

func init() {
	validate = validator.New()

	// Use JSON tag names in error messages
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// Register custom validators
	validate.RegisterValidation("slug", validateSlug)
	validate.RegisterValidation("password", validatePassword)
}

// Validate validates a struct
func Validate(i interface{}) error {
	return validate.Struct(i)
}

// ValidateVar validates a single variable
func ValidateVar(field interface{}, tag string) error {
	return validate.Var(field, tag)
}

// FormatValidationErrors converts validator errors to domain format
func FormatValidationErrors(err error) domain.ValidationErrors {
	var errors domain.ValidationErrors

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrs {
			errors = append(errors, domain.ValidationError{
				Field:   e.Field(),
				Message: getErrorMessage(e),
			})
		}
	}

	return errors
}

func getErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		if e.Type().Kind() == reflect.String {
			return "Must be at least " + e.Param() + " characters"
		}
		return "Must be at least " + e.Param()
	case "max":
		if e.Type().Kind() == reflect.String {
			return "Must be at most " + e.Param() + " characters"
		}
		return "Must be at most " + e.Param()
	case "len":
		return "Must be exactly " + e.Param() + " characters"
	case "uuid":
		return "Must be a valid UUID"
	case "url":
		return "Must be a valid URL"
	case "oneof":
		return "Must be one of: " + e.Param()
	case "slug":
		return "Must be a valid slug (lowercase letters, numbers, and hyphens)"
	case "password":
		return "Password must be at least 8 characters with uppercase, lowercase, and number"
	case "eqfield":
		return "Must match " + e.Param()
	case "gtfield":
		return "Must be greater than " + e.Param()
	case "ltfield":
		return "Must be less than " + e.Param()
	case "gte":
		return "Must be greater than or equal to " + e.Param()
	case "lte":
		return "Must be less than or equal to " + e.Param()
	default:
		return "Invalid value"
	}
}

func validateSlug(fl validator.FieldLevel) bool {
	return slugRegex.MatchString(fl.Field().String())
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 8 {
		return false
	}

	var hasUpper, hasLower, hasNumber bool
	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasNumber = true
		}
	}

	return hasUpper && hasLower && hasNumber
}
