package cdbpool

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/std0d9k81/kate/bigid"
	"github.com/std0d9k81/kate/log"
	"github.com/std0d9k81/sqlparser"
	"github.com/std0d9k81/kate/utils"
)

type insertExecutor struct {
	*RouteInfo
	ctx     context.Context
	dbc     *Conn
	ast     *sqlparser.Insert
	table   string
	columns string
	values  string
}

func (exr *insertExecutor) Run() (result driver.Result, err error) {
	if err = exr.parse(); err != nil {
		return
	}

	var (
		req   *CdbPoolRequest
		resp  *CdbPoolResponse
		logId = fmt.Sprintf("%s.%s.insert", exr.DBName, exr.table)
	)

	req = &CdbPoolRequest{
		Logid:               logId,
		Command:             "ori_insert",
		Bigid:               exr.BigId,
		RequestOfflineMysql: false,
		NeedSqlInfo:         true,
		Req: &CdbPoolRequest_OriInsertReq{
			&OriInsertRequest{
				Dbname:  exr.DBName,
				Table:   exr.table,
				Columns: exr.columns,
				Values:  exr.values,
			},
		},
	}

	if resp, err = exr.dbc.call(exr.ctx, req); err != nil {
		log.Error(exr.ctx, "db.insert", "logid", req.Logid, "vsid", bigid.GetVSId(exr.BigId), "conn_id", exr.dbc.id, "server_addr", exr.dbc.Addr, "error", err)
		return nil, driver.ErrBadConn
	}

	if ResultCode(resp.GetError()) != ResultCode_RC_SUCCESS {
		sqlInfo := resp.GetSqlInfo()
		if sqlInfo == nil {
			sqlInfo = &MysqlInfo{
				Sql:    astValue(exr.ast),
				Vsid:   utils.GetInt32(bigid.GetVSId(exr.BigId)),
				Dbname: exr.DBName,
			}
		}
		err = NewDBError(resp.GetError(), resp.GetErrMsg(), sqlInfo)
		log.Error(exr.ctx, "db.insert", "logid", req.Logid, "vsid", bigid.GetVSId(exr.BigId), "conn_id", exr.dbc.id, "server_addr", exr.dbc.Addr, "error", err)
		return
	}

	insertResp := resp.GetInsertResp()
	if insertResp == nil {
		log.Error(exr.ctx, "db.insert", "logid", req.Logid, "vsid", bigid.GetVSId(exr.BigId), "conn_id", exr.dbc.id, "server_addr", exr.dbc.Addr, "error", "no insert response")
		return nil, driver.ErrBadConn
	}

	lastInsertId := int64(insertResp.GetLastInsertid())
	if lastInsertId == 0 {
		lastInsertId = int64(exr.BigId)
	}

	result = &Result{
		lastInsertId: lastInsertId,
		rowsAffected: int64(insertResp.GetAffectRows()),
	}

	return
}

func (exr *insertExecutor) parse() error {
	if exr.DBName == "" {
		exr.DBName = exr.dbc.DBName
	}

	exr.table = astValue(exr.ast.Table)
	exr.columns = astValue(exr.ast.Columns)
	exr.values = strings.TrimPrefix(astValue(exr.ast.Rows), "values ") + astValue(exr.ast.OnDup)
	return nil
}
