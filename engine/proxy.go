package engine

import (
	"net"
	"io"
	"sync"
	"github.com/dotcloud/beam"
)

func Proxy(frontend net.Listener, backend beam.Connector) error {
	for {
		fConn, err := frontend.Accept()
		if err != nil {
			return err
		}
		go func(fConn net.Conn) {
			ProxyConn(fConn, backend)
			fConn.Close()
		}(fConn)
	}
	return nil
}

func ProxyConn(fConn net.Conn, backend beam.Connector) error {
	bConn, err := backend.Connect()
	if err != nil {
		return err
	}
	defer bConn.Close()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		io.Copy(fConn, bConn)
		wg.Add(-1)
	}()
	go func() {
		io.Copy(bConn, fConn)
		wg.Add(-1)
	}()
	wg.Wait()
	return nil
}

type PipeHub struct {
	conns chan net.Conn
	err error
}

func NewPipeHub() *PipeHub {
	return &PipeHub{
		conns: make(chan net.Conn),
	}
}

func (p *PipeHub) Accept() (net.Conn, error) {
	if p.err != nil {
		return nil, p.err
	}
	return <-p.conns, nil
}

func (p *PipeHub) Close() error {
	if p.err != nil {
		return p.err
	}
	p.err = io.EOF
	close(p.conns)
	return nil
}

func (p *PipeHub) Addr() net.Addr {
	return p
}

func (p *PipeHub) Network() string {
	return "pipe"
}

func (p *PipeHub) String() string {
	return "pipe"
}

func (p *PipeHub) Connect() (net.Conn, error) {
	if p.err != nil {
		return nil, p.err
	}
	handzel, gretel := net.Pipe()
	p.conns <- gretel
	return handzel, nil
}
