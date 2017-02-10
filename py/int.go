package py

/*
#include "whiskey_py.h"
*/
import "C"

import (
	"github.com/pkg/errors"
)

// Int wraps a Python integer.
type Int struct {
	Object
}

// NewInt converts a Go int into a Python Int.
func NewInt(n int) (Int, error) {
	var pn Int
	pn.PyObject = C.PyInt_FromLong(C.long(n))
	if pn.PyObject == nil {
		return pn, errors.Wrap(GetError(), "error converting to Python int")
	}
	return pn, nil
}

// GoInt converts the Python int into a Go int.
func (pn Int) GoInt() (int, error) {
	n := C.PyInt_AsLong(pn.PyObject)
	if n == -1 && C.PyErr_Occurred() != nil {
		return 0, errors.Wrap(GetError(), "error converting to Go int")
	}
	return int(n), nil
}
