package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/std0d9k81/kate/orm"
)

type OrderInfo struct {
	Id    string `orm:"pk"`
	Goods string
}

func main() {
	var (
		driverName      = "cdbpool"
		dataSource      = "tcp(192.168.0.132:9123,192.168.0.133:9123)/orders?timeout=30s&readTimeout=4s&writeTimeout=15s"
		maxIdle         = 20
		maxOpen         = 100
		connMaxLifetime = 30 * time.Minute
	)
	orm.RegisterDB("default", driverName, dataSource, maxIdle, maxOpen, connMaxLifetime)
	orm.RegisterModel("orders", new(OrderInfo))

	var (
		orders []*OrderInfo
		limit  = 50
	)

	err := orm.NewOrm(context.TODO()).QueryTable(new(OrderInfo)).Filter("id__gt", 0).Limit(limit).All(&orders)
	if err != nil {
		fmt.Printf("query table failed: %v\n", err)
		os.Exit(-1)
	}

	for _, order := range orders {
		fmt.Printf("order_id=%v, goods=%v\n", order.Id, order.Goods)
	}
}
