package logic

import (
	"net"
)

const (
	disconnect = iota
	connecting
	connected
)

type tcpSendInfo struct {
	data  []byte
	print bool
}

type socketTcp struct {
	status   int
	conn     net.Conn // 连接
	sink     dataSink
	dataChan chan *tcpSendInfo
}

func (p *socketTcp) connect(addr string) error {
	if p.status != disconnect {
		return nil
	}

	p.status = connecting
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		gLogger.Debug("connect failed:%s", err.Error())
		p.status = disconnect
		return err
	}
	p.status = connected
	p.conn = conn

	if p.sink != nil {
		p.sink.onConnectTcp()
	}
	go p.recvLoop()

	return nil
}

func (p *socketTcp) close() {
	if p.conn != nil {
		p.conn.Close()
		p.conn = nil
	}
	p.status = disconnect
}

func (p *socketTcp) send(data []byte, print bool) {
	p.dataChan <- &tcpSendInfo{data: data, print: print}
}

func (p *socketTcp) sendLoop() {
	for {
		select {
		case data, ok := <-p.dataChan:
			if ok {
				if p.conn != nil {
					if data.print {
						gLogger.Debug("tcp send data to %+v %+v %s", p.conn, data.data[0:5], string(data.data[5:]))
					}
					p.conn.Write(data.data)
				}
			}
		}
	}
}

func (p *socketTcp) recvLoop() {
	defer p.close()

	buf := make([]byte, bufferLength)
	for {
		n, err := p.conn.Read(buf)
		if err != nil {
			gLogger.Debug("recv error:%s", err.Error())
			if p.sink != nil {
				p.sink.onClose(err)
			}
			break
		}

		if p.sink != nil {
			p.sink.onReceiveTcp(buf[0:n])
		}
	}
}

func newSocketTcp(sink dataSink) *socketTcp {
	p := new(socketTcp)
	p.sink = sink
	p.dataChan = make(chan *tcpSendInfo, 100)
	go p.sendLoop()
	return p
}
