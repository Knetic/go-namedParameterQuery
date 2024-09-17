package namedparameter

import (
	"context"
	"database/sql"
)

// StatementWrapper wraps a sql.Stmt and adds methods that can be used with parameterized queries.
type StatementWrapper struct {
	stmt       *sql.Stmt
	paramQuery *Query
}

// Prepare prepares a statement from a parameterized query and returns a wrapper that can be used with
// named parameters.
func (w *DBObjectWrapper) Prepare(query string) (*StatementWrapper, error) {
	paramQuery := NewQuery(query)
	if stmt, err := w.wrappedDBObject.Prepare(paramQuery.GetParsedQuery()); err != nil {
		return nil, err
	} else {
		return &StatementWrapper{stmt: stmt, paramQuery: paramQuery}, nil
	}
}

// PrepareContext prepares a statement from a parameterized query and returns a wrapper that can be used with
// named parameters.
func (w *DBObjectWrapper) PrepareContext(ctx context.Context, query string) (*StatementWrapper, error) {
	paramQuery := NewQuery(query)
	if stmt, err := w.wrappedDBObject.PrepareContext(ctx, paramQuery.GetParsedQuery()); err != nil {
		return nil, err
	} else {
		return &StatementWrapper{stmt: stmt, paramQuery: paramQuery}, nil
	}
}

// PrepareContext prepares a statement from a parameterized query and returns a wrapper that can be used with
// named parameters.
func (w *ConnectionWrapper) PrepareContext(ctx context.Context, query string) (*StatementWrapper, error) {
	paramQuery := NewQuery(query)
	if stmt, err := w.conn.PrepareContext(ctx, paramQuery.GetParsedQuery()); err != nil {
		return nil, err
	} else {
		return &StatementWrapper{stmt: stmt, paramQuery: paramQuery}, nil
	}
}

// Query performs a parameterized prepared query using the expanded args to feed the parameter values.
// This method expects the parameter arguments to either be a map[string]any, or to come in pairs,
// which are processed as key, value pairs, and in that case, the keys are expected to be strings.
// If the number of arguments is not even, or a key value is not a string, an error is returned.
func (w *StatementWrapper) Query(args ...any) (*sql.Rows, error) {
	mappedParams, err := convertArgsToMap(args)
	if err != nil {
		return nil, err
	}
	if mappedParams == nil {
		return w.stmt.Query()
	}
	w.paramQuery.SetValuesFromMap(mappedParams)
	return w.stmt.Query(w.paramQuery.GetParsedParameters()...)
}

// QueryContext performs a parameterized prepared query using the expanded args to feed the parameter values.
// This method expects the parameter arguments to either be a map[string]any, or to come in pairs,
// which are processed as key, value pairs, and in that case, the keys are expected to be strings.
// If the number of arguments is not even, or a key value is not a string, an error is returned.
func (w *StatementWrapper) QueryContext(ctx context.Context, args ...any) (*sql.Rows, error) {
	mappedParams, err := convertArgsToMap(args)
	if err != nil {
		return nil, err
	}
	if mappedParams == nil {
		return w.stmt.QueryContext(ctx)
	}
	w.paramQuery.SetValuesFromMap(mappedParams)
	return w.stmt.QueryContext(ctx, w.paramQuery.GetParsedParameters()...)
}

// QueryRow performs a parameterized prepared query using the expanded args to feed the parameter values and returns one row.
// This method expects the parameter arguments to either be a map[string]any, or to come in pairs,
// which are processed as key, value pairs, and in that case, the keys are expected to be strings.
// If the number of arguments is not even, or a key value is not a string, an error is returned.
func (w *StatementWrapper) QueryRow(args ...any) (*sql.Row, error) {
	mappedParams, err := convertArgsToMap(args)
	if err != nil {
		return nil, err
	}
	if mappedParams == nil {
		return w.stmt.QueryRow(), nil
	}
	w.paramQuery.SetValuesFromMap(mappedParams)
	return w.stmt.QueryRow(w.paramQuery.GetParsedParameters()...), nil
}

// QueryRowContext performs a parameterized prepared query using the expanded args to feed the parameter values and returns one row.
// This method expects the parameter arguments to either be a map[string]any, or to come in pairs,
// which are processed as key, value pairs, and in that case, the keys are expected to be strings.
// If the number of arguments is not even, or a key value is not a string, an error is returned.
func (w *StatementWrapper) QueryRowContext(ctx context.Context, args ...any) (*sql.Row, error) {
	mappedParams, err := convertArgsToMap(args)
	if err != nil {
		return nil, err
	}
	if mappedParams == nil {
		return w.stmt.QueryRowContext(ctx), nil
	}
	w.paramQuery.SetValuesFromMap(mappedParams)
	return w.stmt.QueryRowContext(ctx, w.paramQuery.GetParsedParameters()...), nil
}

// Exec executes a parameterized prepared sql instruction using the expanded args to feed the parameter values and returns the results.
// This method expects the parameter arguments to either be a map[string]any, or to come in pairs,
// which are processed as key, value pairs, and in that case, the keys are expected to be strings.
// If the number of arguments is not even, or a key value is not a string, an error is returned.
func (w *StatementWrapper) Exec(args ...any) (sql.Result, error) {
	mappedParams, err := convertArgsToMap(args)
	if err != nil {
		return nil, err
	}
	if mappedParams == nil {
		return w.stmt.Exec()
	}
	w.paramQuery.SetValuesFromMap(mappedParams)
	return w.stmt.Exec(w.paramQuery.GetParsedParameters()...)
}

// ExecContext executes a parameterized prepared sql instruction using the expanded args to feed the parameter values and returns the results.
// This method expects the parameter arguments to either be a map[string]any, or to come in pairs,
// which are processed as key, value pairs, and in that case, the keys are expected to be strings.
// If the number of arguments is not even, or a key value is not a string, an error is returned.
func (w *StatementWrapper) ExecContext(ctx context.Context, args ...any) (sql.Result, error) {
	mappedParams, err := convertArgsToMap(args)
	if err != nil {
		return nil, err
	}
	if mappedParams == nil {
		return w.stmt.ExecContext(ctx)
	}
	w.paramQuery.SetValuesFromMap(mappedParams)
	return w.stmt.ExecContext(ctx, w.paramQuery.GetParsedParameters()...)
}
