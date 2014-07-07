package mysqldb

import (
	"database/sql/driver"
	"errors"
	"io"

	"carlsverre.com/mysqldb/bridge"
)

var (
	rowsClosed = errors.New("Unable to interact with an already closed Rows object")
)

type column struct {
	*bridge.MySQLField
}

type streamingResult struct {
	c       *Conn
	r       bridge.MySQLResHandle
	columns []column
	closed  bool
}

func newStreamingResult(c *Conn) (*streamingResult, error) {
	var err error

	res := new(streamingResult)
	res.c = c

	res.r, err = bridge.MySQLUseResult(c.handle)
	if err != nil {
		return nil, err
	}

	fields := bridge.MySQLFetchFields(res.r)
	res.columns = make([]column, len(fields))

	for i, field := range fields {
		res.columns[i] = column{&field}
	}

	return res, nil
}

func (r *streamingResult) Close() error {
	if !r.closed {
		r.closed = true
		return bridge.MySQLFlushUseResult(r.c.handle, r.r)
	}
	return nil
}

func (r *streamingResult) Columns() []string {
	out := make([]string, len(r.columns))
	for i, c := range r.columns {
		out[i] = c.Name
	}
	return out
}

func (r *streamingResult) Next(dest []driver.Value) error {
	if r.closed {
		return rowsClosed
	}

	row, err := bridge.MySQLFetchRow(r.c.handle, r.r)
	if err != nil {
		return err
	} else if row == nil {
		return io.EOF
	}

	for i, field := range *row {
		if field == nil {
			dest[i] = nil
		} else {
			dest[i] = *field
		}
	}

	return err
}
