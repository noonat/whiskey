// Package prefork provides a way to run processes in a prefork server model,
// where a manager process launches a number of subprocesses, all intended to
// communicate in parallel over a shared listener.
//
// This model isn't usually necessary in Go, but it can be helpeful when
// wrapping other software that aren't compatible with Go's concurrency model.
// Python is a good example of this, as it requires a mutex lock across
// threads, and shares global state at the process level.
package prefork

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"sync"

	"github.com/pkg/errors"
)

func execWorker(wg *sync.WaitGroup, lnf *os.File, env []string, logger Logger) (*Pipe, error) {
	mp, wp, err := NewPipes()
	if err != nil {
		return nil, errors.Wrap(err, "error creating worker pipes")
	}

	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Env = env
	cmd.ExtraFiles = []*os.File{lnf, wp.ReadFile, wp.WriteFile}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Start(); err != nil {
		mp.Close()
		wp.Close()
		return nil, errors.Wrap(err, "error starting worker")
	}

	// Workers processes should gracefully stop when they receieve an
	// interrupt. Ignore the first interrupt we receive, to give them a chance
	// to stop. But if we receive a second one, allow this process to stop.
	closing := false
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		<-interrupt
		signal.Stop(interrupt)
		closing = true
	}()

	// This keepalive communication between processes isn't very advanced
	// right now. It just writes a byte 1 once a second from the worker to
	// the manager to indicate that it's still alive. When it is shutting
	// down, it sends a 0.
	//
	// This pipe will eventually be used to communicate things like stats
	// up to the manager, to coordinate shutdown, etc.
	go func() {
		b := []byte{1}
		for {
			_, err := mp.Read(b)
			if err != nil {
				if !closing {
					logger.Println("error reading keepalive:", err)
				}
				break
			}
			if b[0] == 0 {
				break
			}
		}
	}()

	go func(cmd *exec.Cmd) {
		pid := cmd.Process.Pid
		if err := cmd.Wait(); err != nil {
			logger.Printf("worker %d wait returned error: %s", pid, err)
		} else {
			logger.Printf("worker %d stopped", pid)
		}
		wg.Done()
		wp.Close()
		mp.Close()
	}(cmd)

	return mp, nil
}

func runManager(w Worker, addr string, numWorkers int, logger Logger) error {
	logger.SetPrefix(fmt.Sprintf("manager\t[%d]\t", os.Getpid()))

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.Wrap(err, "error creating listener")
	}
	logger.Println("listening on", addr)

	if numWorkers == 0 {
		// FIXME: it might be better to reuse the code in worker.go, so that
		// listener would get closed on sigint.
		logger.Println("worker count is 0, running in single process mode")
		return w.Serve(ln, logger)
	}

	lnf, err := ln.(*net.TCPListener).File()
	if err != nil {
		return errors.Wrap(err, "error getting file for listener")
	}

	// FIXME: The worker management is pretty naive right now. It just starts
	// them all up once and assumes they'll keep running. Eventually, the
	// manager should kill workers that are running but non-responsive, and
	// should ensure that the appropriate number of workers are always running.
	//
	// It would also be nice if it also supported some of the signals that
	// gunicorn does for worker management.
	logger.Println("starting", numWorkers, "workers")
	wg := &sync.WaitGroup{}
	wg.Add(numWorkers)
	env := append([]string{}, os.Environ()...)
	env = append(env, "PREFORK_WORKER=1")
	for i := 0; i < numWorkers; i++ {
		if _, err := execWorker(wg, lnf, env, logger); err != nil {
			return err
		}
	}
	wg.Wait()

	return nil
}

// Run starts the process. It handles either starting the manager or starting
// the worker, as appropriate, depending on the execution environment.
//
// The manager will create a listener on the given address, and launch
// subprocesses for the number of workers specified. It runs the workers with
// the same arguments the original process was passed, and also adds a
// PREFORK_WORKER environment variable. It uses the presence of that variable
// to determine that the subprocess should act as a worker.
//
// When this function is invoked in a worker, the worker calls w.Serve(...)
// and passes it a listener.
func Run(w Worker, addr string, numWorkers int, logger Logger) error {
	if os.Getenv("PREFORK_WORKER") != "" {
		return runWorker(w, logger)
	}
	return runManager(w, addr, numWorkers, logger)
}
