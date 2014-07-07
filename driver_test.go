package mysqldb

import (
	"database/sql"
	"fmt"

	. "gopkg.in/check.v1"
)

type DriverSuite struct {
	dsn string
	db  *sql.DB
}

var _ = Suite(&DriverSuite{})

func (s *DriverSuite) SetUpSuite(c *C) {
	var err error

	s.dsn = "root@localhost"

	s.db, err = sql.Open("mysql", s.dsn)
	c.Assert(err, IsNil)

	_, err = s.db.Exec("SELECT 1")
	c.Assert(err, IsNil)
}

func (s *DriverSuite) SetUpTest(c *C) {
	s.mustExec(c, "DROP DATABASE IF EXISTS gotests")
	s.mustExec(c, "CREATE DATABASE gotests")

	s.mustExec(c, "USE gotests")
	s.mustExec(c, "CREATE TABLE x (id bigint auto_increment primary key, foo varchar(255))")
}

func (s *DriverSuite) TearDownTest(c *C) {
	s.mustExec(c, "DROP DATABASE gotests")
}

func (s *DriverSuite) TearDownSuite(c *C) {
	err := s.db.Close()
	c.Assert(err, IsNil)
}

func (s *DriverSuite) mustExec(c *C, query string, args ...interface{}) sql.Result {
	res, err := s.db.Exec(query, args...)
	c.Assert(err, IsNil)
	return res
}

func (s *DriverSuite) mustQuery(c *C, query string, args ...interface{}) *sql.Rows {
	res, err := s.db.Query(query, args...)
	c.Assert(err, IsNil)
	return res
}

func (s *DriverSuite) TestBasic(c *C) {
	rows := s.mustQuery(c, "SHOW DATABASES")

	for rows.Next() {
		var name string

		err := rows.Scan(&name)
		c.Assert(err, IsNil)

		fmt.Printf("db_name: %s\n", name)
	}
}

func (s *DriverSuite) BenchmarkBasic(c *C) {
	query := "INSERT INTO x (foo) VALUES (%s)"
	for i := 0; i < c.N; i++ {
		s.db.Exec(query, "asdf")
	}
}
