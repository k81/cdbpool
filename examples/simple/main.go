package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/std0d9k81/cdbpool"
)

var (
	mctx = context.Background()
)

func main() {
	db, err := sql.Open("cdbpool", "tcp(192.168.0.132:9123)/orders?timeout=30s&readTimeout=4s&writeTimeout=15s")
	if err != nil {
		fmt.Printf("sql.Open() error: %v\n", err)
		return
	}

	var (
		orderId string
		goods   string
		rows    *sql.Rows
		bigId   = uint64(391165030316115969)
		limit   = 50
		ctx     = cdbpool.SetRoute(mctx, "orders", bigId, false)
	)

	rows, err = db.QueryContext(ctx, "select id, goods from order_info where 1=1 limit ?", limit)
	if err != nil {
		fmt.Printf("db.QueryRowContext() error: %v\n", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(&orderId, &goods); err != nil {
			fmt.Printf("rows.Scan() error: %v\n", err)
			return
		}

		fmt.Printf("order_id=%v, goods=%v\n", orderId, goods)
	}
}
