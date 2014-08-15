package escape

import (
	"database/sql/driver"
	"time"

	. "gopkg.in/check.v1"
)

type EscapeSuite struct{}

var _ = Suite(&EscapeSuite{})

type strTuple struct {
	first, second string
}
type timeTuple struct {
	first  time.Time
	second string
}

func escapeAndCompare(c *C, start, end interface{}) {
	out, err := Escape(start)
	c.Assert(err, IsNil)
	c.Check(out, Equals, end)
}

func (s *EscapeSuite) TestBasic(c *C) {
	escapeAndCompare(c, 5, "5")
	escapeAndCompare(c, 123.12312, "123.12312")
	escapeAndCompare(c, true, "true")
	escapeAndCompare(c, false, "false")
	escapeAndCompare(c, nil, "NULL")

	testStrings := []strTuple{
		strTuple{"test", "'test'"},
		strTuple{"1 OR 1=1", "'1 OR 1=1'"},
		strTuple{"\xbf\x27 OR 1=1 /*", "'\xbf\\' OR 1=1 /*'"},
		strTuple{"\" OR 1=1 /*", "'\\\" OR 1=1 /*'"},
		strTuple{"%foobar%", "'%foobar%'"},
	}

	for _, tuple := range testStrings {
		escapeAndCompare(c, tuple.first, tuple.second)
	}

	testTimes := []timeTuple{
		timeTuple{time.Time{}, "'0000-00-00'"},
		timeTuple{time.Date(2014, 7, 6, 5, 2, 32, 123, time.UTC), "'2014-07-06T05:02:32Z'"},
		timeTuple{
			time.Date(2014, 7, 6, 5, 2, 32, 123, time.FixedZone("America/San_Francisco", -28800)),
			"'2014-07-06T05:02:32-08:00'",
		},
	}

	for _, tuple := range testTimes {
		escapeAndCompare(c, tuple.first, tuple.second)
	}
}

func mustEscapeQuery(c *C, expected, query string, args ...driver.Value) {
	out, err := EscapeQuery(query, args)
	c.Assert(err, IsNil)
	c.Check(out, Equals, expected)
}

func (s *EscapeSuite) TestQuery(c *C) {
	mustEscapeQuery(c,
		"SELECT * FROM foo WHERE a = 'test'",
		"SELECT * FROM foo WHERE a = %s", "test",
	)
	mustEscapeQuery(c,
		"SELECT * FROM foo WHERE a = '1 OR 1=1'",
		"SELECT * FROM foo WHERE a = %s", "1 OR 1=1",
	)
	mustEscapeQuery(c,
		"SELECT * FROM foo WHERE a = '\xbf\\' OR 1=1 /*'",
		"SELECT * FROM foo WHERE a = %s", "\xbf\x27 OR 1=1 /*",
	)

	mustEscapeQuery(c,
		"5 123.123 true false NULL 'foo bar' '2014-07-06T05:02:32Z' %",
		"%s %s %s %s %s %s %s %%",
		5, 123.123, true, false, nil, "foo bar", time.Date(2014, 7, 6, 5, 2, 32, 123, time.UTC))

	// no format params
	mustEscapeQuery(c,
		"tj 3ljr3jrlk23jlrj23r jl23jr 2jriajfajf jiwr;jg;rj;r;3rhgahb ai3 airhaw3r",
		"tj 3ljr3jrlk23jlrj23r jl23jr 2jriajfajf jiwr;jg;rj;r;3rhgahb ai3 airhaw3r")

	mustEscapeQuery(c, "", "")

	// no params, but one target
	_, err := EscapeQuery("%s", []driver.Value{})
	c.Assert(err, Not(IsNil))

	// more params than targets
	_, err = EscapeQuery("%s", []driver.Value{1, 2})
	c.Assert(err, Not(IsNil))
}

func (s *EscapeSuite) BenchmarkQueryAllTypes(c *C) {
	t := time.Date(2014, 7, 6, 5, 2, 32, 123, time.UTC)

	for i := 0; i < c.N; i++ {
		EscapeQuery("%s %s %s %s %s %s %s", []driver.Value{5, 123.123, true, false, nil, "foo bar", t})
	}
}

func (s *EscapeSuite) BenchmarkQueryBasicTypes(c *C) {
	for i := 0; i < c.N; i++ {
		EscapeQuery("%s %s %s %s %s", []driver.Value{5, 123.123, true, false, nil})
	}
}

func (s *EscapeSuite) BenchmarkQueryString(c *C) {
	for i := 0; i < c.N; i++ {
		EscapeQuery("%s %s %s %s %s", []driver.Value{"foo", "bar", "baz", "quoox", "whee"})
	}
}

func (s *EscapeSuite) BenchmarkQueryNoParams(c *C) {
	for i := 0; i < c.N; i++ {
		EscapeQuery("foo", []driver.Value{})
	}
}

func (s *EscapeSuite) BenchmarkEscapeString(c *C) {
	for i := 0; i < c.N; i++ {
		Escape("foo")
	}
}

func (s *EscapeSuite) BenchmarkEscapeInt(c *C) {
	for i := 0; i < c.N; i++ {
		Escape(1)
	}
}
