package namedparameter

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestStatementWrapper_Query(t *testing.T) {
	type args []any

	tests := []struct {
		name      string
		query     string
		args      args
		wantQuery string
		wantArgs  []driver.Value
		wantErr   bool
	}{
		{
			name:      "Very simple case",
			query:     "SELECT id, name, age FROM employees WHERE name LIKE :name",
			args:      args{"name", "%Smith%"},
			wantQuery: "SELECT id, name, age FROM employees WHERE name LIKE \\?",
			wantArgs:  []driver.Value{"%Smith%"},
		},
		{
			name:      "No parameters",
			query:     "SELECT id, name, age FROM employees",
			args:      args{},
			wantQuery: "SELECT id, name, age FROM employees",
			wantArgs:  []driver.Value{},
		},
		{
			name:      "Nil parameters",
			query:     "SELECT id, name, age FROM employees",
			args:      nil,
			wantQuery: "SELECT id, name, age FROM employees",
			wantArgs:  []driver.Value{},
		},
		{
			name:      "Multiple arguments in list mode",
			query:     "SELECT id, name, age FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{"lastname", "Smith", "date", "2020-01-01", "baseSalary", 100000},
			wantQuery: "SELECT id, name, age FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
		},
		{
			name:      "Multiple arguments in list mode - missing argument",
			query:     "SELECT id, name, age FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{"lastname", "Smith", "date", "2020-01-01"},
			wantQuery: "SELECT id, name, age FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
			wantErr:   true,
		},
		{
			name:      "Multiple arguments in map",
			query:     "SELECT id, name, age FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{map[string]any{"lastname": "Smith", "date": "2020-01-01", "baseSalary": 100000}},
			wantQuery: "SELECT id, name, age FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
		},
		{
			name:      "Multiple arguments in map - missing argument",
			query:     "SELECT id, name, age FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{map[string]any{"date": "2020-01-01", "baseSalary": 100000}},
			wantQuery: "SELECT id, name, age FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
			wantErr:   true,
		},
		{
			name:      "Arguments used multiple times - list",
			query:     "SELECT id, name, age FROM employees WHERE (last_name = :lastname AND salary > :baseSalary) OR (last_name = :lastname AND start_time < :date) OR salary < :baseSalary",
			args:      args{"lastname", "Smith", "date", "2020-01-01", "baseSalary", 100000},
			wantQuery: "SELECT id, name, age FROM employees WHERE \\(last_name = \\? AND salary > \\?\\) OR \\(last_name = \\? AND start_time < \\?\\) OR salary < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "Smith", "2020-01-01", 100000},
		},
		{
			name:      "Arguments used multiple times - map",
			query:     "SELECT id, name, age FROM employees WHERE (last_name = :lastname AND salary > :baseSalary) OR (last_name = :lastname AND start_time < :date) OR salary < :baseSalary",
			args:      args{map[string]any{"lastname": "Smith", "date": "2020-01-01", "baseSalary": 100000}},
			wantQuery: "SELECT id, name, age FROM employees WHERE \\(last_name = \\? AND salary > \\?\\) OR \\(last_name = \\? AND start_time < \\?\\) OR salary < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "Smith", "2020-01-01", 100000},
		},
		{
			name:      "Only one parameter - no map",
			query:     "SELECT id, name, age FROM employees WHERE name LIKE :name",
			args:      args{"%Smith%"},
			wantQuery: "SELECT id, name, age FROM employees WHERE name LIKE \\?",
			wantArgs:  []driver.Value{"%Smith%"},
			wantErr:   true,
		},
		{
			name:      "Odd number of parameters",
			query:     "SELECT id, name, age FROM employees WHERE name LIKE :name AND salary > :baseSalary",
			args:      args{"name", "%Smith%", 100000},
			wantQuery: "SELECT id, name, age FROM employees WHERE name LIKE \\? AND salary > \\?",
			wantArgs:  []driver.Value{"%Smith%", 100000},
			wantErr:   true,
		},
	}

	var (
		db    *sql.DB
		mock  sqlmock.Sqlmock
		dbErr error
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, dbErr = sqlmock.New()
			if dbErr != nil {
				t.Errorf("Error initializing Query test error = %v", dbErr)
			}

			mock.ExpectPrepare(tt.wantQuery).ExpectQuery().WithArgs(tt.wantArgs...).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "age"}))

			stmt, prepareErr := Using(db).Prepare(tt.query)
			_, err := stmt.Query(tt.args...)
			dbErr = mock.ExpectationsWereMet()
			if (prepareErr != nil || err != nil || dbErr != nil) != tt.wantErr {
				t.Errorf("Query() error = %v, dbErr = %v, prepareErr = %v, wantErr %v", err, dbErr, prepareErr, tt.wantErr)
				return
			}
		})
	}
}

func TestStatementWrapper_QueryContext(t *testing.T) {
	type args []any

	tests := []struct {
		name      string
		query     string
		args      args
		wantQuery string
		wantArgs  []driver.Value
		wantErr   bool
	}{
		{
			name:      "Very simple case",
			query:     "SELECT id, name, age FROM employees WHERE name LIKE :name",
			args:      args{"name", "%Smith%"},
			wantQuery: "SELECT id, name, age FROM employees WHERE name LIKE \\?",
			wantArgs:  []driver.Value{"%Smith%"},
		},
		{
			name:      "No parameters",
			query:     "SELECT id, name, age FROM employees",
			args:      args{},
			wantQuery: "SELECT id, name, age FROM employees",
			wantArgs:  []driver.Value{},
		},
		{
			name:      "Nil parameters",
			query:     "SELECT id, name, age FROM employees",
			args:      nil,
			wantQuery: "SELECT id, name, age FROM employees",
			wantArgs:  []driver.Value{},
		},
		{
			name:      "Multiple arguments in list mode",
			query:     "SELECT id, name, age FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{"lastname", "Smith", "date", "2020-01-01", "baseSalary", 100000},
			wantQuery: "SELECT id, name, age FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
		},
		{
			name:      "Multiple arguments in list mode - missing argument",
			query:     "SELECT id, name, age FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{"lastname", "Smith", "date", "2020-01-01"},
			wantQuery: "SELECT id, name, age FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
			wantErr:   true,
		},
		{
			name:      "Multiple arguments in map",
			query:     "SELECT id, name, age FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{map[string]any{"lastname": "Smith", "date": "2020-01-01", "baseSalary": 100000}},
			wantQuery: "SELECT id, name, age FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
		},
		{
			name:      "Multiple arguments in map - missing argument",
			query:     "SELECT id, name, age FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{map[string]any{"date": "2020-01-01", "baseSalary": 100000}},
			wantQuery: "SELECT id, name, age FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
			wantErr:   true,
		},
		{
			name:      "Arguments used multiple times - list",
			query:     "SELECT id, name, age FROM employees WHERE (last_name = :lastname AND salary > :baseSalary) OR (last_name = :lastname AND start_time < :date) OR salary < :baseSalary",
			args:      args{"lastname", "Smith", "date", "2020-01-01", "baseSalary", 100000},
			wantQuery: "SELECT id, name, age FROM employees WHERE \\(last_name = \\? AND salary > \\?\\) OR \\(last_name = \\? AND start_time < \\?\\) OR salary < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "Smith", "2020-01-01", 100000},
		},
		{
			name:      "Arguments used multiple times - map",
			query:     "SELECT id, name, age FROM employees WHERE (last_name = :lastname AND salary > :baseSalary) OR (last_name = :lastname AND start_time < :date) OR salary < :baseSalary",
			args:      args{map[string]any{"lastname": "Smith", "date": "2020-01-01", "baseSalary": 100000}},
			wantQuery: "SELECT id, name, age FROM employees WHERE \\(last_name = \\? AND salary > \\?\\) OR \\(last_name = \\? AND start_time < \\?\\) OR salary < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "Smith", "2020-01-01", 100000},
		},
		{
			name:      "Only one parameter - no map",
			query:     "SELECT id, name, age FROM employees WHERE name LIKE :name",
			args:      args{"%Smith%"},
			wantQuery: "SELECT id, name, age FROM employees WHERE name LIKE \\?",
			wantArgs:  []driver.Value{"%Smith%"},
			wantErr:   true,
		},
		{
			name:      "Odd number of parameters",
			query:     "SELECT id, name, age FROM employees WHERE name LIKE :name AND salary > :baseSalary",
			args:      args{"name", "%Smith%", 100000},
			wantQuery: "SELECT id, name, age FROM employees WHERE name LIKE \\? AND salary > \\?",
			wantArgs:  []driver.Value{"%Smith%", 100000},
			wantErr:   true,
		},
	}

	var (
		db    *sql.DB
		mock  sqlmock.Sqlmock
		dbErr error
	)

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, dbErr = sqlmock.New()
			if dbErr != nil {
				t.Errorf("Error initializing Query test error = %v", dbErr)
			}

			mock.ExpectPrepare(tt.wantQuery).ExpectQuery().WithArgs(tt.wantArgs...).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "age"}))

			stmt, prepareErr := Using(db).PrepareContext(ctx, tt.query)
			_, err := stmt.QueryContext(context.Background(), tt.args...)
			dbErr = mock.ExpectationsWereMet()
			if (prepareErr != nil || err != nil || dbErr != nil) != tt.wantErr {
				t.Errorf("QueryContext() error = %v, dbErr = %v, prepareErr = %v, wantErr %v", err, dbErr, prepareErr, tt.wantErr)
				return
			}
		})
	}
}

func TestStatementWrapper_QueryRow(t *testing.T) {
	type args []any

	tests := []struct {
		name      string
		query     string
		args      args
		wantQuery string
		wantArgs  []driver.Value
		wantErr   bool
	}{
		{
			name:      "Very simple case",
			query:     "SELECT id, name, age FROM employees WHERE name LIKE :name",
			args:      args{"name", "%Smith%"},
			wantQuery: "SELECT id, name, age FROM employees WHERE name LIKE \\?",
			wantArgs:  []driver.Value{"%Smith%"},
		},
		{
			name:      "No parameters",
			query:     "SELECT id, name, age FROM employees",
			args:      args{},
			wantQuery: "SELECT id, name, age FROM employees",
			wantArgs:  []driver.Value{},
		},
		{
			name:      "Nil parameters",
			query:     "SELECT id, name, age FROM employees",
			args:      nil,
			wantQuery: "SELECT id, name, age FROM employees",
			wantArgs:  []driver.Value{},
		},
		{
			name:      "Multiple arguments in list mode",
			query:     "SELECT id, name, age FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{"lastname", "Smith", "date", "2020-01-01", "baseSalary", 100000},
			wantQuery: "SELECT id, name, age FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
		},
		{
			name:      "Multiple arguments in list mode - missing argument",
			query:     "SELECT id, name, age FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{"lastname", "Smith", "date", "2020-01-01"},
			wantQuery: "SELECT id, name, age FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
			wantErr:   true,
		},
		{
			name:      "Multiple arguments in map",
			query:     "SELECT id, name, age FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{map[string]any{"lastname": "Smith", "date": "2020-01-01", "baseSalary": 100000}},
			wantQuery: "SELECT id, name, age FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
		},
		{
			name:      "Multiple arguments in map - missing argument",
			query:     "SELECT id, name, age FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{map[string]any{"date": "2020-01-01", "baseSalary": 100000}},
			wantQuery: "SELECT id, name, age FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
			wantErr:   true,
		},
		{
			name:      "Arguments used multiple times - list",
			query:     "SELECT id, name, age FROM employees WHERE (last_name = :lastname AND salary > :baseSalary) OR (last_name = :lastname AND start_time < :date) OR salary < :baseSalary",
			args:      args{"lastname", "Smith", "date", "2020-01-01", "baseSalary", 100000},
			wantQuery: "SELECT id, name, age FROM employees WHERE \\(last_name = \\? AND salary > \\?\\) OR \\(last_name = \\? AND start_time < \\?\\) OR salary < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "Smith", "2020-01-01", 100000},
		},
		{
			name:      "Arguments used multiple times - map",
			query:     "SELECT id, name, age FROM employees WHERE (last_name = :lastname AND salary > :baseSalary) OR (last_name = :lastname AND start_time < :date) OR salary < :baseSalary",
			args:      args{map[string]any{"lastname": "Smith", "date": "2020-01-01", "baseSalary": 100000}},
			wantQuery: "SELECT id, name, age FROM employees WHERE \\(last_name = \\? AND salary > \\?\\) OR \\(last_name = \\? AND start_time < \\?\\) OR salary < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "Smith", "2020-01-01", 100000},
		},
		{
			name:      "Only one parameter - no map",
			query:     "SELECT id, name, age FROM employees WHERE name LIKE :name",
			args:      args{"%Smith%"},
			wantQuery: "SELECT id, name, age FROM employees WHERE name LIKE \\?",
			wantArgs:  []driver.Value{"%Smith%"},
			wantErr:   true,
		},
		{
			name:      "Odd number of parameters",
			query:     "SELECT id, name, age FROM employees WHERE name LIKE :name AND salary > :baseSalary",
			args:      args{"name", "%Smith%", 100000},
			wantQuery: "SELECT id, name, age FROM employees WHERE name LIKE \\? AND salary > \\?",
			wantArgs:  []driver.Value{"%Smith%", 100000},
			wantErr:   true,
		},
	}

	var (
		db    *sql.DB
		mock  sqlmock.Sqlmock
		dbErr error
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, dbErr = sqlmock.New()
			if dbErr != nil {
				t.Errorf("Error initializing Query test error = %v", dbErr)
			}

			mock.ExpectPrepare(tt.wantQuery).ExpectQuery().WithArgs(tt.wantArgs...).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "age"}))

			stmt, prepareErr := Using(db).Prepare(tt.query)
			_, err := stmt.QueryRow(tt.args...)
			dbErr = mock.ExpectationsWereMet()
			if (prepareErr != nil || err != nil || dbErr != nil) != tt.wantErr {
				t.Errorf("QueryRow() error = %v, dbErr = %v, prepareErr = %v, wantErr %v", err, dbErr, prepareErr, tt.wantErr)
				return
			}
		})
	}
}

func TestStatementWrapper_QueryRowContext(t *testing.T) {
	type args []any

	tests := []struct {
		name      string
		query     string
		args      args
		wantQuery string
		wantArgs  []driver.Value
		wantErr   bool
	}{
		{
			name:      "Very simple case",
			query:     "SELECT id, name, age FROM employees WHERE name LIKE :name",
			args:      args{"name", "%Smith%"},
			wantQuery: "SELECT id, name, age FROM employees WHERE name LIKE \\?",
			wantArgs:  []driver.Value{"%Smith%"},
		},
		{
			name:      "No parameters",
			query:     "SELECT id, name, age FROM employees",
			args:      args{},
			wantQuery: "SELECT id, name, age FROM employees",
			wantArgs:  []driver.Value{},
		},
		{
			name:      "Nil parameters",
			query:     "SELECT id, name, age FROM employees",
			args:      nil,
			wantQuery: "SELECT id, name, age FROM employees",
			wantArgs:  []driver.Value{},
		},
		{
			name:      "Multiple arguments in list mode",
			query:     "SELECT id, name, age FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{"lastname", "Smith", "date", "2020-01-01", "baseSalary", 100000},
			wantQuery: "SELECT id, name, age FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
		},
		{
			name:      "Multiple arguments in list mode - missing argument",
			query:     "SELECT id, name, age FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{"lastname", "Smith", "date", "2020-01-01"},
			wantQuery: "SELECT id, name, age FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
			wantErr:   true,
		},
		{
			name:      "Multiple arguments in map",
			query:     "SELECT id, name, age FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{map[string]any{"lastname": "Smith", "date": "2020-01-01", "baseSalary": 100000}},
			wantQuery: "SELECT id, name, age FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
		},
		{
			name:      "Multiple arguments in map - missing argument",
			query:     "SELECT id, name, age FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{map[string]any{"date": "2020-01-01", "baseSalary": 100000}},
			wantQuery: "SELECT id, name, age FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
			wantErr:   true,
		},
		{
			name:      "Arguments used multiple times - list",
			query:     "SELECT id, name, age FROM employees WHERE (last_name = :lastname AND salary > :baseSalary) OR (last_name = :lastname AND start_time < :date) OR salary < :baseSalary",
			args:      args{"lastname", "Smith", "date", "2020-01-01", "baseSalary", 100000},
			wantQuery: "SELECT id, name, age FROM employees WHERE \\(last_name = \\? AND salary > \\?\\) OR \\(last_name = \\? AND start_time < \\?\\) OR salary < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "Smith", "2020-01-01", 100000},
		},
		{
			name:      "Arguments used multiple times - map",
			query:     "SELECT id, name, age FROM employees WHERE (last_name = :lastname AND salary > :baseSalary) OR (last_name = :lastname AND start_time < :date) OR salary < :baseSalary",
			args:      args{map[string]any{"lastname": "Smith", "date": "2020-01-01", "baseSalary": 100000}},
			wantQuery: "SELECT id, name, age FROM employees WHERE \\(last_name = \\? AND salary > \\?\\) OR \\(last_name = \\? AND start_time < \\?\\) OR salary < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "Smith", "2020-01-01", 100000},
		},
		{
			name:      "Only one parameter - no map",
			query:     "SELECT id, name, age FROM employees WHERE name LIKE :name",
			args:      args{"%Smith%"},
			wantQuery: "SELECT id, name, age FROM employees WHERE name LIKE \\?",
			wantArgs:  []driver.Value{"%Smith%"},
			wantErr:   true,
		},
		{
			name:      "Odd number of parameters",
			query:     "SELECT id, name, age FROM employees WHERE name LIKE :name AND salary > :baseSalary",
			args:      args{"name", "%Smith%", 100000},
			wantQuery: "SELECT id, name, age FROM employees WHERE name LIKE \\? AND salary > \\?",
			wantArgs:  []driver.Value{"%Smith%", 100000},
			wantErr:   true,
		},
	}

	var (
		db    *sql.DB
		mock  sqlmock.Sqlmock
		dbErr error
	)

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, dbErr = sqlmock.New()
			if dbErr != nil {
				t.Errorf("Error initializing Query test error = %v", dbErr)
			}

			mock.ExpectPrepare(tt.wantQuery).ExpectQuery().WithArgs(tt.wantArgs...).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "age"}))

			stmt, prepareErr := Using(db).PrepareContext(ctx, tt.query)
			_, err := stmt.QueryRowContext(ctx, tt.args...)
			dbErr = mock.ExpectationsWereMet()
			if (prepareErr != nil || err != nil || dbErr != nil) != tt.wantErr {
				t.Errorf("QueryRowContext() error = %v, dbErr = %v, prepareErr = %v, wantErr %v", err, dbErr, prepareErr, tt.wantErr)
				return
			}
		})
	}
}

func TestStatementWrapper_Exec(t *testing.T) {
	type args []any

	tests := []struct {
		name      string
		query     string
		args      args
		wantQuery string
		wantArgs  []driver.Value
		wantErr   bool
	}{
		{
			name:      "Very simple case",
			query:     "DELETE FROM employees WHERE name LIKE :name",
			args:      args{"name", "%Smith%"},
			wantQuery: "DELETE FROM employees WHERE name LIKE \\?",
			wantArgs:  []driver.Value{"%Smith%"},
		},
		{
			name:      "No parameters",
			query:     "DELETE FROM employees",
			args:      args{},
			wantQuery: "DELETE FROM employees",
			wantArgs:  []driver.Value{},
		},
		{
			name:      "Nil parameters",
			query:     "DELETE FROM employees",
			args:      nil,
			wantQuery: "DELETE FROM employees",
			wantArgs:  []driver.Value{},
		},
		{
			name:      "Multiple arguments in list mode",
			query:     "DELETE FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{"lastname", "Smith", "date", "2020-01-01", "baseSalary", 100000},
			wantQuery: "DELETE FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
		},
		{
			name:      "Multiple arguments in list mode - missing argument",
			query:     "DELETE FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{"lastname", "Smith", "date", "2020-01-01"},
			wantQuery: "DELETE FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
			wantErr:   true,
		},
		{
			name:      "Multiple arguments in map",
			query:     "DELETE FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{map[string]any{"lastname": "Smith", "date": "2020-01-01", "baseSalary": 100000}},
			wantQuery: "DELETE FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
		},
		{
			name:      "Multiple arguments in map - missing argument",
			query:     "DELETE FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{map[string]any{"date": "2020-01-01", "baseSalary": 100000}},
			wantQuery: "DELETE FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
			wantErr:   true,
		},
		{
			name:      "Arguments used multiple times - list",
			query:     "DELETE FROM employees WHERE (last_name = :lastname AND salary > :baseSalary) OR (last_name = :lastname AND start_time < :date) OR salary < :baseSalary",
			args:      args{"lastname", "Smith", "date", "2020-01-01", "baseSalary", 100000},
			wantQuery: "DELETE FROM employees WHERE \\(last_name = \\? AND salary > \\?\\) OR \\(last_name = \\? AND start_time < \\?\\) OR salary < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "Smith", "2020-01-01", 100000},
		},
		{
			name:      "Arguments used multiple times - map",
			query:     "DELETE FROM employees WHERE (last_name = :lastname AND salary > :baseSalary) OR (last_name = :lastname AND start_time < :date) OR salary < :baseSalary",
			args:      args{map[string]any{"lastname": "Smith", "date": "2020-01-01", "baseSalary": 100000}},
			wantQuery: "DELETE FROM employees WHERE \\(last_name = \\? AND salary > \\?\\) OR \\(last_name = \\? AND start_time < \\?\\) OR salary < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "Smith", "2020-01-01", 100000},
		},
		{
			name:      "Only one parameter - no map",
			query:     "DELETE FROM employees WHERE name LIKE :name",
			args:      args{"%Smith%"},
			wantQuery: "DELETE FROM employees WHERE name LIKE \\?",
			wantArgs:  []driver.Value{"%Smith%"},
			wantErr:   true,
		},
		{
			name:      "Odd number of parameters",
			query:     "DELETE FROM employees WHERE name LIKE :name AND salary > :baseSalary",
			args:      args{"name", "%Smith%", 100000},
			wantQuery: "DELETE FROM employees WHERE name LIKE \\? AND salary > \\?",
			wantArgs:  []driver.Value{"%Smith%", 100000},
			wantErr:   true,
		},
	}

	var (
		db    *sql.DB
		mock  sqlmock.Sqlmock
		dbErr error
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, dbErr = sqlmock.New()
			if dbErr != nil {
				t.Errorf("Error initializing Query test error = %v", dbErr)
			}

			mock.ExpectPrepare(tt.wantQuery).ExpectExec().WithArgs(tt.wantArgs...).WillReturnResult(sqlmock.NewResult(0, 1))

			stmt, prepareErr := Using(db).Prepare(tt.query)
			_, err := stmt.Exec(tt.args...)
			dbErr = mock.ExpectationsWereMet()
			if (prepareErr != nil || err != nil || dbErr != nil) != tt.wantErr {
				t.Errorf("Exec() error = %v, dbErr = %v, prepareErr = %v, wantErr %v", err, dbErr, prepareErr, tt.wantErr)
				return
			}
		})
	}
}

func TestStatementWrapper_ExecContext(t *testing.T) {
	type args []any

	tests := []struct {
		name      string
		query     string
		args      args
		wantQuery string
		wantArgs  []driver.Value
		wantErr   bool
	}{
		{
			name:      "Very simple case",
			query:     "DELETE FROM employees WHERE name LIKE :name",
			args:      args{"name", "%Smith%"},
			wantQuery: "DELETE FROM employees WHERE name LIKE \\?",
			wantArgs:  []driver.Value{"%Smith%"},
		},
		{
			name:      "No parameters",
			query:     "DELETE FROM employees",
			args:      args{},
			wantQuery: "DELETE FROM employees",
			wantArgs:  []driver.Value{},
		},
		{
			name:      "Nil parameters",
			query:     "DELETE FROM employees",
			args:      nil,
			wantQuery: "DELETE FROM employees",
			wantArgs:  []driver.Value{},
		},
		{
			name:      "Multiple arguments in list mode",
			query:     "DELETE FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{"lastname", "Smith", "date", "2020-01-01", "baseSalary", 100000},
			wantQuery: "DELETE FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
		},
		{
			name:      "Multiple arguments in list mode - missing argument",
			query:     "DELETE FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{"lastname", "Smith", "date", "2020-01-01"},
			wantQuery: "DELETE FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
			wantErr:   true,
		},
		{
			name:      "Multiple arguments in map",
			query:     "DELETE FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{map[string]any{"lastname": "Smith", "date": "2020-01-01", "baseSalary": 100000}},
			wantQuery: "DELETE FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
		},
		{
			name:      "Multiple arguments in map - missing argument",
			query:     "DELETE FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{map[string]any{"date": "2020-01-01", "baseSalary": 100000}},
			wantQuery: "DELETE FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
			wantErr:   true,
		},
		{
			name:      "Arguments used multiple times - list",
			query:     "DELETE FROM employees WHERE (last_name = :lastname AND salary > :baseSalary) OR (last_name = :lastname AND start_time < :date) OR salary < :baseSalary",
			args:      args{"lastname", "Smith", "date", "2020-01-01", "baseSalary", 100000},
			wantQuery: "DELETE FROM employees WHERE \\(last_name = \\? AND salary > \\?\\) OR \\(last_name = \\? AND start_time < \\?\\) OR salary < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "Smith", "2020-01-01", 100000},
		},
		{
			name:      "Arguments used multiple times - map",
			query:     "DELETE FROM employees WHERE (last_name = :lastname AND salary > :baseSalary) OR (last_name = :lastname AND start_time < :date) OR salary < :baseSalary",
			args:      args{map[string]any{"lastname": "Smith", "date": "2020-01-01", "baseSalary": 100000}},
			wantQuery: "DELETE FROM employees WHERE \\(last_name = \\? AND salary > \\?\\) OR \\(last_name = \\? AND start_time < \\?\\) OR salary < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "Smith", "2020-01-01", 100000},
		},
		{
			name:      "Only one parameter - no map",
			query:     "DELETE FROM employees WHERE name LIKE :name",
			args:      args{"%Smith%"},
			wantQuery: "DELETE FROM employees WHERE name LIKE \\?",
			wantArgs:  []driver.Value{"%Smith%"},
			wantErr:   true,
		},
		{
			name:      "Odd number of parameters",
			query:     "DELETE FROM employees WHERE name LIKE :name AND salary > :baseSalary",
			args:      args{"name", "%Smith%", 100000},
			wantQuery: "DELETE FROM employees WHERE name LIKE \\? AND salary > \\?",
			wantArgs:  []driver.Value{"%Smith%", 100000},
			wantErr:   true,
		},
	}

	var (
		db    *sql.DB
		mock  sqlmock.Sqlmock
		dbErr error
	)

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, dbErr = sqlmock.New()
			if dbErr != nil {
				t.Errorf("Error initializing Query test error = %v", dbErr)
			}

			mock.ExpectPrepare(tt.wantQuery).ExpectExec().WithArgs(tt.wantArgs...).WillReturnResult(sqlmock.NewResult(0, 1))

			stmt, prepareErr := Using(db).PrepareContext(ctx, tt.query)
			_, err := stmt.ExecContext(ctx, tt.args...)
			dbErr = mock.ExpectationsWereMet()
			if (prepareErr != nil || err != nil || dbErr != nil) != tt.wantErr {
				t.Errorf("ExecContext() error = %v, dbErr = %v, prepareErr = %v, wantErr %v", err, dbErr, prepareErr, tt.wantErr)
				return
			}
		})
	}
}

func TestStatementWrapper_QueryForConnection(t *testing.T) {
	type args []any

	tests := []struct {
		name      string
		query     string
		args      args
		wantQuery string
		wantArgs  []driver.Value
		wantErr   bool
	}{
		{
			name:      "Very simple case",
			query:     "SELECT id, name, age FROM employees WHERE name LIKE :name",
			args:      args{"name", "%Smith%"},
			wantQuery: "SELECT id, name, age FROM employees WHERE name LIKE \\?",
			wantArgs:  []driver.Value{"%Smith%"},
		},
		{
			name:      "No parameters",
			query:     "SELECT id, name, age FROM employees",
			args:      args{},
			wantQuery: "SELECT id, name, age FROM employees",
			wantArgs:  []driver.Value{},
		},
		{
			name:      "Nil parameters",
			query:     "SELECT id, name, age FROM employees",
			args:      nil,
			wantQuery: "SELECT id, name, age FROM employees",
			wantArgs:  []driver.Value{},
		},
		{
			name:      "Multiple arguments in list mode",
			query:     "SELECT id, name, age FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{"lastname", "Smith", "date", "2020-01-01", "baseSalary", 100000},
			wantQuery: "SELECT id, name, age FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
		},
		{
			name:      "Multiple arguments in list mode - missing argument",
			query:     "SELECT id, name, age FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{"lastname", "Smith", "date", "2020-01-01"},
			wantQuery: "SELECT id, name, age FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
			wantErr:   true,
		},
		{
			name:      "Multiple arguments in map",
			query:     "SELECT id, name, age FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{map[string]any{"lastname": "Smith", "date": "2020-01-01", "baseSalary": 100000}},
			wantQuery: "SELECT id, name, age FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
		},
		{
			name:      "Multiple arguments in map - missing argument",
			query:     "SELECT id, name, age FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{map[string]any{"date": "2020-01-01", "baseSalary": 100000}},
			wantQuery: "SELECT id, name, age FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
			wantErr:   true,
		},
		{
			name:      "Arguments used multiple times - list",
			query:     "SELECT id, name, age FROM employees WHERE (last_name = :lastname AND salary > :baseSalary) OR (last_name = :lastname AND start_time < :date) OR salary < :baseSalary",
			args:      args{"lastname", "Smith", "date", "2020-01-01", "baseSalary", 100000},
			wantQuery: "SELECT id, name, age FROM employees WHERE \\(last_name = \\? AND salary > \\?\\) OR \\(last_name = \\? AND start_time < \\?\\) OR salary < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "Smith", "2020-01-01", 100000},
		},
		{
			name:      "Arguments used multiple times - map",
			query:     "SELECT id, name, age FROM employees WHERE (last_name = :lastname AND salary > :baseSalary) OR (last_name = :lastname AND start_time < :date) OR salary < :baseSalary",
			args:      args{map[string]any{"lastname": "Smith", "date": "2020-01-01", "baseSalary": 100000}},
			wantQuery: "SELECT id, name, age FROM employees WHERE \\(last_name = \\? AND salary > \\?\\) OR \\(last_name = \\? AND start_time < \\?\\) OR salary < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "Smith", "2020-01-01", 100000},
		},
		{
			name:      "Only one parameter - no map",
			query:     "SELECT id, name, age FROM employees WHERE name LIKE :name",
			args:      args{"%Smith%"},
			wantQuery: "SELECT id, name, age FROM employees WHERE name LIKE \\?",
			wantArgs:  []driver.Value{"%Smith%"},
			wantErr:   true,
		},
		{
			name:      "Odd number of parameters",
			query:     "SELECT id, name, age FROM employees WHERE name LIKE :name AND salary > :baseSalary",
			args:      args{"name", "%Smith%", 100000},
			wantQuery: "SELECT id, name, age FROM employees WHERE name LIKE \\? AND salary > \\?",
			wantArgs:  []driver.Value{"%Smith%", 100000},
			wantErr:   true,
		},
	}

	var (
		db    *sql.DB
		mock  sqlmock.Sqlmock
		dbErr error
	)

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, dbErr = sqlmock.New()
			if dbErr != nil {
				t.Errorf("Error initializing Query test error = %v", dbErr)
			}

			mock.ExpectPrepare(tt.wantQuery).ExpectQuery().WithArgs(tt.wantArgs...).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "age"}))

			conn, _ := db.Conn(ctx)
			stmt, prepareErr := UsingConnection(conn).PrepareContext(ctx, tt.query)
			_, err := stmt.QueryContext(ctx, tt.args...)
			dbErr = mock.ExpectationsWereMet()
			if (prepareErr != nil || err != nil || dbErr != nil) != tt.wantErr {
				t.Errorf("QueryContext() - connection case error = %v, dbErr = %v, prepareErr = %v, wantErr %v", err, dbErr, prepareErr, tt.wantErr)
				return
			}
		})
	}
}

func TestStatementWrapper_ExecForTransaction(t *testing.T) {
	type args []any

	tests := []struct {
		name      string
		query     string
		args      args
		wantQuery string
		wantArgs  []driver.Value
		wantErr   bool
	}{
		{
			name:      "Very simple case",
			query:     "DELETE FROM employees WHERE name LIKE :name",
			args:      args{"name", "%Smith%"},
			wantQuery: "DELETE FROM employees WHERE name LIKE \\?",
			wantArgs:  []driver.Value{"%Smith%"},
		},
		{
			name:      "No parameters",
			query:     "DELETE FROM employees",
			args:      args{},
			wantQuery: "DELETE FROM employees",
			wantArgs:  []driver.Value{},
		},
		{
			name:      "Nil parameters",
			query:     "DELETE FROM employees",
			args:      nil,
			wantQuery: "DELETE FROM employees",
			wantArgs:  []driver.Value{},
		},
		{
			name:      "Multiple arguments in list mode",
			query:     "DELETE FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{"lastname", "Smith", "date", "2020-01-01", "baseSalary", 100000},
			wantQuery: "DELETE FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
		},
		{
			name:      "Multiple arguments in list mode - missing argument",
			query:     "DELETE FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{"lastname", "Smith", "date", "2020-01-01"},
			wantQuery: "DELETE FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
			wantErr:   true,
		},
		{
			name:      "Multiple arguments in map",
			query:     "DELETE FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{map[string]any{"lastname": "Smith", "date": "2020-01-01", "baseSalary": 100000}},
			wantQuery: "DELETE FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
		},
		{
			name:      "Multiple arguments in map - missing argument",
			query:     "DELETE FROM employees WHERE last_name = :lastname AND salary > :baseSalary AND start_time < :date",
			args:      args{map[string]any{"date": "2020-01-01", "baseSalary": 100000}},
			wantQuery: "DELETE FROM employees WHERE last_name = \\? AND salary > \\? AND start_time < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "2020-01-01"},
			wantErr:   true,
		},
		{
			name:      "Arguments used multiple times - list",
			query:     "DELETE FROM employees WHERE (last_name = :lastname AND salary > :baseSalary) OR (last_name = :lastname AND start_time < :date) OR salary < :baseSalary",
			args:      args{"lastname", "Smith", "date", "2020-01-01", "baseSalary", 100000},
			wantQuery: "DELETE FROM employees WHERE \\(last_name = \\? AND salary > \\?\\) OR \\(last_name = \\? AND start_time < \\?\\) OR salary < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "Smith", "2020-01-01", 100000},
		},
		{
			name:      "Arguments used multiple times - map",
			query:     "DELETE FROM employees WHERE (last_name = :lastname AND salary > :baseSalary) OR (last_name = :lastname AND start_time < :date) OR salary < :baseSalary",
			args:      args{map[string]any{"lastname": "Smith", "date": "2020-01-01", "baseSalary": 100000}},
			wantQuery: "DELETE FROM employees WHERE \\(last_name = \\? AND salary > \\?\\) OR \\(last_name = \\? AND start_time < \\?\\) OR salary < \\?",
			wantArgs:  []driver.Value{"Smith", 100000, "Smith", "2020-01-01", 100000},
		},
		{
			name:      "Only one parameter - no map",
			query:     "DELETE FROM employees WHERE name LIKE :name",
			args:      args{"%Smith%"},
			wantQuery: "DELETE FROM employees WHERE name LIKE \\?",
			wantArgs:  []driver.Value{"%Smith%"},
			wantErr:   true,
		},
		{
			name:      "Odd number of parameters",
			query:     "DELETE FROM employees WHERE name LIKE :name AND salary > :baseSalary",
			args:      args{"name", "%Smith%", 100000},
			wantQuery: "DELETE FROM employees WHERE name LIKE \\? AND salary > \\?",
			wantArgs:  []driver.Value{"%Smith%", 100000},
			wantErr:   true,
		},
	}

	var (
		db    *sql.DB
		mock  sqlmock.Sqlmock
		dbErr error
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, dbErr = sqlmock.New()
			if dbErr != nil {
				t.Errorf("Error initializing Query test error = %v", dbErr)
			}

			mock.ExpectBegin()
			mock.ExpectPrepare(tt.wantQuery).ExpectExec().WithArgs(tt.wantArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()

			tx, _ := db.Begin()
			stmt, prepareErr := Using(tx).Prepare(tt.query)
			_, err := stmt.Exec(tt.args...)
			tx.Commit()
			dbErr = mock.ExpectationsWereMet()
			if (prepareErr != nil || err != nil || dbErr != nil) != tt.wantErr {
				t.Errorf("Exec() - case for transaction error = %v, dbErr = %v, prepareErr = %v, wantErr %v", err, dbErr, prepareErr, tt.wantErr)
				return
			}
		})
	}
}

func TestConnectionWrapper_FailPrepare(t *testing.T) {
	t.Run("Failed prepare for DB", func(t *testing.T) {
		db, mock, dbErr := sqlmock.New()
		if dbErr != nil {
			t.Errorf("Error initializing Query test error = %v", dbErr)
		}

		mock.ExpectPrepare("SELECT id, name, age FROM employees WHERE name LIKE \\?").WillReturnError(errors.New("failed to prepare"))

		_, prepareErr := Using(db).Prepare("SELECT id, name, age FROM employees WHERE name LIKE :name")
		if prepareErr == nil {
			t.Errorf("It should have failed to prepare for a DB")
			return
		}
		if dbErr = mock.ExpectationsWereMet(); dbErr != nil {
			t.Errorf("Error preparing query for db, err = %v", dbErr)
		}
	})
	t.Run("Failed prepare for DB with context", func(t *testing.T) {
		db, mock, dbErr := sqlmock.New()
		if dbErr != nil {
			t.Errorf("Error initializing Query test error = %v", dbErr)
		}

		mock.ExpectPrepare("SELECT id, name, age FROM employees WHERE name LIKE \\?").WillReturnError(errors.New("failed to prepare"))

		ctx := context.Background()
		_, prepareErr := Using(db).PrepareContext(ctx, "SELECT id, name, age FROM employees WHERE name LIKE :name")
		if prepareErr == nil {
			t.Errorf("It should have failed to prepare for a DB with context")
			return
		}
		if dbErr = mock.ExpectationsWereMet(); dbErr != nil {
			t.Errorf("Error preparing query for db with context, err = %v", dbErr)
		}
	})
	t.Run("Failed prepare for Tx", func(t *testing.T) {
		db, mock, dbErr := sqlmock.New()
		if dbErr != nil {
			t.Errorf("Error initializing Query test error = %v", dbErr)
		}

		mock.ExpectBegin()
		mock.ExpectPrepare("SELECT id, name, age FROM employees WHERE name LIKE \\?").WillReturnError(errors.New("failed to prepare"))

		tx, _ := db.Begin()
		_, prepareErr := Using(tx).Prepare("SELECT id, name, age FROM employees WHERE name LIKE :name")
		if prepareErr == nil {
			t.Errorf("It should have failed to prepare for a Transaction")
			return
		}
		if dbErr = mock.ExpectationsWereMet(); dbErr != nil {
			t.Errorf("Error preparing query for tx, err = %v", dbErr)
		}
	})
	t.Run("Failed prepare for Connection", func(t *testing.T) {
		db, mock, dbErr := sqlmock.New()
		if dbErr != nil {
			t.Errorf("Error initializing Query test error = %v", dbErr)
		}

		mock.ExpectPrepare("SELECT id, name, age FROM employees WHERE name LIKE \\?").WillReturnError(errors.New("failed to prepare"))

		ctx := context.Background()
		conn, _ := db.Conn(ctx)
		_, prepareErr := UsingConnection(conn).PrepareContext(ctx, "SELECT id, name, age FROM employees WHERE name LIKE :name")
		if prepareErr == nil {
			t.Errorf("It should have failed to prepare for a Connection")
			return
		}
		if dbErr = mock.ExpectationsWereMet(); dbErr != nil {
			t.Errorf("Error preparing query for Connection, err = %v", dbErr)
		}
	})
}
