package namedparameter

import (
	"context"
	"database/sql"
)

// ConnectionWrapper wraps a sql.Conn and adds methods that can be used with parameterized queries.
type ConnectionWrapper struct {
	conn *sql.Conn
}

// UsingConnection wraps a *sql.Conn, in order to decorate it with parameterized methods.
func UsingConnection(conn *sql.Conn) *ConnectionWrapper {
	return &ConnectionWrapper{conn: conn}
}

// QueryContext performs a parameterized query using the expanded args to feed the parameter values.
// This method expects the parameter arguments to either be a map[string]any, or to come in pairs,
// which are processed as key, value pairs, and in that case, the keys are expected to be strings.
// If the number of arguments is not even, or a key value is not a string, an error is returned.
func (w *ConnectionWrapper) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	parsedQuery, params, err := parse(query, args)
	if err != nil {
		return nil, err
	}
	if len(params) == 0 {
		return w.conn.QueryContext(ctx, parsedQuery)
	}
	return w.conn.QueryContext(ctx, parsedQuery, params...)
}

// QueryRowContext performs a parameterized query using the expanded args to feed the parameter values and returns one row.
// This method expects the parameter arguments to either be a map[string]any, or to come in pairs,
// which are processed as key, value pairs, and in that case, the keys are expected to be strings.
// If the number of arguments is not even, or a key value is not a string, an error is returned.
func (w *ConnectionWrapper) QueryRowContext(ctx context.Context, query string, args ...any) (*sql.Row, error) {
	parsedQuery, params, err := parse(query, args)
	if err != nil {
		return nil, err
	}
	if len(params) == 0 {
		return w.conn.QueryRowContext(ctx, parsedQuery), nil
	}
	return w.conn.QueryRowContext(ctx, parsedQuery, params...), nil
}

// ExecContext executes a parameterized sql instruction using the expanded args to feed the parameter values and returns the results.
// This method expects the parameter arguments to either be a map[string]any, or to come in pairs,
// which are processed as key, value pairs, and in that case, the keys are expected to be strings.
// If the number of arguments is not even, or a key value is not a string, an error is returned.
func (w *ConnectionWrapper) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	parsedQuery, params, err := parse(query, args)
	if err != nil {
		return nil, err
	}
	if len(params) == 0 {
		return w.conn.ExecContext(ctx, parsedQuery)
	}
	return w.conn.ExecContext(ctx, parsedQuery, params...)
}
