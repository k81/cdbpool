package cdbpool

import (
	"bytes"
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"

	"github.com/std0d9k81/sqlparser"
)

var (
	ErrSqlNotSupport    = errors.New("sql not supported")
	ErrDeprecated       = errors.New("interface deprecated")
	ErrMissingRouteInfo = errors.New("missing route info")
	ErrSqlTooLarge      = errors.New("sql too large")
)

type Stmt struct {
	ctx        context.Context
	dbc        *Conn
	query      string
	paramCount int
}

func newStmt(ctx context.Context, dbc *Conn, query string) *Stmt {
	stmt := &Stmt{
		ctx:        ctx,
		dbc:        dbc,
		query:      query,
		paramCount: strings.Count(query, "?"),
	}
	return stmt
}

func (stmt *Stmt) Close() error {
	if stmt.dbc == nil || stmt.dbc.client == nil || !stmt.dbc.client.IsConnected() {
		return driver.ErrBadConn
	}
	return nil
}

func (stmt *Stmt) NumInput() int {
	return stmt.paramCount
}

func (stmt *Stmt) Exec(args []driver.Value) (driver.Result, error) {
	return stmt.ExecContext(stmt.ctx, driverArgs(args))
}

func (stmt *Stmt) Query(args []driver.Value) (driver.Rows, error) {
	return stmt.QueryContext(stmt.ctx, driverArgs(args))
}

func (stmt *Stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (result driver.Result, err error) {
	var (
		route     *RouteInfo
		q         string
		statement sqlparser.Statement
	)

	if route = GetRoute(ctx); route == nil {
		if stmt.dbc.ctx != nil {
			ctx = stmt.dbc.ctx
			route = GetRoute(ctx)
		}
	}

	if route == nil {
		panic(ErrMissingRouteInfo)
	}

	if q, err = stmt.interpolateParams(args); err != nil {
		return
	}

	if statement, err = sqlparser.Parse(q); err != nil {
		return
	}

	switch ast := statement.(type) {
	case *sqlparser.Update:
		exr := &updateExecutor{
			RouteInfo: route,
			ctx:       ctx,
			dbc:       stmt.dbc,
			ast:       ast,
		}
		return exr.Run()
	case *sqlparser.Insert:
		exr := &insertExecutor{
			RouteInfo: route,
			ctx:       ctx,
			dbc:       stmt.dbc,
			ast:       ast,
		}
		return exr.Run()
	case *sqlparser.Delete:
		exr := &deleteExecutor{
			RouteInfo: route,
			ctx:       ctx,
			dbc:       stmt.dbc,
			ast:       ast,
		}
		return exr.Run()
	default:
		panic(ErrSqlNotSupport)
	}
}

func (stmt *Stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (rows driver.Rows, err error) {
	var (
		route     *RouteInfo
		q         string
		statement sqlparser.Statement
		ast       *sqlparser.Select
		ok        bool
	)

	if route = GetRoute(ctx); route == nil {
		if stmt.dbc.ctx != nil {
			ctx = stmt.dbc.ctx
			route = GetRoute(ctx)
		}
	}

	if route == nil {
		panic(ErrMissingRouteInfo)
	}

	if q, err = stmt.interpolateParams(args); err != nil {
		return
	}

	if statement, err = sqlparser.Parse(q); err != nil {
		return
	}

	if ast, ok = statement.(*sqlparser.Select); !ok {
		panic(ErrSqlNotSupport)
	}

	exr := &selectExecutor{
		RouteInfo: route,
		ctx:       ctx,
		dbc:       stmt.dbc,
		ast:       ast,
	}

	return exr.Run()
}

func (stmt *Stmt) interpolateParams(args []driver.NamedValue) (string, error) {
	// Number of ? should be same to len(args)
	if stmt.paramCount != len(args) {
		return "", driver.ErrSkip
	}

	var (
		buf    bytes.Buffer // TODO: cache buffers
		argPos = 0
	)

	for i := 0; i < len(stmt.query); i++ {
		q := strings.IndexByte(stmt.query[i:], '?')
		if q == -1 {
			buf.WriteString(stmt.query[i:])
			break
		}
		buf.WriteString(stmt.query[i : i+q])
		i += q

		arg := args[argPos].Value
		argPos++

		if arg == nil {
			buf.WriteString("NULL")
			continue
		}

		switch v := arg.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			buf.WriteString(fmt.Sprint(v))
		case bool:
			if v {
				buf.WriteByte('1')
			} else {
				buf.WriteByte('0')
			}
		case string:
			buf.WriteByte('\'')
			escapeStringVal(&buf, v)
			buf.WriteByte('\'')
		default:
			return "", driver.ErrSkip
		}

		if stmt.dbc.MaxAllowedPacket > 0 && buf.Len() > stmt.dbc.MaxAllowedPacket {
			return "", ErrSqlTooLarge
		}
	}
	if argPos != len(args) {
		return "", driver.ErrSkip
	}
	return buf.String(), nil
}
