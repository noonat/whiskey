package wsgi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/noonat/whiskey/py"
	"github.com/pkg/errors"
)

var (
	pyCreateRequestObjects py.Object
	wsgiVersion            py.Tuple

	moduleSource = `
import _whiskey


__version__ = '0.1.0'


class ErrorsWriter(object):

    def __init__(self, index):
        self._index = index

    def flush(self):
        _whiskey.call("wsgi_errors_flush", (self._index,))

    def write(self, string):
        _whiskey.call("wsgi_errors_write", (self._index, string))

    def writelines(self, strings):
        for string in strings:
            _whiskey.call("wsgi_errors_write", (self._index, string))


class InputReader(object):

    def __init__(self, index):
        self._index = index

    def __iter__(self):
        return self

    def next(self):
        line = _whiskey.call("wsgi_input_read_line", (self._index, None))
        if line == '':
            raise StopIteration()
        return line

    def read(self, size=None):
        return _whiskey.call("wsgi_input_read", (self._index, size))

    def readline(self, size=None):
        return _whiskey.call("wsgi_input_read_line", (self._index, size))

    def readlines(self, sizehint=None):
        return list(self)


def create_request_objects(index):
    def start_response(status, headers, exc_info=None):
        return _whiskey.call("wsgi_start_response", (index, status, headers,
                                                     exc_info))
    return start_response, InputReader(index), ErrorsWriter(index)
`
)

func init() {
	py.AddInitializer(func() error {
		registerCallbacks()
		m, err := py.NewModuleString("whiskey", moduleSource)
		if err != nil {
			return err
		}
		defer m.DecRef()
		pyCreateRequestObjects, err = m.GetAttrString("create_request_objects")
		if err != nil {
			return err
		}
		return nil
	})
	py.AddFinalizer(func() error {
		if pyCreateRequestObjects.PyObject != nil {
			pyCreateRequestObjects.DecRef()
			pyCreateRequestObjects.PyObject = nil
		}
		return nil
	})
}

func loadApplication(moduleName, applicationName string) (py.Object, error) {
	m, err := py.ImportModule(moduleName)
	if err != nil {
		return py.Object{}, err
	}
	defer m.DecRef()

	application, err := m.GetAttrString(applicationName)
	if err != nil {
		return py.Object{}, err
	} else if !application.IsCallable() {
		application.DecRef()
		return py.Object{}, errors.New("application is not callable")
	}

	return application, nil
}

// callApplication runs the Python WSGI application function.
//
// The application function receives a start_response function as one of its
// arguments. It can later call this function to give the WSGI server a status
// code and headers. Because we may have multiple requests in progress at any
// given time, we need to associate the start_response function that we pass
// to the application with the WSGIRequest for this request. We do that by
// generating a start_response function per request and binding the WSGIRequest
// index as the "self" object associated with the function, so we can look it
// up again later.
func callApplication(wr *Request) (py.Iter, error) {
	environ, err := createEnviron(wr)
	if err != nil {
		return py.Iter{}, err
	}
	response, err := wr.application.Call(environ.Object, wr.startResponse)
	if err != nil {
		return py.Iter{}, err
	}
	return response.Iter()
}

// writeResponse iterates over the value returned by the WSGI application
// function, and writes each chunk out to the http.ResponseWriter.
func writeResponse(wr *Request, iter py.Iter) error {
	wroteHeaders := false
	for {
		value, err := iter.Next()
		if err != nil {
			return err
		} else if value.PyObject == nil {
			break
		}
		s, err := value.GoString()
		value.DecRef()
		if err != nil {
			return err
		}

		b := []byte(s)
		if len(b) == 0 {
			continue
		}
		if !wroteHeaders {
			for k, vs := range wr.headers {
				for _, v := range vs {
					wr.w.Header().Add(k, v)
				}
			}
			wr.w.WriteHeader(wr.code)
			wroteHeaders = true
		}
		wr.w.Write(b)
	}

	return nil
}

// convertHeaders converts the WSGI header list into an http.Header object.
//
// WSGI specifies that headers must be a list of tuples, where each tuple is a
// of header name and value. Iterate over them and convert them to the Go
// http.Header type.
func convertHeaders(wsgiHeaders py.List) (http.Header, error) {
	headers := http.Header{}

	length := wsgiHeaders.Len()
	for i := 0; i < length; i++ {
		h, err := wsgiHeaders.GetItem(i)
		if err != nil {
			return headers, err
		}
		ht, err := h.Tuple()
		if err != nil {
			return headers, err
		}
		var k, v string
		if err := ht.GetItems(&k, &v); err != nil {
			return headers, err
		}
		headers.Add(k, v)
	}

	return headers, nil
}

// convertStatus converts the WSGI status string into an integer code.
//
// WSGI specifies that status must be a string of the form "200 OK". We only
// care about the code itself, so convert that part to an integer that we can
// send to WriteHeader later.
func convertStatus(status py.String) (int, error) {
	s, err := status.GoString()
	if err != nil {
		return 500, err
	}
	i := strings.IndexRune(s, ' ')
	if i == -1 {
		return 500, errors.New("invalid status string, couldn't find space character")
	}
	code, err := strconv.Atoi(s[:i])
	if err != nil {
		return 500, errors.Wrap(err, "error converting status to integer")
	}
	return code, nil
}

func createRequestObjects(index int) (startResponse, wsgiInput, wsgiErrors py.Object, err error) {
	pi, err := py.NewInt(index)
	if err != nil {
		return
	}
	defer pi.DecRef()

	r, err := pyCreateRequestObjects.Call(pi.Object)
	if err != nil {
		return
	}
	defer r.DecRef()
	rt, err := r.Tuple()
	if err != nil {
		return
	}

	items := [3]py.Object{}
	for i := 0; i < 3; i++ {
		items[i], err = rt.GetItem(i)
		if err != nil {
			return
		}
	}
	startResponse, wsgiInput, wsgiErrors = items[0], items[1], items[2]
	return
}

func sicscs(d py.Dict, k, v string) error {
	pk, err := py.CachedString(k)
	if err != nil {
		return err
	}
	pv, err := py.CachedString(v)
	if err != nil {
		return err
	}
	return d.SetItem(pk.Object, pv.Object)
}

func sicss(d py.Dict, k, v string) error {
	pk, err := py.CachedString(k)
	if err != nil {
		return err
	}
	return d.SetItemString(pk.Object, v)
}

func sicsi(d py.Dict, k string, v py.Object) error {
	pk, err := py.CachedString(k)
	if err != nil {
		return err
	}
	return d.SetItem(pk.Object, v)
}

// createEnviron returns a WSGI environ dict for the given request.
func createEnviron(wr *Request) (py.Dict, error) {
	// The comments in this function come from the descriptions for keys in
	// PEP-3333. (https://www.python.org/dev/peps/pep-3333/)

	var d py.Dict

	if wsgiVersion.PyObject == nil {
		wv, err := py.NewTuple(2)
		if err != nil {
			return d, err
		} else if err := wv.SetItemInt(0, 1); err != nil {
			wv.DecRef()
			return d, err
		} else if err := wv.SetItemInt(1, 0); err != nil {
			wv.DecRef()
			return d, err
		}
		wsgiVersion = wv
	}

	d, err := py.NewDict()
	if err != nil {
		return d, err
	}

	// The HTTP request method, such as "GET" or "POST" . This cannot ever be
	// an empty string, and so is always required.
	sicscs(d, "REQUEST_METHOD", wr.req.Method)

	// The initial portion of the request URL's "path" that corresponds to the
	// application object, so that the application knows its virtual
	// "location". This may be an empty string, if the application corresponds
	// to the "root" of the server.
	sicscs(d, "SCRIPT_NAME", "")

	// The remainder of the request URL's "path", designating the virtual
	// "location" of the request's target within the application. This may be
	// an empty string, if the request URL targets the application root and
	// does not have a trailing slash.
	sicss(d, "PATH_INFO", wr.req.URL.Path)

	// The portion of the request URL that follows the "?", if any. May be
	// empty or absent.
	sicss(d, "QUERY_STRING", wr.req.URL.RawQuery)

	// The contents of any Content-Type fields in the HTTP request. May be
	// empty or absent.
	sicss(d, "CONTENT_TYPE", wr.req.Header.Get("Content-Type"))

	// The contents of any Content-Length fields in the HTTP request. May be
	// empty or absent.
	sicss(d, "CONTENT_LENGTH", wr.req.Header.Get("Content-Length"))

	// When combined with SCRIPT_NAME and PATH_INFO, these variables can be
	// used to complete the URL. Note, however, that HTTP_HOST, if present,
	// should be used in preference to SERVER_NAME for reconstructing the
	// request URL. SERVER_NAME and SERVER_PORT can never be empty strings, and
	// so are always required.
	sicss(d, "SERVER_NAME", "127.0.0.1")
	sicss(d, "SERVER_PORT", "8080")

	// The version of the protocol the client used to send the request.
	// Typically this will be something like "HTTP/1.0" or "HTTP/1.1" and may
	// be used by the application to determine how to treat any HTTP request
	// headers. (This variable should probably be called REQUEST_PROTOCOL,
	// since it denotes the protocol used in the request, and is not
	// necessarily the protocol that will be used in the server's response.
	// However, for compatibility with CGI we have to keep the existing name.)
	sicss(d, "SERVER_PROTOCOL", wr.req.Proto)

	// The tuple (1, 0), representing WSGI version 1.0.
	sicsi(d, "wsgi.version", wsgiVersion.Object)

	// A string representing the "scheme" portion of the URL at which the
	// application is being invoked. Normally, this will have the value "http"
	// or "https", as appropriate.
	var scheme string
	if wr.req.TLS != nil {
		scheme = "https"
	} else {
		scheme = "http"
	}
	sicscs(d, "wsgi.url_scheme", scheme)

	// An input stream (file-like object) from which the HTTP request body can
	// be read. (The server or gateway may perform reads on-demand as requested
	// by the application, or it may pre-read the client's request body and
	// buffer it in-memory or on disk, or use any other technique for providing
	// such an input stream, according to its preference.)
	sicsi(d, "wsgi.input", wr.wsgiInput)

	// An output stream (file-like object) to which error output can be written,
	// for the purpose of recording program or other errors in a standardized
	// and possibly centralized location. This should be a "text mode" stream;
	// i.e., applications should use "\n" as a line ending, and assume that it
	// will be converted to the correct line ending by the server/gateway.
	//
	// For many servers, wsgi.errors will be the server's main error log.
	// Alternatively, this may be sys.stderr, or a log file of some sort. The
	// server's documentation should include an explanation of how to configure
	// this or where to find the recorded output. A server or gateway may
	// supply different error streams to different applications, if this is
	// desired.
	sicsi(d, "wsgi.errors", wr.wsgiErrors)

	// This value should evaluate true if the application object may be
	// simultaneously invoked by another thread in the same process, and should
	// evaluate false otherwise.
	sicsi(d, "wsgi.multithread", py.True)

	// This value should evaluate true if an equivalent application object may
	// be simultaneously invoked by another process, and should evaluate false
	// otherwise.
	sicsi(d, "wsgi.multiprocess", py.True)

	// This value should evaluate true if the server or gateway expects (but
	// does not guarantee!) that the application will only be invoked this one
	// time during the life of its containing process. Normally, this will only
	// be true for a gateway based on CGI (or something similar).
	sicsi(d, "wsgi.run_once", py.False)

	return d, nil
}
