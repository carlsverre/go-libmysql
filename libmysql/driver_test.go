package libmysql

import (
	"database/sql"

	. "gopkg.in/check.v1"
)

type DriverSuite struct {
	dsn string
	db  *sql.DB
}

var _ = Suite(&DriverSuite{})

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

func (s *DriverSuite) SetUpSuite(c *C) {
	var err error

	s.dsn = "root@127.0.0.1:3306"

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

func (s *DriverSuite) TestBasic(c *C) {
	expectedString := "look a ♞, oh wow"

	for i := 0; i < 10; i++ {
		_, err := s.db.Exec("INSERT INTO x (id, foo) VALUES (%s, %s)", i+1, expectedString)
		c.Assert(err, IsNil)
	}

	rows := s.mustQuery(c, "SELECT * FROM x ORDER BY id ASC")
	count := 0

	for rows.Next() {
		var id int
		var name string

		err := rows.Scan(&id, &name)
		c.Assert(err, IsNil)

		count += 1
		c.Assert(id, Equals, count)
		c.Assert(name, Equals, expectedString)
	}

	c.Assert(count, Equals, 10)
}

func (s *DriverSuite) BenchmarkBasicInsertWithEscape(c *C) {
	query := "INSERT INTO x (foo) VALUES (%s)"
	for i := 0; i < c.N; i++ {
		s.db.Exec(query, "asdf")
	}
}

func (s *DriverSuite) BenchmarkBasicInsertNoEscape(c *C) {
	query := "INSERT INTO x (foo) VALUES ('asdf')"
	s.db.Exec(query)
	for i := 0; i < c.N; i++ {
		s.db.Exec(query)
	}
}

func (s *DriverSuite) BenchmarkBasicKVSelectWithEscape(c *C) {
	var id int
	var name string

	s.db.Exec("INSERT INTO x (foo) VALUES ('asdf')")

	for i := 0; i < c.N; i++ {
		rows := s.db.QueryRow("SELECT * FROM x WHERE id = %s", 1)
		err := rows.Scan(&id, &name)
		c.Assert(err, IsNil)
	}
}

func (s *DriverSuite) BenchmarkBasicKVSelectNoEscape(c *C) {
	var id int
	var name string

	s.db.Exec("INSERT INTO x (foo) VALUES ('asdf')")

	for i := 0; i < c.N; i++ {
		rows := s.db.QueryRow("SELECT * FROM x WHERE id = 1")
		err := rows.Scan(&id, &name)
		c.Assert(err, IsNil)
	}
}
