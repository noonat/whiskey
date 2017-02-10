package py

// This file contains helper functions for the Python C API. Many of the Python C
// API functions are a bit cumbersome to use, especially from cgo, so these
// helpers make life a little easier.

/*
#cgo pkg-config: python2
#include "whiskey_py.h"
*/
import "C"

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/pkg/errors"
)

type manyErrors struct {
	err  error
	errs []error
}

func (e *manyErrors) Error() string {
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "%+v\n", e.err)
	fmt.Fprintf(buf, "%d errors occurred, listed below:\n", len(e.errs))
	for i, err := range e.errs {
		fmt.Fprintf(buf, "\n- error %d: %+v\n", i+1, err)
	}
	return buf.String()
}

var (
	// None is a wrapper for the Python None value.
	None Object

	// True is a wrapper for the Python True value.
	True Object

	// False is a wrapper for the Python False value.
	False Object

	initialized  = false
	initializers = []func() error{}
	finalizers   = []func() error{}
	mutex        = &sync.Mutex{}
)

func AddInitializer(fn func() error) error {
	initializers = append(initializers, fn)
	return nil
}

func AddFinalizer(fn func() error) error {
	finalizers = append(finalizers, fn)
	return nil
}

func initialize() error {
	if C.whiskey_initialize() != 0 {
		return errors.New("whiskey_initialize failed")
	}
	C.PyEval_InitThreads()

	None.PyObject = C.whiskey_none
	True.PyObject = C.whiskey_true
	False.PyObject = C.whiskey_false
	resetStringCache()

	var errs []error
	for _, fn := range initializers {
		if err := fn(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return &manyErrors{
			err:  errors.New("error calling initialization functions"),
			errs: errs,
		}
	}

	return nil
}

func finalize() error {
	errs := []error{}
	for _, fn := range finalizers {
		if err := fn(); err != nil {
			errs = append(errs, err)
		}
	}

	None.PyObject = nil
	True.PyObject = nil
	False.PyObject = nil
	resetStringCache()

	C.whiskey_finalize()

	if len(errs) > 0 {
		return &manyErrors{
			err:  errors.New("error calling finalization functions"),
			errs: errs,
		}
	}

	return nil
}

func Initialize() error {
	if err := initialize(); err != nil {
		return err
	}
	return nil
}
