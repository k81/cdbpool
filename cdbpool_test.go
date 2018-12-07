package cdbpool

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/k81/kate/bigid"
)

var (
	db  *sql.DB
	err error
)

func TestMysqlErrCode(t *testing.T) {
	var (
		bigId = bigid.New(2)
		ctx   = SetRoute(context.Background(), "test", bigId, false)
	)

	if _, err = db.ExecContext(ctx, "delete from test where 1=1"); err != nil {
		t.Fatalf("clean db, err=%v", err)
	}

	if _, err = db.ExecContext(ctx, "insert into test(id, value) values(?, ?)", bigId, fmt.Sprint(bigId)); err != nil {
		t.Fatalf("first insert: id=%v, err=%v", bigId, err)
	}
	if _, err = db.ExecContext(ctx, "insert into test(id, value) values(?, ?)", bigId, fmt.Sprint(bigId)); err != nil {
		if dberr, ok := err.(*DBError); ok {
			if 1062 == dberr.GetMysqlErrno() {
				t.Log("ignore duplicated key")
				return
			}
		}
		t.Fatalf("first insert: id=%v, err=%v", bigId, err)
	}
}

func TestMain(m *testing.M) {
	db, err = sql.Open("cdbpool", "tcp(127.0.0.1:9123)/test?timeout=30s&readTimeout=4s&writeTimeout=15s")
	if err != nil {
		fmt.Printf("sql.Open() error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
