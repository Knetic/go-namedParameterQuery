package namedparameter

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestConnectionWrapper_QueryContext(t *testing.T) {
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

			ctx := context.Background()
			conn, dbErr := db.Conn(ctx)
			if dbErr != nil {
				t.Errorf("Error obtaining connection - error = %v", dbErr)
			}

			mock.ExpectQuery(tt.wantQuery).WithArgs(tt.wantArgs...).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "age"}))

			_, err := UsingConnection(conn).QueryContext(ctx, tt.query, tt.args...)
			dbErr = mock.ExpectationsWereMet()
			if (err != nil || dbErr != nil) != tt.wantErr {
				t.Errorf("QueryContext() error = %v, dbErr = %v, wantErr %v", err, dbErr, tt.wantErr)
				return
			}
		})
	}
}

func TestConnectionWrapper_QueryRowContext(t *testing.T) {
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

			ctx := context.Background()
			conn, dbErr := db.Conn(ctx)
			if dbErr != nil {
				t.Errorf("Error obtaining connection - error = %v", dbErr)
			}

			mock.ExpectQuery(tt.wantQuery).WithArgs(tt.wantArgs...).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "age"}))

			_, err := UsingConnection(conn).QueryRowContext(ctx, tt.query, tt.args...)
			dbErr = mock.ExpectationsWereMet()
			if (err != nil || dbErr != nil) != tt.wantErr {
				t.Errorf("QueryRowContext() error = %v, dbErr = %v, wantErr %v", err, dbErr, tt.wantErr)
				return
			}
		})
	}
}

func TestConnectionWrapper_ExecContext(t *testing.T) {
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

			ctx := context.Background()
			conn, dbErr := db.Conn(ctx)
			if dbErr != nil {
				t.Errorf("Error obtaining connection - error = %v", dbErr)
			}

			mock.ExpectExec(tt.wantQuery).WithArgs(tt.wantArgs...).WillReturnResult(sqlmock.NewResult(0, 1))

			_, err := UsingConnection(conn).ExecContext(context.Background(), tt.query, tt.args...)
			dbErr = mock.ExpectationsWereMet()
			if (err != nil || dbErr != nil) != tt.wantErr {
				t.Errorf("ExecContext() error = %v, dbErr = %v, wantErr %v", err, dbErr, tt.wantErr)
				return
			}
		})
	}
}
