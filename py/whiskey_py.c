#include "whiskey_py.h"
#include "_cgo_export.h"

static PyObject * _call(PyObject * const self, PyObject * const args)
{
  char * callName;
  PyObject * callArgs;
  if (!PyArg_ParseTuple(args, "sO:call", &callName, &callArgs)) {
    return NULL;
  }
  return whiskeyCall(callName, callArgs);
}

static PyMethodDef _module_defs[] = {
  {
    "call",
    _call,
    METH_VARARGS,
    "Call a method in Go."
  },
	{NULL, NULL, 0, NULL}
};

PyObject * whiskey_none = NULL;
PyObject * whiskey_true = NULL;
PyObject * whiskey_false = NULL;
PyObject * whiskey_module = NULL;

int whiskey_initialize() {
  Py_Initialize();
	Py_InitModule3("_whiskey", _module_defs, "Whiskey WSGI internals.");

  whiskey_none = Py_None;
  whiskey_true = Py_True;
  whiskey_false = Py_False;
  Py_INCREF(whiskey_none);
  Py_INCREF(whiskey_true);
  Py_INCREF(whiskey_false);

  return 0;
}

void whiskey_finalize() {
  if (whiskey_none != NULL) {
    Py_DECREF(whiskey_none);
    whiskey_none = NULL;
  }
  if (whiskey_true != NULL) {
    Py_DECREF(whiskey_true);
    whiskey_true = NULL;
  }
  if (whiskey_false != NULL) {
    Py_DECREF(whiskey_false);
    whiskey_false = NULL;
  }
  if (whiskey_module != NULL) {
    Py_DECREF(whiskey_module);
    whiskey_module = NULL;
  }
  Py_Finalize();
}

int whiskey_check_int(PyObject * o) {
  return PyInt_Check(o);
}

int whiskey_check_list(PyObject * o) {
  return PyList_Check(o);
}

int whiskey_check_string(PyObject * o) {
  return PyString_Check(o);
}

int whiskey_check_tuple(PyObject * o) {
  return PyTuple_Check(o);
}
