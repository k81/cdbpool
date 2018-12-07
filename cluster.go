package cdbpool

import (
	"context"
	"database/sql"
	"math/rand"
	"strings"
	"time"

	"github.com/k81/kate/log"
)

type Cluster struct {
	addrs []string
	pools []*sql.DB
}

func NewCluster(dsn string) *Cluster {
	conf, err := ParseDSN(dsn)
	if err != nil {
		log.Fatal(mctx, "invalid dsn", "dsn", dsn, "error", err)
	}

	var (
		hosts = strings.Split(conf.Addr, ",")
		addrs = make([]string, 0, len(hosts))
	)

	for _, host := range hosts {
		conf.Addr = strings.TrimSpace(host)
		if len(conf.Addr) > 0 {
			addrs = append(addrs, conf.FormatDSN())
		}
	}

	cluster := &Cluster{
		addrs: make([]string, len(addrs)),
		pools: make([]*sql.DB, len(addrs)),
	}
	copy(cluster.addrs, addrs)

	for idx, addr := range cluster.addrs {
		db, err := sql.Open("cdbpool", addr)
		if err != nil {
			log.Fatal(mctx, "failed to open database", "id", idx, "addr", addr, "error", err)
		}

		cluster.pools[idx] = db
	}
	return cluster
}

func (c *Cluster) Close() (err error) {
	for _, pool := range c.pools {
		pool.Close()
	}
	return
}

func (c *Cluster) SetMaxIdleConns(n int) {
	for _, pool := range c.pools {
		pool.SetMaxIdleConns(n)
	}
}

func (c *Cluster) SetMaxOpenConns(n int) {
	for _, pool := range c.pools {
		pool.SetMaxOpenConns(n)
	}
}

func (c *Cluster) SetConnMaxLifetime(d time.Duration) {
	for _, pool := range c.pools {
		pool.SetConnMaxLifetime(d)
	}
}

func (c *Cluster) GetDB(ctx context.Context, dbName string, bigId uint64, offline bool) *DB {
	db := &DB{
		ctx: SetRoute(ctx, dbName, bigId, offline),
		DB:  c.pools[rand.Intn(len(c.pools))],
	}
	return db
}
