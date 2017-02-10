package py

/*
#include "whiskey_py.h"
*/
import "C"

import (
	"sync"
	"unsafe"

	"github.com/pkg/errors"
)

// String wraps a Python string.
type String struct {
	Object
}

// String converts a Go string to a Python string.
func NewString(s string) (String, error) {
	var ps String
	cs := C.CString(s)
	ps.PyObject = C.PyString_FromString(cs)
	C.free(unsafe.Pointer(cs))
	if ps.PyObject == nil {
		return ps, errors.Wrap(GetError(), "error converting to Python string")
	}
	return ps, nil
}

// GoString converts the Python string into a Go string.
func (s String) GoString() (string, error) {
	// NOTE: Don't call C.free(cs) here, because the Python C API says that
	// the returned char* shouldn't be modified or freed.
	cs := C.PyString_AsString(s.PyObject)
	if cs == nil {
		return "", errors.Wrap(GetError(), "error converting to C string")
	}
	return C.GoString(cs), nil
}

// Join joins all the items in the list by this string.
// It's equivalent to s.join(l) in Python.
func (s String) Join(l List) (String, error) {
	var joined String
	joined.PyObject = C.PyUnicode_Join(s.PyObject, l.PyObject)
	if joined.PyObject == nil {
		return joined, errors.Wrap(GetError(), "error joining list")
	}
	return joined, nil
}

var (
	stringCache      = map[string]String{}
	stringCacheMutex = &sync.Mutex{}
)

func resetStringCache() {
	stringCacheMutex.Lock()
	for _, v := range stringCache {
		v.DecRef()
	}
	stringCache = map[string]String{}
	stringCacheMutex.Unlock()
}

// CachedString converts a Go string to a Python string and caches it.
func CachedString(s string) (String, error) {
	stringCacheMutex.Lock()
	ps, ok := stringCache[s]
	if !ok {
		var err error
		ps, err = NewString(s)
		if err != nil {
			return ps, err
		}
		stringCache[s] = ps
	}
	stringCacheMutex.Unlock()
	ps.IncRef()
	return ps, nil
}
