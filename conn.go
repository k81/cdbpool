package cdbpool

import (
	"context"
	"database/sql/driver"
	"fmt"
	"sync/atomic"

	"github.com/k81/kate/log"
	"github.com/k81/knet"
)

var nextConnId = uint64(0)

type Conn struct {
	knet.IoHandlerAdapter
	*Config
	id     uint64
	client knet.Client
	ctx    context.Context
}

func newConn() *Conn {
	return &Conn{
		id: atomic.AddUint64(&nextConnId, 1),
	}
}

func (c *Conn) Prepare(query string) (driver.Stmt, error) {
	panic(ErrDeprecated)
}

func (c *Conn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	if c.client == nil || !c.client.IsConnected() {
		return nil, driver.ErrBadConn
	}

	if GetRoute(ctx) == nil {
		if c.ctx == nil {
			panic(ErrMissingRouteInfo)
		}
		ctx = c.ctx
	}

	return newStmt(ctx, c, query), nil
}

func (c *Conn) Close() error {
	if Debug {
		log.Debug(mctx, "close connection", "conn_id", c.id)
	}

	if c.client != nil {
		c.client.Close()
	}

	if Debug {
		log.Debug(mctx, "close connection end", "conn_id", c.id)
	}
	return nil
}

func (c *Conn) Begin() (driver.Tx, error) {
	panic(ErrDeprecated)
}

func (c *Conn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	route := GetRoute(ctx)
	if route == nil {
		panic(ErrMissingRouteInfo)
	}

	exr := &beginExecutor{
		RouteInfo: route,
		ctx:       ctx,
		dbc:       c,
	}
	return exr.Run()
}

func (c *Conn) OnIdle(session *knet.IoSession) error {
	if session.GetIdleCount() > 2 {
		return knet.ErrPeerDead
	}

	pkt := newPingPacket()
	session.Send(mctx, pkt)
	return nil
}

func (c *Conn) OnError(session *knet.IoSession, err error) {
	log.Error(mctx, "connection error", "conn_id", c.id, "server_addr", c.Addr, "error", err)
}

func (c *Conn) OnDisconnected(session *knet.IoSession) {
	log.Info(mctx, "disconnected", "conn_id", c.id, "server_addr", c.Addr)
}

func (c *Conn) Ping(ctx context.Context) error {
	pkt := newPingPacket()
	if _, err := c.client.Call(ctx, pkt); err != nil {
		return driver.ErrBadConn
	}

	return nil
}

func (c *Conn) call(ctx context.Context, req *CdbPoolRequest) (resp *CdbPoolResponse, err error) {
	var (
		seq     = nextRequestId()
		pkt     *Packet
		reply   interface{}
		session *knet.IoSession
		ok      bool
	)

	if c.client == nil || !c.client.IsConnected() {
		log.Error(ctx, "server not connected", "conn_id", c.id, "server_addr", c.Addr)
		return nil, driver.ErrBadConn
	}

	session = c.client.GetSession()

	req.Logid = fmt.Sprintf("%s.%v", req.Logid, seq)
	pkt = newQueryPacket(seq, req)

	if Debug {
		log.Debug(ctx, "cdbpool conn.call begin",
			"conn_id", c.id,
			"local_addr", session.LocalAddr(),
			"remote_addr", session.RemoteAddr(),
			"log_id", req.Logid,
		)
	}

	if reply, err = c.client.Call(ctx, pkt); err != nil {
		log.Error(ctx, "cdbpool conn.call",
			"conn_id", c.id,
			"local_addr", session.LocalAddr(),
			"server_addr", c.Addr,
			"log_id", req.Logid,
			"error", err,
		)
		return nil, driver.ErrBadConn
	}

	if Debug {
		log.Debug(ctx, "cdbpool conn.call end",
			"conn_id", c.id,
			"local_addr", session.LocalAddr(),
			"remote_addr", session.RemoteAddr(),
			"log_id", req.Logid,
		)
	}

	if resp, ok = reply.(*Packet).Message.(*CdbPoolResponse); !ok {
		log.Error(ctx, "cdbpool conn.call",
			"conn_id", c.id,
			"local_addr", session.LocalAddr(),
			"server_addr", c.Addr,
			"log_id", req.Logid,
			"error", "invalid response",
		)
		return nil, driver.ErrBadConn
	}
	return
}
