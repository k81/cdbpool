package cdbpool

import (
	"context"
	"database/sql/driver"
	"fmt"

	"github.com/std0d9k81/kate/bigid"
	"github.com/std0d9k81/kate/log"
	"github.com/std0d9k81/kate/utils"
)

type beginExecutor struct {
	*RouteInfo
	ctx context.Context
	dbc *Conn
}

func (exr *beginExecutor) Run() (tx driver.Tx, err error) {
	if exr.dbc.client == nil || !exr.dbc.client.IsConnected() {
		return nil, driver.ErrBadConn
	}

	if exr.DBName == "" {
		exr.DBName = exr.dbc.DBName
	}

	var (
		req   *CdbPoolRequest
		resp  *CdbPoolResponse
		logId = fmt.Sprintf("%s.transaction.begin", exr.DBName)
	)

	req = &CdbPoolRequest{
		Logid:       logId,
		Command:     "transfer",
		Bigid:       exr.BigId,
		NeedSqlInfo: true,
		Req: &CdbPoolRequest_TransferReq{
			&TransferRequest{
				Dbname:  exr.DBName,
				Command: "begin",
			},
		},
	}

	if resp, err = exr.dbc.call(exr.ctx, req); err != nil {
		log.Error(exr.ctx, "db.transaction.begin", "logid", req.Logid, "vsid", bigid.GetVSId(exr.BigId), "conn_id", exr.dbc.id, "server_addr", exr.dbc.Addr, "error", err)
		return nil, driver.ErrBadConn
	}

	if ResultCode(resp.GetError()) != ResultCode_RC_SUCCESS {
		sqlInfo := resp.GetSqlInfo()
		if sqlInfo == nil {
			sqlInfo = &MysqlInfo{
				Sql:    "begin",
				Vsid:   utils.GetInt32(bigid.GetVSId(exr.BigId)),
				Dbname: exr.DBName,
			}
		}
		err = NewDBError(resp.GetError(), resp.GetErrMsg(), sqlInfo)
		log.Error(exr.ctx, "db.transaction.begin", "logid", req.Logid, "vsid", bigid.GetVSId(exr.BigId), "conn_id", exr.dbc.id, "server_addr", exr.dbc.Addr, "error", err)
		return
	}

	tx = newTx(exr.ctx, exr.dbc, exr.RouteInfo)

	return
}
