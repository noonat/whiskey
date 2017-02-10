package prefork

import (
	"net"
	"sync"
	"testing"
	"time"
)

func TestWorkerListener(t *testing.T) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	wl := WorkerListener(l, 2, 30*time.Second)
	conns := []net.Conn{}
	go func() {
		for i := 0; i < 3; i++ {
			c, err := wl.Accept()
			if err != nil {
				t.Error(err)
				break
			}
			conns = append(conns, c)
		}
		wg.Done()
	}()
	defer func() {
		for _, c := range conns {
			c.Close()
		}
	}()

	connected := make(chan struct{})
	go func() {
		for i := 0; i < 3; i++ {
			_, err := net.Dial("tcp", wl.Addr().String())
			if err != nil {
				t.Error(err)
			}
		}
		close(connected)
	}()

	timer := time.NewTimer(time.Second)
	select {
	case <-timer.C:
		t.Fatal("timed out")
	case <-connected:
		timer.Stop()
	}

	if len(conns) != 2 {
		t.Fatalf("expected 2 conns, got %d", len(conns))
	}
	conns[0].Close()
	wg.Wait()
	if len(conns) != 3 {
		t.Fatalf("expected 3 conns, got %d", len(conns))
	}
}
