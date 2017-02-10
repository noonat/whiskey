package py

/*
#include "whiskey_py.h"
*/
import "C"

import "log"

// CallbackFunc is used with RegisterCallback.
type CallbackFunc func(args Tuple) (Object, error)

var callbacks = map[string]CallbackFunc{}

// RegisterCallback adds a callback function to the internal map for the
// Whiskey Python module. This provides a way to easily invoke Go functions
// from Python without needing to create exported cgo functions and Python
// wrappers for each one.
func RegisterCallback(name string, fn CallbackFunc) {
	callbacks[name] = fn
}

//export whiskeyCall
func whiskeyCall(name *C.char, args *C.PyObject) *C.PyObject {
	k := C.GoString(name)
	fn, ok := callbacks[k]
	if !ok {
		log.Printf("error: whiskey_call called for unknown callback %q\n", k)
		None.IncRef()
		return None.PyObject
	}
	result, err := fn(Tuple{Object{args}})
	if err != nil {
		log.Printf("error: whiskey_call for %q failed: %+v\n", k, err)
		None.IncRef()
		return None.PyObject
	}
	return result.PyObject
}
