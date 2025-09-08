package stripe_sub

import "fmt"

type errKind string

const (
    kindConfig     errKind = "config"
    kindNotFound   errKind = "not_found"
    kindConflict   errKind = "conflict"
    kindValidation errKind = "validation"
    kindInternal   errKind = "internal"
)

type Error struct {
    Kind errKind
    Msg  string
}

func (e Error) Error() string { return fmt.Sprintf("%s: %s", e.Kind, e.Msg) }

func ErrConfig(msg string) error     { return Error{Kind: kindConfig, Msg: msg} }
func ErrNotFound(msg string) error   { return Error{Kind: kindNotFound, Msg: msg} }
func ErrConflict(msg string) error   { return Error{Kind: kindConflict, Msg: msg} }
func ErrValidation(msg string) error { return Error{Kind: kindValidation, Msg: msg} }
func ErrInternal(msg string) error   { return Error{Kind: kindInternal, Msg: msg} }

