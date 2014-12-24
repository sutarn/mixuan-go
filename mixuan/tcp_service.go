package mixuan

import (
	"fmt"
	"log"
	"net"
	"runtime"
	"runtime/debug"
	"time"
)

type TcpService struct {
	timeout         time.Duration
	keepAlive       bool
	keepAlivePeriod time.Duration
	noDelay         bool
	readTimeout     time.Duration
	readBuffer      int
	writeTimeout    time.Duration
	writeBuffer     int
}

func NewTcpService() *TcpService {
	return &TcpService{}
}

func (service *TcpService) SetTimeout(d time.Duration) {
	service.timeout = d
}

func (service *TcpService) SetKeepAlive(keepalive bool) {
	service.keepAlive = keepalive
}

func (service *TcpService) SetKeepAlivePeriod(d time.Duration) {
	service.keepAlivePeriod = d
}

func (service *TcpService) SetNoDelay(noDelay bool) {
	service.noDelay = noDelay
}

func (service *TcpService) SetReadTimeout(d time.Duration) {
	service.readTimeout = d
}

func (service *TcpService) SetReadBuffer(bytes int) {
	service.readBuffer = bytes
}

func (service *TcpService) SetWriteTimeout(d time.Duration) {
	service.writeTimeout = d
}

func (service *TcpService) SetWriteBuffer(bytes int) {
	service.writeBuffer = bytes
}

func (service *TcpService) Handle(data []byte, context interface{}) (output []byte) {
	strText := fmt.Sprintf("back:%s", string(data))
	log.Println(strText)
	return []byte(strText)
}

func (service *TcpService) ServeTCP(conn *net.TCPConn) (err error) {
	if 0 < service.timeout {
		if err = conn.SetDeadline(time.Now().Add(service.timeout)); nil != err {
			return err
		}
	}
	if true == service.keepAlive {
		if err = conn.SetKeepAlive(service.keepAlive); nil != err {
			return err
		}
	}
	if 0 < service.keepAlivePeriod {
		if err = conn.SetKeepAlivePeriod(service.keepAlivePeriod); nil != err {
			return err
		}
	}
	if true == service.noDelay {
		if err = conn.SetNoDelay(service.noDelay); nil != err {
			return err
		}
	}
	if 0 < service.readTimeout {
		if err = conn.SetReadDeadline(time.Now().Add(service.readTimeout)); nil != err {
			return err
		}
	}
	if 0 < service.readBuffer {
		if err = conn.SetReadBuffer(service.readBuffer); nil != err {
			return err
		}
	}
	if 0 < service.writeTimeout {
		if err = conn.SetWriteDeadline(time.Now().Add(service.writeTimeout)); nil != err {
			return err
		}
	}
	if 0 < service.writeBuffer {
		if err = conn.SetWriteBuffer(service.writeBuffer); nil != err {
			return err
		}
	}
	go func(conn net.Conn) {
		fmt.Printf("Client Address:%s.\n", conn.RemoteAddr().String())
		var err error
		var data []byte
		for {
			if 0 < service.readTimeout {
				err = conn.SetReadDeadline(time.Now().Add(service.readTimeout))
			}
			if nil == err {
				data, err = receiveDataOverTcp(conn)
			}
			if err == nil {
				data = service.Handle(data, conn)
				if 0 < service.writeTimeout {
					err = conn.SetWriteDeadline(time.Now().Add(service.writeTimeout))
				}
				if err == nil {
					err = sendDataOverTcp(conn, data)
				}
			}
			if err != nil {
				conn.Close()
				break
			}
		}
	}(conn)
	return nil
}

func (service *TcpService) fireErrorEvent(err error, context interface{}) {

}

type TcpServer struct {
	*TcpService
	Port         uint
	ThreadCount  int
	DebugEnabled bool
	listener     *net.TCPListener
}

func NewTcpServer(port uint) *TcpServer {
	if 0 == port {
		port = 80
	}
	runtime.GOMAXPROCS(runtime.NumCPU())
	return &TcpServer{
		TcpService:  NewTcpService(),
		Port:        port,
		ThreadCount: runtime.NumCPU(),
		listener:    nil,
	}
}

func (server *TcpServer) handle() (err error) {
	defer func() {
		if e := recover(); e != nil && err == nil {
			if server.DebugEnabled {
				err = fmt.Errorf("%v\r\n%s", e, debug.Stack())
			} else {
				err = fmt.Errorf("%v", e)
			}
		}
	}()
	if nil == server.listener {
		return nil
	}

	conn, err := server.listener.AcceptTCP()
	if nil != err {
		return err
	}
	return server.ServeTCP(conn)
}

func (server *TcpServer) Start() (err error) {
	if nil == server.listener {
		var addr *net.TCPAddr
		if addr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", server.Port)); err != nil {
			return err
		}
		if server.listener, err = net.ListenTCP("tcp", addr); nil != err {
			return nil
		}
		for {
			if nil != server.listener {
				if err := server.handle(); err != nil {
					server.fireErrorEvent(err, nil)
				}
			} else {
				break
			}
		}
	}
	return nil
}

func (server *TcpServer) Stop() {
	if server.listener != nil {
		listener := server.listener
		server.listener = nil
		listener.Close()
	}
}
