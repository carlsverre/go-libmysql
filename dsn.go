package mysqldb

import (
	"errors"
	"regexp"
	"strconv"
)

var (
	rDSN = regexp.MustCompile(`^(?:(?P<user>[[:word:]-.]+?)(?::(?P<pass>[[:word:]-.]+?))?@)?(?P<host>[[:word:]-.]+?)(?::(?P<port>[[:word:]-.]+?))?(?:\/(?P<database>[[:word:]-.]+?))?$`)

	errInvalidDSN  = errors.New("Failed to parse DSN")
	errInvalidPort = errors.New("Failed to parse valid port number from DSN")
)

// parse the provided dsn into a new config object
// currently must include all fields:
// user:password@host:port/database
func parseDSN(dsn string) (*config, error) {
	cfg := &config{}
	var err error
	var port uint64

	match := rDSN.FindStringSubmatch(dsn)
	if match == nil {
		return nil, errInvalidDSN
	}

	for i, name := range rDSN.SubexpNames() {
		switch name {
		case "user":
			cfg.user = match[i]
		case "pass":
			cfg.pass = match[i]
		case "host":
			cfg.host = match[i]
		case "port":
			if match[i] != "" {
				port, err = strconv.ParseUint(match[i], 10, 0)
				if err != nil {
					return nil, errInvalidPort
				}
				cfg.port = int(port)
			}
		case "database":
			cfg.database = match[i]
		default:
			continue
		}
	}

	return cfg, nil
}
