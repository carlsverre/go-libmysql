#include "bridge.h"
#include <stdio.h>

void m_init() {
	mysql_library_init(0, 0, 0);
}

int m_escape_string(char *out, char *in, unsigned long length) {
	return mysql_escape_string(out, in, length);
}

int m_connect(M_HANDLE *conn, const char *host, unsigned int port, const char *user, const char *pass, const char *database) {
	MYSQL *c;

	mysql_thread_init();

	conn->mysql = mysql_init(0);
	return !mysql_real_connect(conn->mysql, host, user, pass, database, port, 0, 0);
}

void m_close(M_HANDLE *conn) {
	if (conn->mysql) {
		mysql_close(conn->mysql);
		conn->mysql = 0;
	}
}

int m_errno(M_HANDLE *conn) {
	return mysql_errno(conn->mysql);
}

const char *m_error(M_HANDLE *conn) {
	return mysql_error(conn->mysql);
}

void m_clear_result(M_HANDLE *conn) {
	conn->affected_rows = 0;
	conn->insert_id = 0;
	conn->num_fields = 0;
	conn->fields = 0;
	conn->result = 0;
}

int m_query(M_HANDLE *conn, const char *query, unsigned long len, int prep_result) {
	m_clear_result(conn);

	if (mysql_real_query(conn->mysql, query, len) != 0) {
		return 1;
	}

	if (prep_result) {
		conn->result = mysql_use_result(conn->mysql);
	} else {
		conn->result = mysql_store_result(conn->mysql);
		conn->affected_rows = mysql_affected_rows(conn->mysql);
	}

	if (conn->result == 0) {
		// could be an error or an insert stmt
		if (mysql_errno(conn->mysql) != 0) {
			return 1;
		}
		conn->insert_id = mysql_insert_id(conn->mysql);
	} else {
		conn->num_fields = mysql_num_fields(conn->result);
		conn->fields = mysql_fetch_fields(conn->result);
	}

	if (!prep_result) {
		// clear the result set for the next query
		mysql_free_result(conn->result);
		conn->result = 0;
	}

	return 0;
}

void m_flush(M_HANDLE *conn) {
	if (conn->result) {
		while (mysql_fetch_row(conn->result)) { }
		mysql_free_result(conn->result);
	}

	m_clear_result(conn);
}

M_ROW m_fetch_row(M_HANDLE *conn) {
	M_ROW row = {0, 0, 0};
	if (conn->num_fields == 0) {
		return row;
	}

	row.mysql_row = mysql_fetch_row(conn->result);
	if (!row.mysql_row) {
		if (mysql_errno(conn->mysql)) {
			row.has_error = 1;
			return row;
		}
	} else {
		row.lengths = mysql_fetch_lengths(conn->result);
	}

	return row;
}
