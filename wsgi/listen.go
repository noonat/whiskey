package wsgi

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/noonat/whiskey/prefork"
	"github.com/noonat/whiskey/py"
	"github.com/pkg/errors"
)

// Worker serves requests using a Python WSGI application.
type Worker struct {
	Module   string
	NumConns int
}

var requests []*Request

// Serve accepts incoming HTTP connections on the listener l, creating a new
// service goroutines for each. The service goroutines invoke the Python WSGI
// application to handle the request.
func (wrk *Worker) Serve(ln net.Listener, logger prefork.Logger) error {
	if err := py.Initialize(); err != nil {
		return err
	}
	module := strings.Split(wrk.Module, ":")
	application, err := loadApplication(module[0], module[1])
	if err != nil {
		return err
	}
	defer application.DecRef()

	ts := py.GetThreadState()
	ts.Release()

	pool := make(chan *Request, wrk.NumConns)
	requests = make([]*Request, wrk.NumConns)
	for i := 0; i < wrk.NumConns; i++ {
		wr, err := NewRequest(i, application, ts.New())
		if err != nil {
			return err
		}
		pool <- wr
		requests[i] = wr
	}

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		wr := <-pool
		wr.Reset(w, req)
		wr.ts.Acquire()
		defer func() {
			wr.ts.Release()
			wr.Reset(nil, nil)
			pool <- wr
		}()

		response, err := callApplication(wr)
		if err == nil {
			err = writeResponse(wr, response)
		}
		if err != nil {
			logger.Printf("error serving request: %+v\n", err)
		}
	})

	srv := &http.Server{}
	ln = prefork.WorkerListener(ln, wrk.NumConns, 3*time.Minute)
	if err := srv.Serve(ln); err != nil {
		return errors.Wrap(err, "error serving in worker")
	}

	return nil
}
