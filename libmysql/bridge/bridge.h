#include <mysql.h>

typedef struct m_handle {
	MYSQL			*mysql;
	my_ulonglong	affected_rows;
	my_ulonglong	insert_id;
	unsigned int	num_fields;
	MYSQL_FIELD		*fields;
	MYSQL_RES		*result;
} M_HANDLE;

typedef struct m_row {
	MYSQL_ROW		mysql_row;
	unsigned long	*lengths;
	int				has_error;
} M_ROW;

// Initialize the underlying MySQL library
void m_init();
int m_escape_string(char *out, char *in, unsigned long length);

int m_connect(M_HANDLE *conn, const char *host, unsigned int port, const char *user, const char *pass, const char *database);
void m_close(M_HANDLE *conn);

int m_errno(M_HANDLE *conn);
const char *m_error(M_HANDLE *conn);

/**
 * Send a query to the database.
 *
 * query		the SQL query to send to the server
 * len			the length of the SQL query
 * prep_result	whether or not to prepare a streaming result set
 */
int m_query(M_HANDLE *conn, const char *query, unsigned long len, int prep_result);

void m_flush(M_HANDLE *conn);

M_ROW m_fetch_row(M_HANDLE *conn);
