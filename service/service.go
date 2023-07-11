package service

import (
	"encoding/json"
	"errors"
	"fmt"
	cmap "github.com/orcaman/concurrent-map/v2"
	"go.uber.org/zap"
	"net"
	"sync"
	"tcp-transmit/utils"
	"unsafe"
)

type TcpTransmit struct {
	logger     *zap.Logger
	ConnRemote net.Conn     //目标TCP连接
	Listener   net.Listener //监听的端口
	Clients    cmap.ConcurrentMap[*net.Conn, bool]
	ClientsBuf chan ConnMsgT
	TargetBuf  chan ConnMsgT

	wg        sync.WaitGroup
	closeChan chan struct{}
	IsClose   bool
	closeMux  sync.Mutex
}

type ConnMsgT struct {
	IpAddr string `json:"ipAddr"`
	Msg    string `json:"msg"`
}

func NewTcpTransmit() *TcpTransmit {
	s := &TcpTransmit{
		Clients: cmap.NewWithCustomShardingFunction[*net.Conn, bool](func(key *net.Conn) uint32 {
			return uint32(uintptr(unsafe.Pointer(key)))
		}),
		ClientsBuf: make(chan ConnMsgT, 2048),
		TargetBuf:  make(chan ConnMsgT, 2048),
		closeChan:  make(chan struct{}),
		logger:     zap.NewExample(),
	}
	s.SetLogger(s, s.logger)
	return s
}

func (t *TcpTransmit) SetLogger(obj interface{}, log *zap.Logger) {
	t.logger = log.Named(utils.GetType(obj))
}

func (t *TcpTransmit) Start(targetIp, targetPort, listenIp, listenPort string) error {
	conn, err := net.Dial("tcp", targetIp+":"+targetPort)
	if err != nil {
		return err
	}
	t.ConnRemote = conn

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", listenIp, listenPort))
	if err != nil {
		return err
	}
	t.Listener = listener

	go t.TargetMsgProc()

	go func() {
		t.wg.Add(1)
		defer func() {
			t.wg.Done()
			t.Close()
		}()
		for {
			// 等待连接
			client, err := listener.Accept()
			if err != nil {
				t.logger.Info("accept err")
				return
			}
			t.Clients.Set(&client, true)
			go t.ClientMsgProc(&client)
		}
	}()

	go t.WaitClose()

	t.logger.Info("TcpTransmit start success", zap.Any("serviceId", fmt.Sprintf("%v", &t)))
	return nil
}

func (t *TcpTransmit) Stop() {
	t.Close()
	return
}

func (t *TcpTransmit) Close() {
	t.closeMux.Lock()
	defer t.closeMux.Unlock()
	if t.IsClose {
		return
	}
	if t.ConnRemote != nil {
		_ = t.ConnRemote.Close()
	}
	if t.Listener != nil {
		_ = t.Listener.Close()
	}
	for k := range t.Clients.Items() {
		conn := *k
		_ = conn.Close()
	}
	t.IsClose = true
	close(t.closeChan)
}

func (t *TcpTransmit) TargetMsgProc() {
	t.wg.Add(1)
	defer func() {
		t.wg.Done()
		t.Close()
	}()

	for {
		msg := make([]byte, 10240)
		n, err := t.ConnRemote.Read(msg)
		if err != nil {
			t.logger.Info("TargetMsg read err")
			return
		}
		if n > 0 {
			//t.logger.Info("TargetMsg read", zap.String("msg", string(msg[:n])))
		}

		for k := range t.Clients.Items() {
			con := *k
			_, _ = con.Write(msg[:n])
		}
		t.TargetBuf <- ConnMsgT{
			IpAddr: t.ConnRemote.RemoteAddr().String(),
			Msg:    string(msg[:n]),
		}
	}
}

func (t *TcpTransmit) ClientMsgProc(cli *net.Conn) {
	conn := *cli
	t.wg.Add(1)
	defer func() {
		t.wg.Done()
		t.Clients.Remove(cli)
		_ = conn.Close()
	}()

	for {
		msg := make([]byte, 10240)
		n, err := conn.Read(msg)
		if err != nil {
			t.logger.Info("TargetMsg read err")
			return
		}

		_, _ = t.ConnRemote.Write(msg[:n])
		t.ClientsBuf <- ConnMsgT{
			IpAddr: conn.RemoteAddr().String(),
			Msg:    string(msg[:n]),
		}
	}
}

func (t *TcpTransmit) ReadRemoteChanMsg() ([]byte, error) {
	select {
	case msg := <-t.TargetBuf:
		result, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		} else {
			return result, nil
		}
	case <-t.closeChan:
		return nil, errors.New("remote msg chan close")
	default:
		return nil, errors.New("empty")
	}
}

func (t *TcpTransmit) ReadClientChanMsg() ([]byte, error) {
	select {
	case msg := <-t.ClientsBuf:
		result, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		} else {
			return result, nil
		}
	case <-t.closeChan:
		return nil, errors.New("clients msg chan close")
	default:
		return nil, errors.New("empty")
	}
}

func (t *TcpTransmit) WaitClose() {
	<-t.closeChan
	t.wg.Wait()
	t.logger.Info("TcpTransmit Close success", zap.Any("serviceId", fmt.Sprintf("%v", &t)))
}

func (t *TcpTransmit) GetClientsList() string {
	type result struct {
		ClientsList []string `json:"clientsList"`
	}
	var resultData result
	for k := range t.Clients.Items() {
		resultData.ClientsList = append(resultData.ClientsList, (*k).RemoteAddr().String())
	}
	res, _ := json.Marshal(resultData)
	return string(res)
}
