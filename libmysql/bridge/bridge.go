package bridge

/*
#cgo LDFLAGS: -L/usr/lib/x86_64-linux-gnu -lmysqlclient_r -lpthread -lz -lm -ldl
#cgo CFLAGS: -I/usr/include/mysql -DBIG_JOINS=1 -fno-strict-aliasing -g -DNDEBUG -ggdb -fPIC -Werror=implicit

#include <stdlib.h>
#include "bridge.h"
*/
import "C"
import "unsafe"

const (
	maxSize = 1 << 20
)

type Bridge struct {
	h C.M_HANDLE
}

type MySQLField struct {
	ColumnType byte
	Name       string
}

func init() {
	// bootstrap the mysql library at import time
	C.m_init()
}

func EscapeString(val string) string {
	in := C.CString(val)
	defer C.free(unsafe.Pointer(in))

	out := make([]*C.char, (len(val)*2)+3)
	cOut := (*C.char)(unsafe.Pointer(&out[0]))

	l := C.m_escape_string(cOut, in, C.ulong(len(val)))
	return C.GoStringN(cOut, C.int(l))
}

func NewBridge(host string, port int, user, pass, database string) (*Bridge, error) {
	bridge := new(Bridge)

	cHost := C.CString(host)
	defer C.free(unsafe.Pointer(cHost))

	cPort := C.uint(port)

	cUser := C.CString(user)
	defer C.free(unsafe.Pointer(cUser))

	cPass := C.CString(pass)
	defer C.free(unsafe.Pointer(cPass))

	cDatabase := C.CString(database)
	defer C.free(unsafe.Pointer(cDatabase))

	if C.m_connect(&bridge.h, cHost, cPort, cUser, cPass, cDatabase) != 0 {
		defer bridge.Close()
		return nil, bridge.lastError()
	}

	return bridge, nil
}

func (b *Bridge) lastError() error {
	if errno := C.m_errno(&b.h); errno != 0 {
		err := C.m_error(&b.h)
		return &MySQLError{uint16(errno), C.GoString(err)}
	}
	return nil
}

func (b *Bridge) query(query string, prepResult int) error {
	q := C.CString(query)
	defer C.free(unsafe.Pointer(q))

	if C.m_query(&b.h, q, C.ulong(len(query)), C.int(prepResult)) != 0 {
		return b.lastError()
	}

	return nil
}

func (b *Bridge) Close() {
	C.m_close(&b.h)
}

func (b *Bridge) IsClosed() bool {
	return b.h.mysql == nil
}

func (b *Bridge) Query(query string) error {
	return b.query(query, 1)
}

func (b *Bridge) Execute(query string) error {
	return b.query(query, 0)
}

func (b *Bridge) Flush() {
	C.m_flush(&b.h)
}

func (b *Bridge) Fields() []MySQLField {
	nFields := int(b.h.num_fields)
	if nFields == 0 {
		return nil
	}

	cFields := (*[maxSize]C.MYSQL_FIELD)(unsafe.Pointer(b.h.fields))

	fields := make([]MySQLField, nFields)
	for i := 0; i < nFields; i++ {
		fields[i].Name = C.GoStringN(cFields[i].name, C.int(cFields[i].name_length))
		fields[i].ColumnType = byte(cFields[i]._type)
	}

	return fields
}

func (b *Bridge) FetchRow() (*[][]byte, error) {
	mRow := C.m_fetch_row(&b.h)
	if mRow.has_error != 0 {
		return nil, b.lastError()
	}

	rowPtr := (*[maxSize]*[maxSize]byte)(unsafe.Pointer(mRow.mysql_row))
	if rowPtr == nil {
		return nil, nil
	}

	nFields := int(b.h.num_fields)
	cLengths := (*[maxSize]uint64)(unsafe.Pointer(mRow.lengths))

	totalLength := uint64(0)
	for i := 0; i < nFields; i++ {
		totalLength += cLengths[i]
	}

	arena := make([]byte, 0, int(totalLength))
	row := make([][]byte, nFields)

	for i := 0; i < nFields; i++ {
		fieldLength := cLengths[i]
		fieldPtr := rowPtr[i]
		if fieldPtr == nil {
			continue
		}

		start := len(arena)
		arena = append(arena, fieldPtr[:fieldLength]...)
		row[i] = arena[start : start+int(fieldLength)]
	}

	return &row, nil
}

func (b *Bridge) RowsAffected() int64 {
	return int64(b.h.affected_rows)
}

func (b *Bridge) LastInsertID() int64 {
	return int64(b.h.insert_id)
}
