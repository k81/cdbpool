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

type selectExecutor struct {
	*RouteInfo
	ctx       context.Context
	dbc       *Conn
	ast       *sqlparser.Select
	table     string
	columns   string
	filters   string
	orderBy   string
	limit     string
	forUpdate int32
}

func (exr *selectExecutor) Run() (rows driver.Rows, err error) {
	if err = exr.parse(); err != nil {
		return
	}

	var (
		req   *CdbPoolRequest
		resp  *CdbPoolResponse
		logId = fmt.Sprintf("%s.%s.select", exr.DBName, exr.table)
	)

	req = &CdbPoolRequest{
		Logid:               logId,
		Command:             "ori_select",
		Bigid:               exr.BigId,
		RequestOfflineMysql: exr.Offline,
		NeedSqlInfo:         true,
		Req: &CdbPoolRequest_OriSelectReq{
			&OriSelectRequest{
				Dbname:        exr.DBName,
				Table:         exr.table,
				Columns:       exr.columns,
				ComplexFilter: exr.filters,
				Orderby:       exr.orderBy,
				Limit:         exr.limit,
				Forupdate:     exr.forUpdate,
			},
		},
	}

	if resp, err = exr.dbc.call(exr.ctx, req); err != nil {
		log.Error(exr.ctx, "db.select", "logid", req.Logid, "vsid", bigid.GetVSId(exr.BigId), "conn_id", exr.dbc.id, "server_addr", exr.dbc.Addr, "error", err)
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
		log.Error(exr.ctx, "db.select", "logid", req.Logid, "vsid", bigid.GetVSId(exr.BigId), "conn_id", exr.dbc.id, "server_addr", exr.dbc.Addr, "error", err)
		return
	}

	selectResp := resp.GetSelectResp()
	if selectResp == nil {
		log.Error(exr.ctx, "db.select", "logid", req.Logid, "vsid", bigid.GetVSId(exr.BigId), "conn_id", exr.dbc.id, "server_addr", exr.dbc.Addr, "error", "no select response")
		return nil, driver.ErrBadConn
	}

	rows = &Rows{
		records: selectResp.GetRecords(),
	}
	return
}

func (exr *selectExecutor) parse() error {
	var (
		groupBy string
		having  string
	)
	if exr.ast.Lock != "" {
		if exr.ast.Lock != " for update" {
			panic(ErrSqlInvalid(fmt.Sprintf("lock type `%v` not supported", exr.ast.Lock)))
		}
		exr.forUpdate = 1
	}

	if len(exr.ast.GroupBy) > 0 {
		if !exr.Offline {
			panic(ErrSqlInvalid("`group by` is only supported for offline db"))
		} else {
			groupBy = astValue(exr.ast.GroupBy)
		}
	}

	if exr.ast.Having != nil {
		if !exr.Offline {
			panic(ErrSqlInvalid("`having` is only supported for offline db"))
		} else {
			having = astValue(exr.ast.Having)
		}
	}

	if exr.ast.Where == nil {
		panic(ErrSqlInvalid("missing `where`"))
	}

	if exr.ast.Where.Type != "where" {
		panic(ErrSqlInvalid(fmt.Sprintf("filters `%v` not supported", exr.ast.Where.Type)))
	}

	if exr.DBName == "" {
		exr.DBName = exr.dbc.DBName
	}

	exr.columns = fmt.Sprint(exr.ast.Distinct, astValue(exr.ast.SelectExprs))
	exr.table = astValue(exr.ast.From)
	exr.filters = fmt.Sprintf("%s%s%s", astValue(exr.ast.Where.Expr), groupBy, having)
	exr.orderBy = strings.TrimPrefix(astValue(exr.ast.OrderBy), " order by ")
	exr.limit = strings.TrimPrefix(astValue(exr.ast.Limit), " limit ")
	return nil
}
