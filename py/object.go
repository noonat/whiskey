package py

/*
#include "whiskey_py.h"
*/
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/pkg/errors"
)

// Object embeds a *C.PyObject, which is the generic object wrapper type
// returned by most of the Python C API.
type Object struct {
	PyObject *C.PyObject
}

// NewModuleString Create a new module with th by compiling src into a string
func NewModuleString(name, src string) (Object, error) {
	var o Object

	cs := C.CString(src)
	cfn := C.CString(fmt.Sprintf("<string src for %s>", name))
	co := C.Py_CompileStringFlags(cs, cfn, C.Py_file_input, nil)
	C.free(unsafe.Pointer(cs))
	C.free(unsafe.Pointer(cfn))
	if co == nil {
		return o, errors.Errorf("error compiling source for module %s: %s", name, GetError())
	}
	defer C.Py_DecRef(co)

	cn := C.CString(name)
	o.PyObject = C.PyImport_ExecCodeModule(cn, co)
	C.free(unsafe.Pointer(cn))
	if o.PyObject == nil {
		return o, errors.Errorf("error executing module %s: %s", name, GetError())
	}
	return o, nil
}

// ImportModule imports a Python module.
func ImportModule(moduleName string) (Object, error) {
	var o Object
	cs := C.CString(moduleName)
	o.PyObject = C.PyImport_ImportModule(cs)
	C.free(unsafe.Pointer(cs))
	if o.PyObject == nil {
		return o, errors.Errorf("error importing module %s: %s", moduleName, GetError())
	}
	return o, nil
}

// Call invokes the associated Python object as a callable.
//
// This is equivalent to PyObject_CallObject. The passed arguments are packed
// into a tuple and passed to the callable as positional arguments.
func (o Object) Call(args ...Object) (Object, error) {
	var result Object
	var t Tuple
	if len(args) > 0 {
		var err error
		t, err = NewTupleObjects(args)
		if err != nil {
			return result, err
		}
		defer t.DecRef()
	}
	result.PyObject = C.PyObject_CallObject(o.PyObject, t.PyObject)
	if result.PyObject == nil {
		return result, errors.Errorf("error calling object: %s", GetError())
	}
	return result, nil
}

// ConvertInto converts the object into a more strict type, and stores the
// result in the given pointer.
//
// This uses a type assertion to figure out what ptr points to, and does any
// necessary validations and conversions. This doesn't support all types, just
// the ones necessary for Whiskey's codebase.
//
// When copying the Object into a pointer to another Object variable, it
// returns a new reference, not a borrowed one.
func (o Object) ConvertInto(ptr interface{}) error {
	switch t := ptr.(type) {
	case *List:
		pl, err := o.List()
		if err != nil {
			return err
		}
		pl.IncRef()
		*t = pl
	case *Object:
		o.IncRef()
		*t = o
	case *String:
		ps, err := o.String()
		if err != nil {
			return err
		}
		ps.IncRef()
		*t = ps
	case *int:
		n, err := o.GoInt()
		if err != nil {
			return err
		}
		*t = n
	case *string:
		s, err := o.GoString()
		if err != nil {
			return err
		}
		*t = s
	default:
		return errors.New("unsupported type for ptr")
	}

	return nil
}

// DecRef decrements the reference count for the Python object.
func (o Object) DecRef() {
	if o.PyObject != nil {
		C.Py_DecRef(o.PyObject)
	}
}

// IncRef increments the reference count for the Python object.
func (o Object) IncRef() {
	if o.PyObject != nil {
		C.Py_IncRef(o.PyObject)
	}
}

// GoInt converts the object into a Go int.
// The underlying type must be a Python int or an error will be returned.
func (o Object) GoInt() (int, error) {
	pn, err := o.Int()
	if err != nil {
		return 0, err
	}
	return pn.GoInt()
}

// GoString converts the object into a Go string.
// The underlying type must be a Python string or an error will be returned.
func (o Object) GoString() (string, error) {
	ps, err := o.String()
	if err != nil {
		return "", err
	}
	return ps.GoString()
}

// Int wraps the object in an Int struct.
// The underlying type must be a Python int or an error will be returned.
func (o Object) Int() (Int, error) {
	n := Int{o}
	if C.whiskey_check_int(o.PyObject) == 0 {
		return n, errors.New("object is not an integer")
	}
	return n, nil
}

// List wraps the object in a List struct.
// The underlying type must be a Python list or an error will be returned.
func (o Object) List() (List, error) {
	l := List{o}
	if C.whiskey_check_list(o.PyObject) == 0 {
		return l, errors.New("object is not a list")
	}
	return l, nil
}

// String wraps the object in a String struct.
// The underlying type must be a Python string or an error will be returned.
func (o Object) String() (String, error) {
	s := String{o}
	if C.whiskey_check_string(o.PyObject) == 0 {
		return s, errors.New("object is not a string")
	}
	return s, nil
}

// Tuple wraps the object in a Tuple struct.
// The underlying type must be a Python tuple or an error will be returned.
func (o Object) Tuple() (Tuple, error) {
	t := Tuple{o}
	if C.whiskey_check_tuple(o.PyObject) == 0 {
		return t, errors.New("object is not a tuple")
	}
	return t, nil
}

// GetAttrString returns the value for the given attribute.
// It's the equivalent of calling getattr(o, attr) in Python.
func (o Object) GetAttrString(attr string) (Object, error) {
	var v Object
	cs := C.CString(attr)
	v.PyObject = C.PyObject_GetAttrString(o.PyObject, cs)
	C.free(unsafe.Pointer(cs))
	if v.PyObject == nil {
		return v, errors.Errorf("error getting attribute %s: %s", attr, GetError())
	}
	return v, nil
}

// IsCallable returns true if the underlying Python object is callable.
func (o Object) IsCallable() bool {
	return C.PyCallable_Check(o.PyObject) != 0
}

// Iter returns an iterator for the object.
// It's the equivalent of calling iter(o) in Python.
func (o Object) Iter() (Iter, error) {
	var it Iter
	it.PyObject = C.PyObject_GetIter(o.PyObject)
	if it.PyObject == nil {
		return it, errors.New("error getting iterator")
	}
	return it, nil
}
