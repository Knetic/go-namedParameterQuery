package namedparameter

import (
	"context"
	"database/sql"
)

// WrappableDBObject is an interface that can cover both a sql.DB or a sql.Tx object, making parameterized
// queries usable for any of these without having to implement wrappers for each.
type WrappableDBObject interface {
	Exec(string, ...any) (sql.Result, error)
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	Query(string, ...any) (*sql.Rows, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRow(string, ...any) *sql.Row
	QueryRowContext(context.Context, string, ...any) *sql.Row
	Prepare(string) (*sql.Stmt, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
}

// DBObjectWrapper wraps either a sql.DB or a sql.Tx and adds methods that can be used with parameterized queries.
type DBObjectWrapper struct {
	wrappedDBObject WrappableDBObject
}

// Using wraps a *sql.DB or a *sql.Tx, in order to decorate them with parameterized methods.
func Using(dbObject WrappableDBObject) *DBObjectWrapper {
	return &DBObjectWrapper{wrappedDBObject: dbObject}
}

// Query performs a parameterized query using the expanded args to feed the parameter values.
// This method expects the parameter arguments to either be a map[string]any, or to come in pairs,
// which are processed as key, value pairs, and in that case, the keys are expected to be strings.
// If the number of arguments is not even, or a key value is not a string, an error is returned.
func (w *DBObjectWrapper) Query(query string, args ...any) (*sql.Rows, error) {
	return execute[*sql.Rows](w.wrappedDBObject.Query, query, args...)
}

// QueryContext performs a parameterized query using the expanded args to feed the parameter values.
// This method expects the parameter arguments to either be a map[string]any, or to come in pairs,
// which are processed as key, value pairs, and in that case, the keys are expected to be strings.
// If the number of arguments is not even, or a key value is not a string, an error is returned.
func (w *DBObjectWrapper) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	parsedQuery, params, err := parse(query, args)
	if err != nil {
		return nil, err
	}
	if len(params) == 0 {
		return w.wrappedDBObject.QueryContext(ctx, parsedQuery)
	}
	return w.wrappedDBObject.QueryContext(ctx, parsedQuery, params...)
}

// QueryRow performs a parameterized query using the expanded args to feed the parameter values and returns one row.
// This method expects the parameter arguments to either be a map[string]any, or to come in pairs,
// which are processed as key, value pairs, and in that case, the keys are expected to be strings.
// If the number of arguments is not even, or a key value is not a string, an error is returned.
func (w *DBObjectWrapper) QueryRow(query string, args ...any) (*sql.Row, error) {
	parsedQuery, params, err := parse(query, args)
	if err != nil {
		return nil, err
	}
	if len(params) == 0 {
		return w.wrappedDBObject.QueryRow(parsedQuery), nil
	}
	return w.wrappedDBObject.QueryRow(parsedQuery, params...), nil
}

// QueryRowContext performs a parameterized query using the expanded args to feed the parameter values and returns one row.
// This method expects the parameter arguments to either be a map[string]any, or to come in pairs,
// which are processed as key, value pairs, and in that case, the keys are expected to be strings.
// If the number of arguments is not even, or a key value is not a string, an error is returned.
func (w *DBObjectWrapper) QueryRowContext(ctx context.Context, query string, args ...any) (*sql.Row, error) {
	parsedQuery, params, err := parse(query, args)
	if err != nil {
		return nil, err
	}
	if len(params) == 0 {
		return w.wrappedDBObject.QueryRowContext(ctx, parsedQuery), nil
	}
	return w.wrappedDBObject.QueryRowContext(ctx, parsedQuery, params...), nil
}

// Exec executes a parameterized sql instruction using the expanded args to feed the parameter values and returns the results.
// This method expects the parameter arguments to either be a map[string]any, or to come in pairs,
// which are processed as key, value pairs, and in that case, the keys are expected to be strings.
// If the number of arguments is not even, or a key value is not a string, an error is returned.
func (w *DBObjectWrapper) Exec(query string, args ...any) (sql.Result, error) {
	parsedQuery, params, err := parse(query, args)
	if err != nil {
		return nil, err
	}
	if len(params) == 0 {
		return w.wrappedDBObject.Exec(parsedQuery)
	}
	return w.wrappedDBObject.Exec(parsedQuery, params...)
}

// ExecContext executes a parameterized sql instruction using the expanded args to feed the parameter values and returns the results.
// This method expects the parameter arguments to either be a map[string]any, or to come in pairs,
// which are processed as key, value pairs, and in that case, the keys are expected to be strings.
// If the number of arguments is not even, or a key value is not a string, an error is returned.
func (w *DBObjectWrapper) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	parsedQuery, params, err := parse(query, args)
	if err != nil {
		return nil, err
	}
	if len(params) == 0 {
		return w.wrappedDBObject.ExecContext(ctx, parsedQuery)
	}
	return w.wrappedDBObject.ExecContext(ctx, parsedQuery, params...)
}

type results interface { *sql.Rows | *sql.Row }

type twoValuesFunction[T results] func(string, ...any) (T, error)

type oneValueFunction[T results] func(string, ...any) (T, error)

func execute[T results](f twoValuesFunction[T], query string, args ...any) (T, error) {
	parsedQuery, params, err := parse(query, args)
	if err != nil {
		return nil, err
	}
	if len(params) == 0 {
		return f(parsedQuery)
	}
	return f(parsedQuery, params...)
}
