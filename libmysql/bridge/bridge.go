package bridge

/*
#cgo LDFLAGS: -L/usr/local/Cellar/mysql/5.6.19/lib -lmysqlclient  -lssl -lcrypto
#cgo CFLAGS: -I/usr/local/Cellar/mysql/5.6.19/include/mysql -Os -g -fno-strict-aliasing


#include <mysql.h>
#include <stdlib.h>
*/
import "C"
import "unsafe"

type MySQLHandle *C.struct_st_mysql
type MySQLResHandle *C.struct_st_mysql_res

type MySQLField struct {
	ColumnType byte
	Name       string
}

func MySQLInit() MySQLHandle {
	return C.mysql_init(nil)
}

func MySQLClose(h MySQLHandle) {
	C.mysql_close(h)
}

func GetMySQLError(h MySQLHandle) error {
	if errno := C.mysql_errno(h); errno != 0 {
		err := C.mysql_error(h)
		return &MySQLError{uint16(errno), C.GoString(err)}
	}
	return nil
}

func MySQLRealConnect(h MySQLHandle, host string, port int, user, pass, database string) error {
	args := []*C.char{
		C.CString(host),
		C.CString(user),
		C.CString(pass),
		C.CString(database),
	}

	C.mysql_real_connect(h, args[0], args[1], args[2], args[3], C.uint(port), nil, 0)

	for i, _ := range args {
		C.free(unsafe.Pointer(args[i]))
	}

	return GetMySQLError(h)
}

func MySQLRealEscapeString(h MySQLHandle, val string) string {
	var l C.ulong
	in := C.CString(val)
	defer C.free(unsafe.Pointer(in))

	out := make([]*C.char, (len(val)*2)+3)

	if h != nil {
		l = C.mysql_real_escape_string(h, (*C.char)(unsafe.Pointer(&out[0])), in, C.ulong(len(val)))
	} else {
		l = C.mysql_escape_string((*C.char)(unsafe.Pointer(&out[0])), in, C.ulong(len(val)))
	}

	return C.GoStringN((*C.char)(unsafe.Pointer(&out[0])), C.int(l))
}

func MySQLRealQuery(h MySQLHandle, query string) error {
	q := C.CString(query)
	ql := C.ulong(len(query))
	defer C.free(unsafe.Pointer(q))

	r := C.mysql_real_query(h, q, ql)
	if r != 0 {
		return GetMySQLError(h)
	}

	return nil
}

func MySQLAffectedRows(h MySQLHandle) int64 {
	return int64(C.mysql_affected_rows(h))
}

func MySQLInsertId(h MySQLHandle) int64 {
	return int64(C.mysql_insert_id(h))
}

func MySQLStoreResult(h MySQLHandle) (MySQLResHandle, error) {
	res := C.mysql_store_result(h)
	return res, GetMySQLError(h)
}

func MySQLUseResult(h MySQLHandle) (MySQLResHandle, error) {
	res, err := C.mysql_use_result(h), GetMySQLError(h)
	return res, err
}

// Flushes the result set so the connection is ready for another query
func MySQLFlushResult(h MySQLHandle) error {
	res := C.mysql_store_result(h)
	if err := GetMySQLError(h); res == nil || err != nil {
		return err
	}

	MySQLFreeResult(res)
	return nil
}

// Flushes the rest of the rows on a connection that has had mysql_use_result
// called on it
func MySQLFlushUseResult(h MySQLHandle, r MySQLResHandle) error {
	defer MySQLFreeResult(r)
	for {
		if row := C.mysql_fetch_row(r); row == nil {
			break
		}
	}

	return GetMySQLError(h)
}

func MySQLFreeResult(r MySQLResHandle) {
	C.mysql_free_result(r)
}

func MySQLFetchFields(r MySQLResHandle) []MySQLField {
	num_fields := int(C.mysql_num_fields(r))
	fields := C.mysql_fetch_fields(r)

	// hack!
	// we cast the c array into a pointer to a theoretical super massive go array
	// then we slice it with the extended slice syntax (:length:cap)
	slice := (*[1 << 30]*C.struct_st_mysql_field)(unsafe.Pointer(&fields))[:num_fields:num_fields]

	out := make([]MySQLField, num_fields)
	for i, field := range slice {
		out[i].ColumnType = byte(field._type)
		out[i].Name = C.GoString(field.name)
	}

	return out
}

func MySQLFetchRow(h MySQLHandle, r MySQLResHandle) (*[]*[]byte, error) {
	row := C.mysql_fetch_row(r)
	if row == nil {
		return nil, GetMySQLError(h)
	}

	num_fields := int(C.mysql_num_fields(r))
	row_lengths := C.mysql_fetch_lengths(r)
	if row_lengths == nil {
		return nil, GetMySQLError(h)
	}

	// hack!
	// we cast the c array into a pointer to a theoretical super massive go array
	// then we slice it with the extended slice syntax (:length:cap)
	field_pointers := (*[1 << 30]*C.char)(unsafe.Pointer(row))[:num_fields:num_fields]
	field_lengths := (*[1 << 30]C.ulong)(unsafe.Pointer(row_lengths))[:num_fields:num_fields]

	out := make([]*[]byte, num_fields)

	for i, field := range field_pointers {
		if field == nil {
			out[i] = nil
		} else {
			tmp := C.GoBytes(unsafe.Pointer(field), C.int(field_lengths[i]))
			out[i] = &tmp
		}
	}

	return &out, nil
}
