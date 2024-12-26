package httpRedirectToHttps

import (
	"net"
)

type headBytesConn struct {
	net.Conn
	head byte
	hasHandle bool
}

func (this *headBytesConn) Read(b []byte) (n int, err error) {
	if this.hasHandle==false {
		b[0] = this.head
		this.hasHandle = true
		n, err = this.Conn.Read(b[1:])
		return n + 1, err
	}
	return this.Conn.Read(b)
}

func (this *headBytesConn) GetUsefulNetConn() net.Conn{
	return this.Conn
}