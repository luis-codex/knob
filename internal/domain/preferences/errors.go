package preferences

import (
	"fmt"

	"knob/internal/domain/errs"
)

// Domain errors. They are compared with errors.Is and carry an errs.Kind
// class so the upper layers can react without knowing this package.
var (
	ErrInvalidVolumeStep = errs.Invalid(fmt.Sprintf("volume_step must be between %d and %d", MinVolumeStep, MaxVolumeStep))
	ErrInvalidColor      = errs.Invalid("color must be a hex value, #RGB or #RRGGBB")
	ErrInvalidMode       = errs.Invalid("mode must be one of auto, light, dark")
)
