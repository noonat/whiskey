package prefork

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/pkg/errors"
)

// These file descriptors are passed from the master to the worker.
const (
	listenFD          = uintptr(3)
	workerPipeReadFD  = uintptr(4)
	workerPipeWriteFD = uintptr(5)
)

// Worker is the interface that workers must implement. The master creates the
// listening TCP socket, and it's passed into each of the workers so they can
// all call Accept() on it.
//
// Serve should block indefinitely and return when the worker should terminate.
type Worker interface {
	Serve(ln net.Listener, logger Logger) error
}

func runWorker(w Worker, logger Logger) error {
	logger.SetPrefix(fmt.Sprintf("worker\t[%d]\t", os.Getpid()))

	// Recreate the TCP listener from the inherited files
	lnf := os.NewFile(listenFD, "")
	ln, err := net.FileListener(lnf)
	if err != nil {
		return errors.Wrap(err, "error creating net.FileListener")
	}

	// Recreate the internal communication pipe from the inherited files
	wp := &Pipe{
		ReadFile:  os.NewFile(workerPipeReadFD, ""),
		WriteFile: os.NewFile(workerPipeWriteFD, ""),
	}
	defer wp.Close()

	// Shutdown on SIGINT or when the done channel is closed
	closed := false
	done := make(chan struct{})
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		select {
		case <-done:
		case <-interrupt:
			close(done)
		}
		signal.Stop(interrupt)
		closed = true
		wp.Write([]byte{0})
		ln.Close()
	}()

	// Send a keepalive to the master
	t := time.NewTicker(time.Second)
	defer t.Stop()
	go func() {
		b := []byte{1}
		var err error
		for range t.C {
			_, err = wp.Write(b)
			if err != nil {
				break
			}
		}
		if !closed && err != nil {
			logger.Println("error writing keepalive:", err)
		}
	}()

	logger.Println("started worker")
	if err := w.Serve(ln, logger); !closed && err != nil {
		return err
	}

	return nil
}
