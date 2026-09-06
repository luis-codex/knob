// Package errs classifies domain failures. It is named errs so it does not
// clash with the stdlib errors.
//
// Concrete errors are declared by each aggregate; only their class lives here.
package errs

import "errors"

// Kind is the failure class, what the upper layers need to decide how to
// react.
type Kind int

const (
	// KindInvalid: the data breaks an invariant.
	KindInvalid Kind = iota
	// KindNotFound: the entity does not exist.
	KindNotFound
	// KindConflict: the operation clashes with the current state.
	KindConflict
)

// Error is a domain failure with its class.
type Error struct {
	Kind  Kind
	Msg   string
	cause error
}

func (e *Error) Error() string { return e.Msg }

func (e *Error) Unwrap() error { return e.cause }

// Constructors. They return error, not *Error, to force comparison by class
// rather than by concrete type.
func Invalid(msg string) error  { return &Error{Kind: KindInvalid, Msg: msg} }
func NotFound(msg string) error { return &Error{Kind: KindNotFound, Msg: msg} }
func Conflict(msg string) error { return &Error{Kind: KindConflict, Msg: msg} }

// Wrap gives a domain class to an external error without losing the original.
func Wrap(kind Kind, msg string, cause error) error {
	return &Error{Kind: kind, Msg: msg, cause: cause}
}

// KindOf returns the error's class. The second value tells an unclassified
// failure apart from an invalid one.
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
