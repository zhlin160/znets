package znets

import (
	"net"
	"sync"
	"time"
)

func NewListener(listener *net.TCPListener, server *Server) *Listener {
	return &Listener{listener, &sync.WaitGroup{}, server}
}

type Listener struct {
	*net.TCPListener
	wg     *sync.WaitGroup
	server *Server
}

func (l *Listener) GetWg() *sync.WaitGroup {
	return l.wg
}

func (l *Listener) Accept() (*net.TCPConn, error) {
	tc, err := l.AcceptTCP()
	if err != nil {
		return nil, err
	}

	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)

	l.wg.Add(1)
	return tc, nil
}

func (l *Listener) Wait() {
	l.wg.Wait()
}

func (l *Listener) GetFd() (uintptr, error) {
	file, err := l.TCPListener.File()
	if err != nil {
		return 0, err
	}
	return file.Fd(), nil
}
