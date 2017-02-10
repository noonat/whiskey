package py

/*
#include "whiskey_py.h"
*/
import "C"

import (
	"github.com/pkg/errors"
)

// Iter wraps a Python iteration object.
type Iter struct {
	Object
}

// Next returns the next item for the iteration, or nil if there aren't
// any more items in the iteration.
func (it Iter) Next() (Object, error) {
	var o Object
	o.PyObject = C.PyIter_Next(it.PyObject)
	if o.PyObject == nil && C.PyErr_Occurred() != nil {
		return o, errors.Wrap(GetError(), "error calling iterator next")
	}
	return o, nil
}
