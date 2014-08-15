package libmysql

import (
	"database/sql/driver"
	"errors"
	"io"

	"github.com/carlsverre/go-libmysql/libmysql/bridge"
)

var (
	rowsClosed = errors.New("Unable to interact with an already closed result object")
)

type streamingResult struct {
	c       *Conn
	columns []bridge.MySQLField
	closed  bool
}

func newStreamingResult(c *Conn) *streamingResult {
	res := new(streamingResult)
	res.c = c
	res.columns = c.bridge.Fields()

	return res
}

func (r *streamingResult) Close() error {
	if !r.closed {
		r.closed = true
		r.c.bridge.Flush()
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

	row, err := r.c.bridge.FetchRow()
	if err != nil {
		return err
	} else if row == nil {
		return io.EOF
	}

	for i, field := range *row {
		if field == nil {
			dest[i] = nil
		} else {
			dest[i] = field
		}
	}

	return err
}
