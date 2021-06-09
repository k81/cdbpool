package main

import (
	"context"
	"fmt"

	"github.com/std0d9k81/cdbpool"
)

var (
	mctx = context.Background()
)

func main() {
	cluster := cdbpool.NewCluster("tcp(192.168.0.132:9123,192.168.0.133:9123)/orders?timeout=30s&readTimeout=4s&writeTimeout=15s")

	var (
		orderId  string
		goods    string
		bigId    = uint64(391165030316115969)
		count    = 50
		useTotal = false
	)

	rows, err := cluster.GetDB(mctx, "orders", bigId, useTotal).Query("select id, goods from order_info where 1=1 limit ?", count)
	if err != nil {
		fmt.Printf("db.Query() error: %v\n", err)
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
