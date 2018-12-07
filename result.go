package cdbpool

type Result struct {
	rowsAffected int64
	lastInsertId int64
}

func (r *Result) LastInsertId() (int64, error) {
	return r.lastInsertId, nil
}

func (r *Result) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}
