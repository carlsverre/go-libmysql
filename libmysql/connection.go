package libmysql

import (
	"database/sql/driver"

	"github.com/carlsverre/go-libmysql/libmysql/bridge"
	"github.com/carlsverre/go-libmysql/libmysql/escape"
)

// implements the sql/driver Conn interface
type Conn struct {
	cfg    *config
	bridge *bridge.Bridge
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
	var err error

	c.bridge, err = bridge.NewBridge(
		c.cfg.host, c.cfg.port,
		c.cfg.user, c.cfg.pass,
		c.cfg.database,
	)

	return err
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
	c.bridge.Close()
	c.bridge = nil
	return nil
}

// implements the sql/driver Execer interface
func (c *Conn) Exec(query string, args []driver.Value) (res driver.Result, err error) {
	query, err = escape.EscapeQuery(query, args)
	if err != nil {
		return nil, err
	}

	if err = c.bridge.Execute(query); err != nil {
		return nil, err
	}

	return &execResult{
		rowsAffected: c.bridge.RowsAffected(),
		lastInsertId: c.bridge.LastInsertID(),
	}, nil
}

// implements the sql/driver Queryer interface
func (c *Conn) Query(query string, args []driver.Value) (res driver.Rows, err error) {
	query, err = escape.EscapeQuery(query, args)
	if err != nil {
		return nil, err
	}

	if err = c.bridge.Query(query); err != nil {
		return nil, err
	}

	return newStreamingResult(c), nil
}
