package audio

import (
	"fmt"

	"settings-cli/internal/domain/errs"
)

// Errores del dominio. Se comparan con errors.Is y llevan clase errs.Kind
// para que las capas superiores reaccionen sin conocer este paquete.
var (
	ErrInvalidID       = errs.Invalid("identificador de dispositivo inválido")
	ErrInvalidStreamID = errs.Invalid("identificador de flujo inválido")
	ErrEmptyName       = errs.Invalid("el nombre no puede estar vacío")
	ErrNameTooLong     = errs.Invalid(fmt.Sprintf("el nombre supera los %d caracteres", MaxNameLength))
	ErrInvalidVolume   = errs.Invalid(fmt.Sprintf("el volumen debe estar entre 0 y %d", MaxVolume))

	ErrWrongDirection = errs.Conflict("el dispositivo no es de ese tipo")

	ErrNotFound       = errs.NotFound("el dispositivo de audio no existe")
	ErrStreamNotFound = errs.NotFound("el flujo de audio ya no existe")
)
