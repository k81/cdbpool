package cdbpool

import (
	"database/sql/driver"
	"io"
)

type Rows struct {
	records []*StoreRecord
	columns []string
	next    int
}

// Columns returns the names of the columns. The number of
// columns of the result is inferred from the length of the
// slice. If a particular column name isn't known, an empty
// string should be returned for that entry.
func (rows *Rows) Columns() []string {
	if len(rows.records) > 0 {
		if len(rows.columns) > 0 {
			return rows.columns
		}

		r := rows.records[0]
		rows.columns = make([]string, len(r.Units))
		for i, kv := range r.Units {
			rows.columns[i] = kv.Key
		}
		return rows.columns
	}

	return []string{}
}

// Close closes the rows iterator.
func (rows *Rows) Close() error {
	return nil
}

// Next is called to populate the next row of data into
// the provided slice. The provided slice will be the same
// size as the Columns() are wide.
//
// Next should return io.EOF when there are no more rows.
func (rows *Rows) Next(dest []driver.Value) (err error) {
	if rows.next >= len(rows.records) {
		return io.EOF
	}

	r := rows.records[rows.next]
	rows.next++

	for i, kv := range r.Units {
		dest[i] = kv.Value
	}
	return nil
}
