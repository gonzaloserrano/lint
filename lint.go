// Package lint runs static analysis tools as go tests.
//
// It is intended to be used as a substitute for an external build
// step that runs tools such as go vet or golint.
//
package lint

import (
	"reflect"
	"strings"
)

type errorList []string

func (e errorList) Errors() []string { return []string(e) }
func (e errorList) Error() string    { return strings.Join(e, "\n") }

type errors interface {
	Errors() []string
}

// Checker is the interface that wraps the Check method.
//
// Check performs a static check of all files in pkgs, which may be fully
// qualified import paths, relative import paths or paths with the wildcard
// suffix ...
type Checker interface {
	Check(pkgs ...string) error
}

type group []Checker

// Group returns a Checker that applies each of checkers in the order provided.
//
// The error returned is either nil or contains errors returned by each Checker.
// These are exposed using the errors interface described in Skip and prefixed with the type of the
// Checker that generated the error. For example, the following error generated by govet.Checker:
//
//    file.go:23: err is unintentionally shadowed.
//
// is converted to:
//
//    govet.Checker: file.go:23: err is unintentionally shadowed.
//
// A checker is not shorted-circuited by a previous checker returning an error.
//
// Any error that implements errors is flattened into the final error list.
func Group(checkers ...Checker) Checker {
	return group(checkers)
}

func (g group) Check(pkgs ...string) error {
	var errs []string
	for _, checker := range g {
		name := reflect.TypeOf(checker).String()
		switch err := checker.Check(pkgs...).(type) {
		case nil:
			continue
		case errors:
			cerrs := err.Errors()
			for _, e := range cerrs {
				errs = append(errs, name+": "+e)
			}
		default:
			errs = append(errs, name+": "+err.Error())
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errorList(errs)
}