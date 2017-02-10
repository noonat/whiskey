package wsgi

import (
	"bufio"
	"net/http"

	"github.com/noonat/whiskey/py"
)

// Request tracks the state associated with a single WSGI request. This is
// necessary because we need to track this data across Python boundaries,
// where we can't pass the Go pointer data into Python.
type Request struct {
	index int

	ts            *py.ThreadState
	application   py.Object
	startResponse py.Object
	wsgiInput     py.Object
	wsgiErrors    py.Object

	w      http.ResponseWriter
	req    *http.Request
	reader *bufio.Reader

	code    int
	headers http.Header
}

// NewRequest creates a new Request object for the given index. This also
// generates the associated Python functions and objects required for it to
// interact with WSGI applications.
func NewRequest(index int, application py.Object, ts *py.ThreadState) (*Request, error) {
	wr := &Request{
		application: application,
		index:       index,
		reader:      bufio.NewReader(nil),
		ts:          ts,
	}

	ts.Acquire()
	var err error
	wr.startResponse, wr.wsgiInput, wr.wsgiErrors, err = createRequestObjects(wr.index)
	ts.Release()
	if err != nil {
		return nil, err
	}

	return wr, nil
}

// Free releases the Python resoures associated with the Request object.
func (wr *Request) Free() {
	if wr.startResponse.PyObject != nil {
		wr.startResponse.DecRef()
		wr.startResponse.PyObject = nil
	}
	if wr.wsgiInput.PyObject != nil {
		wr.wsgiInput.DecRef()
		wr.wsgiInput.PyObject = nil
	}
	if wr.wsgiErrors.PyObject != nil {
		wr.wsgiErrors.DecRef()
		wr.wsgiErrors.PyObject = nil
	}
}

// Reset associates the existing Request object with a new HTTP request.
func (wr *Request) Reset(w http.ResponseWriter, req *http.Request) {
	wr.w = w
	wr.req = req
	wr.code = 0
	wr.headers = nil
	if req != nil {
		wr.reader.Reset(wr.req.Body)
	} else {
		wr.reader.Reset(nil)
	}
}
