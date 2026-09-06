package bluetooth

import (
	"fmt"

	"knob/internal/domain/errs"
)

// Domain errors. They are compared with errors.Is and carry an errs.Kind
// class so the upper layers can react without knowing this package.
//
// Impossible transitions are KindConflict, not KindInvalid: the data is
// correct, what clashes is the device's state.
var (
	ErrInvalidAddress = errs.Invalid("the MAC address is not valid")
	ErrEmptyName      = errs.Invalid("name must not be empty")
	ErrNameTooLong    = errs.Invalid(fmt.Sprintf("name exceeds %d characters", MaxNameLength))
	ErrInvalidBattery = errs.Invalid("battery level must be between 0 and 100")
	ErrInvalidPasskey = errs.Invalid("the passkey must be at most 6 digits")
	ErrUnknownKind    = errs.Invalid("unknown device kind")

	ErrAdapterDisabled  = errs.Conflict("Bluetooth is disabled")
	ErrNotVisible       = errs.Conflict("the machine is not visible to other devices")
	ErrAlreadyKnown     = errs.Conflict("the device is already registered")
	ErrAlreadyPaired    = errs.Conflict("the device is already paired")
	ErrNotPaired        = errs.Conflict("the device is not paired")
	ErrAlreadyConnected = errs.Conflict("the device is already connected")
	ErrNotConnected     = errs.Conflict("the device is not connected")

	ErrNotFound = errs.NotFound("the device does not exist")
)
