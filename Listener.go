package httpRedirectToHttps

import (
	"net"
	"crypto/tls"
	"time"
	"errors"
	"sync"
)

type Listener struct {
	req     NewListenerRequest
	closeCh chan struct{}
	errCh     chan error
	newConnCh chan net.Conn
	limitCh   chan struct{}
	closeOnce sync.Once
	closeErr error
}

type NewListenerRequest struct {
	Ln               net.Listener
	Cnf              *tls.Config
	ReadTimeout      time.Duration // default 10s
	LimitAcceptCount int           // default 1024
}

func NewListener(req NewListenerRequest) net.Listener {
	if req.Cnf == nil {
		panic(`c2u3qzgtd4 tls.config is nil`)
	}
	if req.LimitAcceptCount <= 0 {
		req.LimitAcceptCount = 1024
	}
	if req.ReadTimeout <= 0 {
		req.ReadTimeout = time.Second * 10
	}
	ret := &Listener{
		req:       req,
		errCh:     make(chan error, 1),
		newConnCh: make(chan net.Conn),
		limitCh:   make(chan struct{}, req.LimitAcceptCount),
		closeCh:   make(chan struct{}),
	}
	go ret.acceptThread()
	return ret
}

func (this *Listener) Accept() (net.Conn, error) {
	select {
	case conn := <-this.newConnCh:
		//log.Println("Accept 1")
		return conn, nil
	case err := <-this.errCh:
		//log.Println("Accept 2")
		return nil, err
	case <-this.closeCh:
		//log.Println("Accept 3")
		return nil, errors.New(`use of closed network connection`)
	}
}

func (this *Listener) acceptThread() {
	for {
		conn, err := this.req.Ln.Accept()
		if err != nil {
			select {
			case this.errCh <- err:
				continue
			case <-this.closeCh:
				return
			}
		}
		// limit connect number
		select {
		case this.limitCh <- struct{}{}:
		case <-this.closeCh:
			conn.Close()
			return
		}
		go func() {
			defer func(){
				<-this.limitCh
			}()
			tmp := []byte{0x00}
			finished := make(chan struct{})
			go func() { // improve server shutdown response speed
				select {
				case <-this.closeCh:
					conn.Close()
				case <-finished:
				}
			}()
			err := conn.SetReadDeadline(time.Now().Add(this.req.ReadTimeout))
			if err != nil {
				conn.Close()
				return
			}
			n, err := conn.Read(tmp)
			close(finished)
			if err != nil || n == 0 {
				conn.Close()
				return
			}
			var afterPeek net.Conn = &headBytesConn{
				head: tmp[0],
				Conn: conn,
			}
			// https
			if tmp[0] == 0x16 {
				afterPeek = tls.Server(afterPeek, this.req.Cnf)
			}
			select {
			case this.newConnCh <- afterPeek:
				// do not need close here
			case <-this.closeCh:
				conn.Close()
			}
		}()
	}
}

func (this *Listener) Close() error {
	this.closeOnce.Do(func() {
		this.closeErr = this.req.Ln.Close()
		close(this.closeCh)
	})
	return this.closeErr
}

func (this *Listener) Addr() net.Addr {
	return this.req.Ln.Addr()
}
