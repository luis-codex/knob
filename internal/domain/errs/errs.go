// Package errs clasifica los fallos del dominio. Se llama errs para no chocar
// con el errors de la stdlib.
//
// Los errores concretos los declara cada agregado; aquí vive solo su clase.
package errs

import "errors"

// Kind es la clase de fallo, lo que las capas superiores necesitan para
// decidir cómo reaccionar.
type Kind int

const (
	// KindInvalid: los datos no cumplen una invariante.
	KindInvalid Kind = iota
	// KindNotFound: la entidad no existe.
	KindNotFound
	// KindConflict: la operación choca con el estado actual.
	KindConflict
)

// Error es un fallo de dominio con su clase.
type Error struct {
	Kind  Kind
	Msg   string
	cause error
}

func (e *Error) Error() string { return e.Msg }

func (e *Error) Unwrap() error { return e.cause }

// Constructores. Devuelven error, no *Error, para forzar la comparación por
// clase en vez de por tipo concreto.
func Invalid(msg string) error  { return &Error{Kind: KindInvalid, Msg: msg} }
func NotFound(msg string) error { return &Error{Kind: KindNotFound, Msg: msg} }
func Conflict(msg string) error { return &Error{Kind: KindConflict, Msg: msg} }

// Wrap da clase de dominio a un error externo sin perder el original.
func Wrap(kind Kind, msg string, cause error) error {
	return &Error{Kind: kind, Msg: msg, cause: cause}
}

// KindOf devuelve la clase del error. El segundo valor distingue un fallo sin
// clasificar de uno inválido.
func KindOf(err error) (Kind, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e.Kind, true
	}
	return KindInvalid, false
}

func Is(err error, kind Kind) bool {
	k, ok := KindOf(err)
	return ok && k == kind
}

func IsInvalid(err error) bool  { return Is(err, KindInvalid) }
func IsNotFound(err error) bool { return Is(err, KindNotFound) }
func IsConflict(err error) bool { return Is(err, KindConflict) }
