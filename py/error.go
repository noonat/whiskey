package py

/*
#include "whiskey_py.h"
*/
import "C"

import "github.com/pkg/errors"

var gettingError = false

// GetError returns the currently set exception as a Go error value.
//
// The Python C API doesn't provide a way to convert the current exception to
// a string, so this function imports the traceback module and uses the methods
// from there to convert it into a string.
//
// This clears the Python error as a side effect.
func GetError() error {
	if gettingError {
		// This is ugly but I haven't put time into figuring out a better way
		// yet... if we use the primitives from this package in this function
		// and one of them gets an error, they will try to call back into this
		// function again. Rather than allowing them to do this potentially
		// forever, only allow one level of nested calls here and just print
		// out any errors we encounter while in function.
		C.PyErr_Print()
		return errors.New("(recursive call to GetError, see printed error)")
	}
	gettingError = true
	defer func() {
		gettingError = false
	}()

	var typ, val, tb Object
	C.PyErr_Fetch(&typ.PyObject, &val.PyObject, &tb.PyObject)
	if typ.PyObject == nil {
		return nil
	}

	m, err := ImportModule("traceback")
	if err != nil {
		return err
	}
	defer m.DecRef()

	var s string
	if tb.PyObject == nil {
		s = "format_exception_only"
	} else {
		s = "format_exception"
	}
	fn, err := m.GetAttrString(s)
	if err != nil {
		return err
	}
	defer fn.DecRef()

	var o Object
	if tb.PyObject == nil {
		o, err = fn.Call(typ, val)
	} else {
		o, err = fn.Call(typ, val, tb)
	}
	if err != nil {
		return err
	}
	defer o.DecRef()
	list, err := o.List()
	if err != nil {
		return err
	}

	sep, err := NewString("\n")
	if err != nil {
		return err
	}
	defer sep.DecRef()

	ps, err := sep.Join(list)
	if err != nil {
		return err
	}
	defer ps.DecRef()

	s, err = ps.GoString()
	if err != nil {
		return err
	}

	return errors.New(s)
}
