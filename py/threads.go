package py

/*
#include "whiskey_py.h"
*/
import "C"

type ThreadState struct {
	PyThreadState *C.PyThreadState
}

func GetThreadState() *ThreadState {
	return &ThreadState{PyThreadState: C.PyThreadState_Get()}
}

func (ts *ThreadState) New() *ThreadState {
	return &ThreadState{C.PyThreadState_New(ts.PyThreadState.interp)}
}

func (ts *ThreadState) Acquire() {
	C.PyEval_RestoreThread(ts.PyThreadState)
}

func (ts *ThreadState) Release() {
	ts.PyThreadState = C.PyEval_SaveThread()
}
