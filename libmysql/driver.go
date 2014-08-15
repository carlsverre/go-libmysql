package libmysql

import (
	"database/sql"
	"database/sql/driver"
)

// MySQLDBDriver implements the database/sql Driver interface
type MySQLDBDriver struct{}

// Opens a new connection to the MySQL server specified by the provided DSN
func (d *MySQLDBDriver) Open(dsn string) (driver.Conn, error) {
	c, err := NewConn(dsn)
	return c, err
}

func init() {
	sql.Register("libmysql", &MySQLDBDriver{})
}
