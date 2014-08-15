package libmysql

import (
	"database/sql/driver"

	"github.com/carlsverre/go-libmysql/libmysql/bridge"
	"github.com/carlsverre/go-libmysql/libmysql/escape"
)

// implements the sql/driver Conn interface
type Conn struct {
	cfg    *config
	handle bridge.MySQLHandle
}

func NewConn(dsn string) (*Conn, error) {
	var err error

	c := new(Conn)

	c.cfg, err = parseDSN(dsn)
	if err != nil {
		return nil, err
	}

	if err = c.open(); err != nil {
		return nil, err
	}

	return c, nil
}

// Open the database connection
func (c *Conn) open() error {
	c.handle = bridge.MySQLInit()

	return bridge.MySQLRealConnect(c.handle,
		c.cfg.host, c.cfg.port,
		c.cfg.user, c.cfg.pass,
		c.cfg.database,
	)
}

// Retrieve the last error raised on the associated connection
func (c *Conn) lastError() error {
	return bridge.GetMySQLError(c.handle)
}

// MemSQL does not support prepared statements at this time
func (c *Conn) Prepare(query string) (driver.Stmt, error) {
	// not implemented
	panic("Prepare is not supported by the mysqldb driver")
}

// MemSQL does not support prepared multi-statement transactions at this time
func (c *Conn) Begin() (driver.Tx, error) {
	panic("Begin is not supported by the mysqldb driver")
}

func (c *Conn) Close() error {
	bridge.MySQLClose(c.handle)
	c.handle = nil
	return nil
}

func (c *Conn) query(query string, args []driver.Value) (err error) {
	query, err = escape.EscapeQuery(c.handle, query, args)
	if err != nil {
		return err
	}

	return bridge.MySQLRealQuery(c.handle, query)
}

// implements the sql/driver Execer interface
func (c *Conn) Exec(query string, args []driver.Value) (res driver.Result, err error) {
	if err = c.query(query, args); err != nil {
		return nil, err
	}

	// flush the connection
	if err = bridge.MySQLFlushResult(c.handle); err != nil {
		return nil, err
	}

	return &execResult{
		rowsAffected: bridge.MySQLAffectedRows(c.handle),
		lastInsertId: bridge.MySQLInsertId(c.handle),
	}, nil
}

// implements the sql/driver Queryer interface
func (c *Conn) Query(query string, args []driver.Value) (res driver.Rows, err error) {
	if err = c.query(query, args); err != nil {
		return nil, err
	}

	return newStreamingResult(c)
}
