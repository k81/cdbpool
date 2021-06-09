package cdbpool

import (
	"database/sql/driver"
	"fmt"

	"github.com/std0d9k81/kate/bigid"
	"github.com/std0d9k81/kate/log"
	"github.com/std0d9k81/kate/utils"
)

type commitExecutor struct {
	*Tx
}

func (exr *commitExecutor) Run() (err error) {
	if exr.dbc.client == nil || !exr.dbc.client.IsConnected() {
		return driver.ErrBadConn
	}

	if exr.DBName == "" {
		exr.DBName = exr.dbc.DBName
	}

	var (
		req   *CdbPoolRequest
		resp  *CdbPoolResponse
		logId = fmt.Sprintf("%s.transaction.commit", exr.DBName)
	)

	req = &CdbPoolRequest{
		Logid:       logId,
		Command:     "transfer",
		Bigid:       exr.BigId,
		NeedSqlInfo: true,
		Req: &CdbPoolRequest_TransferReq{
			&TransferRequest{
				Dbname:  exr.DBName,
				Command: "commit",
			},
		},
	}

	if resp, err = exr.dbc.call(exr.ctx, req); err != nil {
		log.Error(exr.ctx, "db.transaction.commit", "logid", req.Logid, "vsid", bigid.GetVSId(exr.BigId), "conn_id", exr.dbc.id, "server_addr", exr.dbc.Addr, "error", err)
		return driver.ErrBadConn
	}

	if ResultCode(resp.GetError()) != ResultCode_RC_SUCCESS {
		sqlInfo := resp.GetSqlInfo()
		if sqlInfo == nil {
			sqlInfo = &MysqlInfo{
				Sql:    "commit",
				Vsid:   utils.GetInt32(bigid.GetVSId(exr.BigId)),
				Dbname: exr.DBName,
			}
		}
		err = NewDBError(resp.GetError(), resp.GetErrMsg(), sqlInfo)
		log.Error(exr.ctx, "db.transaction.commit", "logid", req.Logid, "vsid", bigid.GetVSId(exr.BigId), "conn_id", exr.dbc.id, "server_addr", exr.dbc.Addr, "error", err)
		return
	}
	return
}
