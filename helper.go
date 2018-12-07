package cdbpool

import (
	"bytes"
	"database/sql/driver"
	"fmt"

	"github.com/k81/sqlparser"
)

func driverArgs(args []driver.Value) []driver.NamedValue {
	nargs := make([]driver.NamedValue, len(args))
	for i, v := range args {
		nargs[i] = driver.NamedValue{
			Ordinal: i + 1,
			Value:   v,
		}
	}
	return nargs
}

func escapeStringVal(buf *bytes.Buffer, v string) {
	for i := 0; i < len(v); i++ {
		c := v[i]
		switch c {
		case '\x00':
			buf.WriteByte('\\')
			buf.WriteByte('0')
		case '\n':
			buf.WriteByte('\\')
			buf.WriteByte('n')
		case '\r':
			buf.WriteByte('\\')
			buf.WriteByte('r')
		case '\x1a':
			buf.WriteByte('\\')
			buf.WriteByte('Z')
		case '\'':
			buf.WriteByte('\\')
			buf.WriteByte('\'')
		case '"':
			buf.WriteByte('\\')
			buf.WriteByte('"')
		case '\\':
			buf.WriteByte('\\')
			buf.WriteByte('\\')
		default:
			buf.WriteByte(c)
		}
	}
}

func astValue(v sqlparser.SQLNode) string {
	buf := GetBuffer()
	defer PutBuffer(buf)

	v.Format(buf)
	return buf.String()
}

func ErrSqlInvalid(detail string) error {
	return fmt.Errorf("sql invalid: %s", detail)
}
