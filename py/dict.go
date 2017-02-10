package py

/*
#include "whiskey_py.h"
*/
import "C"

import (
	"github.com/pkg/errors"
)

// Dict wraps a Python dict.
type Dict struct {
	Object
}

// NewDict creates a new Python dictionary.
func NewDict() (Dict, error) {
	var d Dict
	d.PyObject = C.PyDict_New()
	if d.PyObject == nil {
		return d, errors.Wrap(GetError(), "error creating Python dict")
	}
	return d, nil
}

// GetItem gets a value from the Python dict by key.
// This calls IncRef on the value before returning it (rather than returning
// a borrowed reference like PyDict_GetItem).
func (d Dict) GetItem(k Object) (Object, error) {
	var v Object
	v.PyObject = C.PyDict_GetItem(d.PyObject, k.PyObject)
	if v.PyObject == nil {
		return v, errors.Wrap(GetError(), "error getting dict item")
	}
	v.IncRef()
	return v, nil
}

// SetItem sets an item in a Python dict using Python Objects for the key
// and the value.
func (d Dict) SetItem(k, v Object) error {
	if C.PyDict_SetItem(d.PyObject, k.PyObject, v.PyObject) != 0 {
		return errors.Wrap(GetError(), "error setting dict item")
	}
	return nil
}

// SetItemInt sets an item in a Python dict using a Python Object for a key
// and a Go int for a value.
func (d Dict) SetItemInt(k Object, i int) error {
	v, err := NewInt(i)
	if err != nil {
		return err
	}
	return d.SetItem(k, v.Object)
}

// SetItemString sets an item in a Python dict using a Python Object for a key
// and a Go string for a value.
func (d Dict) SetItemString(k Object, s string) error {
	v, err := NewString(s)
	if err != nil {
		return err
	}
	return d.SetItem(k, v.Object)
}
