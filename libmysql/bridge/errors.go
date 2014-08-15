package bridge

import (
	"fmt"
)

type MySQLError struct {
	Errno   uint16
	Message string
}

func (err *MySQLError) Error() string {
	return fmt.Sprintf("Error %d: %s", err.Errno, err.Message)
}
