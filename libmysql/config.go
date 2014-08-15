package libmysql

import "C"

type config struct {
	host     string
	port     int
	user     string
	pass     string
	database string
}
