package cdbpool

import "testing"

func TestParseDSN(t *testing.T) {
	dsn := "tcp(127.0.0.1:9123)/users?timeout=10s&readTimeout=30s&writeTimeout=60s&maxAllowedPacket=1024"
	cfg, err := ParseDSN(dsn)
	if err != nil {
		t.Errorf("parse error:%v", err)
	}

	t.Logf("dsn.Net: %v", cfg.Net)
	t.Logf("dsn.Addr: %v", cfg.Addr)
	t.Logf("dsn.DBName: %v", cfg.DBName)
	t.Logf("dsn.MaxAllowedPacket: %v", cfg.MaxAllowedPacket)
	t.Logf("dsn.Timeout: %v", cfg.Timeout)
	t.Logf("dsn.ReadTimeout: %v", cfg.ReadTimeout)
	t.Logf("dsn.WriteTimeout: %v", cfg.WriteTimeout)
}
