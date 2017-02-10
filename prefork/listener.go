package prefork

import (
	"net"
	"sync"
	"time"
)

// WorkerListener is a wrapper around a net.Listener that adds helpful things
// that prefork workers will often need. It limits the number of simultaneous
// connections to the specified value, and sets keep alive on the accepted
// conn to the given duration.
//
// Usage of this in workers is completely optional.
func WorkerListener(l net.Listener, numConns int, keepAlivePeriod time.Duration) net.Listener {
	return &workerListener{
		TCPListener:     l.(*net.TCPListener),
		throttle:        make(chan struct{}, numConns),
		keepAlivePeriod: keepAlivePeriod,
	}
}

type workerListener struct {
	*net.TCPListener
	throttle        chan struct{}
	keepAlivePeriod time.Duration
}

func (wl workerListener) acquire() {
	wl.throttle <- struct{}{}
}

func (wl workerListener) release() {
	<-wl.throttle
}

func (wl workerListener) Accept() (net.Conn, error) {
	wl.acquire()
	tc, err := wl.TCPListener.AcceptTCP()
	if err != nil {
		wl.release()
		return nil, err
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(wl.keepAlivePeriod)
	return &workerConn{Conn: tc, release: wl.release}, nil
}

type workerConn struct {
	net.Conn
	release     func()
	releaseOnce sync.Once
}

func (wc *workerConn) Close() error {
	err := wc.Conn.Close()
	wc.releaseOnce.Do(wc.release)
	return err
}
