package namedparameter

import (
	"reflect"
	"testing"
)

func Test_convertArgsToMap(t *testing.T) {
	type args []any
	tests := []struct {
		name    string
		args    args
		want    map[string]any
		wantErr bool
	}{
		{
			name: "Simple map case",
			args: args{map[string]any{"id": 1, "name": "John", "job": nil}},
			want: map[string]any{"id": 1, "name": "John", "job": nil},
		},
		{
			name: "Simple list of params",
			args: args{"id", 1, "name", "John", "job", nil},
			want: map[string]any{"id": 1, "name": "John", "job": nil},
		},
		{
			name: "Simple list of params",
			args: args{"id", 1, "name", "John", "job", nil},
			want: map[string]any{"id": 1, "name": "John", "job": nil},
		},
		{
			name: "nil args",
			args: nil,
			want: nil,
		},
		{
			name: "Empty args",
			args: args{},
			want: nil,
		},
		{
			name:    "One argument, not a map",
			args:    args{"id"},
			wantErr: true,
		},
		{
			name:    "Odd number of arguments, not a map",
			args:    args{"id", 1, "name"},
			wantErr: true,
		},
		{
			name:    "Key not a string",
			args:    args{"id", 1, 25, "name"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertArgsToMap(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertArgsToMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertArgsToMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}
