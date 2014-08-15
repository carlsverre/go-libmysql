package libmysql

// implements db/sql Result
type execResult struct {
	rowsAffected int64
	lastInsertId int64
}

func (res *execResult) LastInsertId() (int64, error) {
	return res.lastInsertId, nil
}

func (res *execResult) RowsAffected() (int64, error) {
	return res.rowsAffected, nil
}
