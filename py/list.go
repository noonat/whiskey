package py

/*
#include "whiskey_py.h"
*/
import "C"

import (
	"github.com/pkg/errors"
)

// List wraps a Python list object.
type List struct {
	Object
}

// NewList creates a new Python list with the given size.
func NewList(size int) (l List, err error) {
	l.PyObject = C.PyList_New(C.Py_ssize_t(size))
	if l.PyObject == nil {
		err = errors.Wrapf(GetError(), "error creating Python list of size %d", size)
	}
	return
}

// GetItem gets an item from the Python list.
// This calls IncRef on the value before returning it (rather than returning
// a borrowed reference like PyList_GetItem).
func (l List) GetItem(index int) (Object, error) {
	var o Object
	o.PyObject = C.PyList_GetItem(l.PyObject, C.Py_ssize_t(index))
	if o.PyObject == nil {
		return o, errors.Wrapf(GetError(), "error getting list item %d", index)
	}
	o.IncRef()
	return o, nil
}

// Len returns the size of the Python list.
func (l List) Len() int {
	return int(C.PyList_Size(l.PyObject))
}

// SetItem sets an item in the Python list using a Python object for the value.
//
// Note that, because C.PyList_SetItem steals references, this function calls
// IncRef on the passed object, to make things more consistent with other
// places in the Python C API.
func (l List) SetItem(index int, v Object) error {
	v.IncRef()
	if C.PyList_SetItem(l.PyObject, C.Py_ssize_t(index), v.PyObject) != 0 {
		v.DecRef()
		return errors.Wrapf(GetError(), "error setting tuple item %d", index)
	}
	return nil
}
