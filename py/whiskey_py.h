#ifndef __WSGI_H__
#define __WSGI_H__

#include <Python.h>

extern PyMethodDef whiskey_start_response_def;
extern PyObject * whiskey_none;
extern PyObject * whiskey_true;
extern PyObject * whiskey_false;
extern PyObject * whiskey_module;

int whiskey_initialize();
void whiskey_finalize();
int whiskey_check_int(PyObject * o);
int whiskey_check_list(PyObject * o);
int whiskey_check_string(PyObject * o);
int whiskey_check_tuple(PyObject * o);

#endif
