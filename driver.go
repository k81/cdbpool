package cdbpool

import (
	"context"
	"database/sql"
	"database/sql/driver"

	"github.com/k81/kate/log"
	"github.com/k81/knet"
)

var (
	mctx  = log.SetContext(context.Background(), "module", "cdbpool")
	Debug bool
)

type CdbPoolDriver struct{}

func (d CdbPoolDriver) Open(dsn string) (conn driver.Conn, err error) {
	dbc := newConn()

	if Debug {
		log.Debug(mctx, "driver.Open()", "conn_id", dbc.id)
	}

	if dbc.Config, err = ParseDSN(dsn); err != nil {
		return
	}

	conf := knet.NewTCPClientConfig()
	if dbc.Timeout > 0 {
		conf.DialTimeout = dbc.Timeout
	}
	if dbc.ReadTimeout > 0 {
		conf.Io.ReadTimeout = dbc.ReadTimeout
	}
	if dbc.WriteTimeout > 0 {
		conf.Io.WriteTimeout = dbc.WriteTimeout
	}

	dbc.client = knet.NewTCPClient(mctx, conf)
	if dbc.EnableCircuitBreaker {
		dbc.client = knet.NewCircuitBreakerClient(dbc.client, getBreaker(dbc.Addr))
	}

	dbc.client.SetProtocol(&Protocol{})
	dbc.client.SetIoHandler(dbc)

	if Debug {
		log.Debug(mctx, "dail to server begin", "conn_id", dbc.id, "remote_addr", dbc.Addr)
	}

	if err = dbc.client.Dial(dbc.Addr); err != nil {
		dbc.client.Close()
		return
	}

	if Debug {
		log.Debug(mctx, "dail to server end", "conn_id", dbc.id, "remote_addr", dbc.Addr)
	}

	conn = dbc
	return
}

func init() {
	sql.Register("cdbpool", &CdbPoolDriver{})
}
