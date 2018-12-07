package cdbpool

import "context"

const (
	keyRoute = "__db_route__"
)

type RouteInfo struct {
	DBName  string
	BigId   uint64
	Offline bool
}

func SetRoute(ctx context.Context, dbName string, bigId uint64, offline bool) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	route := &RouteInfo{
		DBName:  dbName,
		BigId:   bigId,
		Offline: offline,
	}
	return context.WithValue(ctx, keyRoute, route)
}

func GetRoute(ctx context.Context) *RouteInfo {
	if route, ok := ctx.Value(keyRoute).(*RouteInfo); ok {
		return route
	}
	return nil
}
