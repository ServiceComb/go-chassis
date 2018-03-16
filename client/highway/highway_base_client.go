package highway

import (
	"crypto/tls"
	"errors"
	"github.com/ServiceComb/go-chassis/core/lager"
	"net"
	"sync"
	"time"
)

//Highway connect parmas
type ConnParams struct {
	Addr      string
	TlsConfig *tls.Config
	Timeout   time.Duration
	ConnNum   int
}

//Highway base client
type HighwayBaseClient struct {
	addr          string
	mtx           sync.Mutex
	mapMutex      sync.Mutex
	msgWaitRspMap map[uint64]*InvocationContext
	highwayConns  []*HighwayClientConnection
	closed        bool
	connParams    *ConnParams
}

//Client cache
var CachedClients *ClientMgr

func init() {
	CachedClients = newClientMgr()
}

//Highway context
type InvocationContext struct {
	Req  *HighwayRequest
	Rsp  *HighwayRespond
	Wait *chan int
}

//Notify done.
func (this *InvocationContext) Done() {
	*this.Wait <- 1
}

//Client manage
type ClientMgr struct {
	mapMutex sync.Mutex
	clients  map[string]*HighwayBaseClient
}

func newClientMgr() *ClientMgr {
	tmp := new(ClientMgr)
	tmp.clients = make(map[string]*HighwayBaseClient)
	return tmp
}

//Obtain  client
func (mgr *ClientMgr) GetClient(connParmas *ConnParams) (*HighwayBaseClient, error) {
	mgr.mapMutex.Lock()
	defer mgr.mapMutex.Unlock()
	if tmp, ok := mgr.clients[connParmas.Addr]; ok {
		if !tmp.Closed() {
			lager.Logger.Info("GetClient from cached addr:" + connParmas.Addr)
			return tmp, nil
		} else {
			delete(mgr.clients, connParmas.Addr)
		}
	}

	lager.Logger.Info("GetClient from new open addr:" + connParmas.Addr)
	tmp := newHighwayBaseClient(connParmas)
	err := tmp.Open()
	if err != nil {
		return nil, err
	} else {
		mgr.clients[connParmas.Addr] = tmp
		return tmp, nil
	}
}

func newHighwayBaseClient(connParmas *ConnParams) *HighwayBaseClient {
	tmp := &HighwayBaseClient{}
	tmp.addr = connParmas.Addr
	tmp.closed = true
	tmp.connParams = connParmas
	tmp.msgWaitRspMap = make(map[uint64]*InvocationContext)
	return tmp
}

//Obtain the address
func (baseClient *HighwayBaseClient) GetAddr() string {
	return baseClient.addr
}

func (baseClient *HighwayBaseClient) makeConnection() (*HighwayClientConnection, error) {
	var baseConn net.Conn
	var errDial error

	if baseClient.connParams.TlsConfig != nil {
		dialer := &net.Dialer{Timeout: baseClient.connParams.Timeout * time.Second}
		baseConn, errDial = tls.DialWithDialer(dialer, "tcp", baseClient.addr, baseClient.connParams.TlsConfig)
	} else {
		baseConn, errDial = net.DialTimeout("tcp", baseClient.addr, baseClient.connParams.Timeout*time.Second)
	}
	if errDial != nil {
		lager.Logger.Error("the addr: "+baseClient.addr, errDial)
		return nil, errDial
	}
	higwayConn := NewHighwayClientConnection(baseConn, baseClient)
	err := higwayConn.Open()
	if err != nil {
		lager.Logger.Error("higwayConn open: "+baseClient.addr, errDial)
		return nil, err
	}

	return higwayConn, nil
}

func (baseClient *HighwayBaseClient) initConns() error {
	if baseClient.connParams.ConnNum == 0 {
		baseClient.connParams.ConnNum = 4
	}

	baseClient.highwayConns = make([]*HighwayClientConnection, baseClient.connParams.ConnNum)
	for i := 0; i < baseClient.connParams.ConnNum; i++ {
		higwayConn, err := baseClient.makeConnection()
		if err != nil {
			return err
		}
		baseClient.highwayConns[i] = higwayConn
	}
	return nil
}

//open  client
func (baseClient *HighwayBaseClient) Open() error {
	baseClient.mtx.Lock()
	defer baseClient.mtx.Unlock()
	err := baseClient.initConns()
	if err != nil {
		baseClient.clearConns()
		return err
	}
	baseClient.closed = false
	return nil
}

//close client
func (baseClient *HighwayBaseClient) Close() {
	baseClient.mtx.Lock()
	defer baseClient.mtx.Unlock()
	baseClient.close()
}

//close client, no mutex
func (baseClient *HighwayBaseClient) close() {
	if baseClient.closed {
		return
	}
	baseClient.mapMutex.Lock()
	for _, v := range baseClient.msgWaitRspMap {
		v.Done()
	}
	baseClient.msgWaitRspMap = make(map[uint64]*InvocationContext)
	baseClient.mapMutex.Unlock()
	baseClient.clearConns()
	baseClient.closed = true
}

func (baseClient *HighwayBaseClient) clearConns() {
	for i := 0; i < baseClient.connParams.ConnNum; i++ {
		conn := baseClient.highwayConns[i]
		if conn != nil {
			conn.Close()
			baseClient.highwayConns[i] = nil
		}
	}
}

func (baseClient *HighwayBaseClient) AddWaitMsg(msgID uint64, result *InvocationContext) {
	baseClient.mapMutex.Lock()
	if baseClient.msgWaitRspMap != nil {
		baseClient.msgWaitRspMap[msgID] = result
	}
	baseClient.mapMutex.Unlock()
}

func (baseClient *HighwayBaseClient) RemoveWaitMsg(msgID uint64) {
	baseClient.mapMutex.Lock()
	if baseClient.msgWaitRspMap != nil {
		delete(baseClient.msgWaitRspMap, msgID)
	}
	baseClient.mapMutex.Unlock()
}

func (baseClient *HighwayBaseClient) Send(req *HighwayRequest, rsp *HighwayRespond, timeout time.Duration) error {
	if baseClient.closed {
		baseClient.mtx.Lock()
		if baseClient.closed {
			baseClient.mtx.Unlock()
			return errors.New("Client is closed.")
		}
		baseClient.mtx.Unlock()
	}

	msgID := req.MsgID
	idx := msgID % uint64(baseClient.connParams.ConnNum)
	highwayConn := baseClient.highwayConns[idx]
	if highwayConn == nil || highwayConn.Closed() {
		baseClient.mtx.Lock()
		highwayConn = baseClient.highwayConns[idx]
		if highwayConn == nil || highwayConn.Closed() {
			highwayConnTmp, err := baseClient.makeConnection()
			if err != nil {
				baseClient.mtx.Unlock()
				return err
			}
			highwayConn = highwayConnTmp
			baseClient.highwayConns[idx] = highwayConn
		}
		baseClient.mtx.Unlock()
	}
	if req.TwoWay {
		wait := make(chan int)
		ctx := &InvocationContext{req, rsp, &wait}
		baseClient.AddWaitMsg(msgID, ctx)

		err := highwayConn.AsyncSendMsg(ctx)
		if err != nil {
			rsp.Err = err.Error()
			lager.Logger.Error("AsyncSendMsg err:", err)
			return err
		}

		var bTimeout bool = false
		select {
		case <-wait:
			bTimeout = false
		case <-time.After(timeout * time.Second):
			bTimeout = true
		}

		baseClient.RemoveWaitMsg(msgID)
		close(wait)
		if bTimeout {
			rsp.Err = "Client send timeout"
			return errors.New("Client send timeout")
		} else {
			return nil
		}
	} else {
		// Respond of postMsg  is  needless
		err := highwayConn.PostMsg(req)
		if err != nil {
			lager.Logger.Error("PostMsg err:", err)
			return err
		}
	}
	return nil
}

func (baseClient *HighwayBaseClient) GetWaitMsg(msgID uint64) *InvocationContext {
	baseClient.mapMutex.Lock()
	defer baseClient.mapMutex.Unlock()

	if _, ok := baseClient.msgWaitRspMap[msgID]; ok {
		return baseClient.msgWaitRspMap[msgID]
	}
	return nil
}

func (baseClient *HighwayBaseClient) Closed() bool {
	return baseClient.closed
}
