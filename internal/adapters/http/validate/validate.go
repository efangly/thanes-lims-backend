package validate

import (
	"fmt"

	"github.com/efangly/thanes-lims-backend/internal/domain/shared"
	"github.com/go-playground/validator/v10"
)

var v = validator.New()

// Struct validates a DTO's `validate` tags and wraps any failure as
// shared.ErrValidation so the central error mapper produces a consistent
// 400 envelope response.
func Struct(s any) error {
	if err := v.Struct(s); err != nil {
		return fmt.Errorf("%w: %s", shared.ErrValidation, err.Error())
	}
	return nil
}
