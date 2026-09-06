package bluetooth

import (
	"fmt"

	"settings-cli/internal/domain/errs"
)

// Errores del dominio. Se comparan con errors.Is y llevan clase errs.Kind
// para que las capas superiores reaccionen sin conocer este paquete.
//
// Las transiciones imposibles son KindConflict y no KindInvalid: los datos son
// correctos, lo que choca es el estado del dispositivo.
var (
	ErrInvalidAddress = errs.Invalid("la dirección MAC no es válida")
	ErrEmptyName      = errs.Invalid("el nombre no puede estar vacío")
	ErrNameTooLong    = errs.Invalid(fmt.Sprintf("el nombre supera los %d caracteres", MaxNameLength))
	ErrInvalidBattery = errs.Invalid("el nivel de batería debe estar entre 0 y 100")
	ErrInvalidPasskey = errs.Invalid("el código debe tener como máximo 6 dígitos")
	ErrUnknownKind    = errs.Invalid("tipo de dispositivo desconocido")

	ErrAdapterDisabled  = errs.Conflict("el Bluetooth está desactivado")
	ErrNotVisible       = errs.Conflict("el equipo no es visible para otros dispositivos")
	ErrAlreadyKnown     = errs.Conflict("el dispositivo ya está registrado")
	ErrAlreadyPaired    = errs.Conflict("el dispositivo ya está emparejado")
	ErrNotPaired        = errs.Conflict("el dispositivo no está emparejado")
	ErrAlreadyConnected = errs.Conflict("el dispositivo ya está conectado")
	ErrNotConnected     = errs.Conflict("el dispositivo no está conectado")

	ErrNotFound = errs.NotFound("el dispositivo no existe")
)
