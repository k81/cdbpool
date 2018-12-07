package cdbpool

import "context"

type Tx struct {
	*RouteInfo
	ctx context.Context
	dbc *Conn
}

func newTx(ctx context.Context, dbc *Conn, route *RouteInfo) *Tx {
	dbc.ctx = ctx

	tx := &Tx{
		RouteInfo: route,
		ctx:       ctx,
		dbc:       dbc,
	}

	return tx
}

func (tx *Tx) Commit() error {
	defer func() {
		tx.dbc.ctx = nil
	}()

	exr := &commitExecutor{tx}
	return exr.Run()
}

func (tx *Tx) Rollback() error {
	defer func() {
		tx.dbc.ctx = nil
	}()

	exr := &rollbackExecutor{tx}
	return exr.Run()
}
