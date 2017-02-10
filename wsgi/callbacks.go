package wsgi

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/noonat/whiskey/py"
)

// registerCallbacks creates callbacks within Whiskey's Python library to
// invoke functions in this package.
func registerCallbacks() {
	py.RegisterCallback("wsgi_errors_flush", wsgiErrorsFlush)
	py.RegisterCallback("wsgi_errors_write", wsgiErrorsWrite)
	py.RegisterCallback("wsgi_input_read", wsgiInputRead)
	py.RegisterCallback("wsgi_input_read_line", wsgiInputReadLine)
	py.RegisterCallback("wsgi_start_response", wsgiStartResponse)
}

func wsgiErrorsFlush(args py.Tuple) (py.Object, error) {
	// At the moment this is a no-op, because write is going to stderr, and
	// Go doesn't buffer stderr. I'm leaving this hook here, though, to
	// allow for the possibility of buffered error streams in the future.
	py.None.IncRef()
	return py.None, nil
}

// wsgiErrorsWrite writes a string to stderr.
func wsgiErrorsWrite(args py.Tuple) (py.Object, error) {
	var index int
	var s string
	if err := args.GetItems(&index, &s); err != nil {
		return py.Object{}, err
	}
	fmt.Fprint(os.Stderr, s)
	py.None.IncRef()
	return py.None, nil
}

// wsgiInputRead can be called from the application to read data from the
// request body, as a string. It can optionally pass an integer to limit the
// amount of data read. It defaults to reading the entire body. It should
// return an empty string to indicate EOF.
func wsgiInputRead(args py.Tuple) (py.Object, error) {
	var index int
	var sizeOrNone py.Object
	if err := args.GetItems(&index, &sizeOrNone); err != nil {
		return py.Object{}, err
	}
	defer sizeOrNone.DecRef()

	wr := requests[index]

	var b []byte
	if sizeOrNone != py.None {
		// Try to read exactly size bytes
		pn, err := sizeOrNone.Int()
		if err != nil {
			return py.Object{}, err
		}
		size, err := pn.GoInt()
		if err != nil {
			return py.Object{}, err
		}

		// FIXME: it might be nice to recycle these byte buffers
		b = make([]byte, size)
		n, err := io.ReadFull(wr.req.Body, b)
		switch err {
		case io.EOF, io.ErrUnexpectedEOF, nil:
			b = b[:n]
		default:
			return py.Object{}, err
		}
	} else {
		// Read until the end of the body
		var err error
		b, err = ioutil.ReadAll(wr.req.Body)
		if err != nil {
			return py.Object{}, err
		}
	}

	ps, err := py.NewString(string(b))
	if err != nil {
		return py.Object{}, err
	}
	return ps.Object, err
}

// wsgiInputReadLine reads a single line from the file.
func wsgiInputReadLine(args py.Tuple) (py.Object, error) {
	var index int
	if err := args.GetItems(&index); err != nil {
		return py.Object{}, err
	}

	wr := requests[index]

	line, err := wr.reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return py.Object{}, err
	}
	pl, err := py.NewString(line)
	if err != nil {
		return py.Object{}, err
	}
	return pl.Object, nil
}

// wsgiStartResponse is called by the WSGI application to specify the status
// code and headers for the response. It can be called more than once,
// although calls after the first are reserved for setting the excInfo
// parameter (to convert the response into an error response).
func wsgiStartResponse(args py.Tuple) (py.Object, error) {
	// FIXME: this doesn't handle excInfo yet

	var index int
	var status py.String
	var headers py.List
	var excInfo py.Object
	if err := args.GetItems(&index, &status, &headers, &excInfo); err != nil {
		return py.Object{}, err
	}
	defer status.DecRef()
	defer headers.DecRef()
	defer excInfo.DecRef()

	c, err := convertStatus(status)
	if err != nil {
		return py.Object{}, err
	}

	h, err := convertHeaders(headers)
	if err != nil {
		return py.Object{}, err
	}

	wr := requests[index]
	wr.code = c
	wr.headers = h

	py.None.IncRef()
	return py.None, nil
}
