package libmysql

import (
	"fmt"

	. "gopkg.in/check.v1"
)

type DSNSuite struct{}

var _ = Suite(&DSNSuite{})

func testCombinations(c *C, user, pass, host string, port int, database string) {
	var err error
	var cfg *config

	authParts := map[string]func(*config){
		fmt.Sprintf("%s:%s@", user, pass): func(cfg *config) {
			c.Assert(cfg.user, Equals, user)
			c.Assert(cfg.pass, Equals, pass)
		},
		fmt.Sprintf("%s@", user): func(cfg *config) {
			c.Assert(cfg.user, Equals, user)
			c.Assert(cfg.pass, Equals, "")
		},
		fmt.Sprintf(""): func(cfg *config) {
			c.Assert(cfg.user, Equals, "")
			c.Assert(cfg.pass, Equals, "")
		},
	}

	connParts := map[string]func(*config){
		fmt.Sprintf("%s:%d", host, port): func(cfg *config) {
			c.Assert(cfg.host, Equals, host)
			c.Assert(cfg.port, Equals, port)
		},
		fmt.Sprintf("%s", host): func(cfg *config) {
			c.Assert(cfg.host, Equals, host)
			c.Assert(cfg.port, Equals, 0)
		},
	}

	dbParts := map[string]func(*config){
		fmt.Sprintf("/%s", database): func(cfg *config) {
			c.Assert(cfg.database, Equals, database)
		},
		fmt.Sprintf(""): func(cfg *config) {
			c.Assert(cfg.database, Equals, "")
		},
	}

	for authPart, authCheck := range authParts {
		for connPart, connCheck := range connParts {
			for dbPart, dbCheck := range dbParts {
				dsn := fmt.Sprintf("%s%s%s", authPart, connPart, dbPart)

				fmt.Printf("Testing dsn: %s\n", dsn)

				cfg, err = parseDSN(dsn)

				c.Assert(err, IsNil)
				authCheck(cfg)
				connCheck(cfg)
				dbCheck(cfg)
			}
		}
	}
}

func (s *DSNSuite) TestBasic(c *C) {
	testCombinations(c, "carl", "test", "host", 3306, "db")
	testCombinations(c, "foo_baz132", "1232", "x.memcompute.com", 2342, "daDJKL123")

	failList := [...]string{
		"user:@asdf:123/db",
		"user:324@asdf:adf/db",
		"user@:adf/db",
		"user@adf/",
		"@adf",
		"@adf/adb",
		"⌘@♞/☎",
	}

	for _, dsn := range failList {
		fmt.Printf("Testing bad dsn: %s\n", dsn)

		_, err := parseDSN(dsn)
		c.Assert(err, Not(IsNil))
	}
}
