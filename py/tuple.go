package py

/*
#include "whiskey_py.h"
*/
import "C"

import (
	"fmt"

	"github.com/pkg/errors"
)

// Tuple wraps a Python tuple object.
type Tuple struct {
	Object
}

// NewTuple creates a new Python tuple with the given size.
func NewTuple(size int) (t Tuple, err error) {
	t.PyObject = C.PyTuple_New(C.Py_ssize_t(size))
	if t.PyObject == nil {
		err = errors.Wrapf(GetError(), "error creating Python tuple of size %d", size)
	}
	return
}

// NewTupleObjects create a new Python tuple from a slice of Python objects.
func NewTupleObjects(objects []Object) (t Tuple, err error) {
	t, err = NewTuple(len(objects))
	if err != nil {
		return
	}
	for i, arg := range objects {
		if err = t.SetItem(i, arg); err != nil {
			t.DecRef()
			return
		}
	}
	return
}

// GetItem gets an item from the Python tuple.
// This calls IncRef on the value before returning it (rather than returning
// a borrowed reference like PyTuple_GetItem).
func (t Tuple) GetItem(index int) (item Object, err error) {
	item.PyObject = C.PyTuple_GetItem(t.PyObject, C.Py_ssize_t(index))
	if item.PyObject == nil {
		err = errors.Wrapf(GetError(), "error getting tuple item %d", index)
	}
	item.IncRef()
	return
}

func (t Tuple) GetItems(ptrs ...interface{}) error {
	for i, ptr := range ptrs {
		o, err := t.GetItem(i)
		if err != nil {
			return err
		}
		defer o.DecRef()
		if err := o.ConvertInto(ptr); err != nil {
			return errors.WithMessage(err, fmt.Sprintf("(for tuple item %d)", i))
		}
	}
	return nil
}

// SetItem sets an item in the Python tuple using a Python object for the value.
//
// Note that, because C.PyTuple_SetItem steals references, this function calls
// IncRef on the passed object, to make things more consistent with other
// places in the Python C API.
func (t Tuple) SetItem(index int, v Object) error {
	v.IncRef()
	if C.PyTuple_SetItem(t.PyObject, C.Py_ssize_t(index), v.PyObject) != 0 {
		v.DecRef()
		return errors.Wrapf(GetError(), "error setting tuple item %d", index)
	}
	return nil
}

func (t Tuple) SetItemInt(index, n int) error {
	pn, err := NewInt(n)
	if err != nil {
		return err
	}
	if err := t.SetItem(index, pn.Object); err != nil {
		return err
	}
	pn.DecRef()
	return nil
}

// Len returns the size of the Python tuple.
func (t Tuple) Len() int {
	return int(C.PyTuple_Size(t.PyObject))
}
