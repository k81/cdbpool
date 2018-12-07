package cdbpool

import (
	"context"
	"database/sql/driver"
	"fmt"

	"github.com/k81/kate/bigid"
	"github.com/k81/kate/log"
	"github.com/k81/sqlparser"
	"github.com/k81/kate/utils"
)

type updateExecutor struct {
	*RouteInfo
	ctx     context.Context
	dbc     *Conn
	ast     *sqlparser.Update
	table   string
	values  string
	filters string
}

func (exr *updateExecutor) Run() (result driver.Result, err error) {
	if err = exr.parse(); err != nil {
		return
	}

	var (
		req   *CdbPoolRequest
		resp  *CdbPoolResponse
		logId = fmt.Sprintf("%s.%s.update", exr.DBName, exr.table)
	)

	req = &CdbPoolRequest{
		Logid:       logId,
		Command:     "ori_update",
		Bigid:       exr.BigId,
		NeedSqlInfo: true,
		Req: &CdbPoolRequest_OriUpdateReq{
			&OriUpdateRequest{
				Dbname:        exr.DBName,
				Table:         exr.table,
				Sets:          exr.values,
				ComplexFilter: exr.filters,
			},
		},
	}

	if resp, err = exr.dbc.call(exr.ctx, req); err != nil {
		log.Error(exr.ctx, "db.update", "logid", req.Logid, "vsid", bigid.GetVSId(exr.BigId), "conn_id", exr.dbc.id, "server_addr", exr.dbc.Addr, "error", err)
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
		log.Error(exr.ctx, "db.update", "logid", req.Logid, "vsid", bigid.GetVSId(exr.BigId), "conn_id", exr.dbc.id, "server_addr", exr.dbc.Addr, "error", err)
		return
	}

	updateResp := resp.GetUpdateResp()
	if updateResp == nil {
		log.Error(exr.ctx, "db.update", "logid", req.Logid, "vsid", bigid.GetVSId(exr.BigId), "conn_id", exr.dbc.id, "server_addr", exr.dbc.Addr, "error", "no update response")
		return nil, driver.ErrBadConn
	}

	result = &Result{
		rowsAffected: int64(updateResp.GetAffectRows()),
		lastInsertId: int64(updateResp.GetLastInsertid()),
	}
	return
}

func (exr *updateExecutor) parse() error {
	if exr.ast.Where == nil {
		panic(ErrSqlInvalid("must have `where` conditions"))
	}

	if exr.ast.Where.Type != "where" {
		panic(ErrSqlInvalid(fmt.Sprintf("filters `%v` not supported", exr.ast.Where.Type)))
	}

	if exr.ast.Limit != nil {
		panic(ErrSqlInvalid("`limit` not supported"))
	}

	if len(astValue(exr.ast.OrderBy)) > 0 {
		panic(ErrSqlInvalid("`order by` not supported"))
	}

	if exr.DBName == "" {
		exr.DBName = exr.dbc.DBName
	}

	exr.table = astValue(exr.ast.Table)
	exr.values = astValue(exr.ast.Exprs)
	exr.filters = astValue(exr.ast.Where.Expr)
	return nil
}
