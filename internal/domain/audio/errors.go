package audio

import (
	"fmt"

	"settings-cli/internal/domain/errs"
)

// Domain errors. They are compared with errors.Is and carry an errs.Kind
// class so the upper layers can react without knowing this package.
var (
	ErrInvalidID       = errs.Invalid("invalid device identifier")
	ErrInvalidStreamID = errs.Invalid("invalid stream identifier")
	ErrEmptyName       = errs.Invalid("name must not be empty")
	ErrNameTooLong     = errs.Invalid(fmt.Sprintf("name exceeds %d characters", MaxNameLength))
	ErrInvalidVolume   = errs.Invalid(fmt.Sprintf("volume must be between 0 and %d", MaxVolume))

	ErrWrongDirection = errs.Conflict("the device is not of that kind")

	ErrNotFound       = errs.NotFound("the audio device does not exist")
	ErrStreamNotFound = errs.NotFound("the audio stream no longer exists")
)
