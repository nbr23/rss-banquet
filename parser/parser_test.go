package parser

import (
	"reflect"
	"testing"
)

func TestOptionsList_Get(t *testing.T) {
	tests := []struct {
		name    string
		options OptionsList
		key     string
		want    interface{}
		wantErr bool
	}{
		{
			name: "String option",
			options: OptionsList{
				{Flag: "str", Type: "string", Value: "test"},
			},
			key:  "str",
			want: "test",
		},
		{
			name: "String option with pointer",
			options: OptionsList{
				{Flag: "str", Type: "string", Value: func() interface{} { s := "test"; return &s }()},
			},
			key:  "str",
			want: "test",
		},
		{
			name: "StringSlice option",
			options: OptionsList{
				{Flag: "slice", Type: "stringSlice", Value: "a,b,c"},
			},
			key:  "slice",
			want: []string{"a", "b", "c"},
		},
		{
			name: "StringSlice option with pointer",
			options: OptionsList{
				{Flag: "slice", Type: "stringSlice", Value: func() interface{} { s := "a,b,c"; return &s }()},
			},
			key:  "slice",
			want: []string{"a", "b", "c"},
		},
		{
			name: "Int option",
			options: OptionsList{
				{Flag: "num", Type: "int", Value: "42"},
			},
			key:  "num",
			want: 42,
		},
		{
			name: "Int option with int value",
			options: OptionsList{
				{Flag: "num", Type: "int", Value: 42},
			},
			key:  "num",
			want: 42,
		},
		{
			name: "Int option with pointer",
			options: OptionsList{
				{Flag: "num", Type: "int", Value: func() interface{} { i := 42; return &i }()},
			},
			key:  "num",
			want: 42,
		},
		{
			name: "Bool option true",
			options: OptionsList{
				{Flag: "flag", Type: "bool", Value: "true"},
			},
			key:  "flag",
			want: true,
		},
		{
			name: "Bool option false",
			options: OptionsList{
				{Flag: "flag", Type: "bool", Value: "false"},
			},
			key:  "flag",
			want: false,
		},
		{
			name: "Bool option 1",
			options: OptionsList{
				{Flag: "flag", Type: "bool", Value: "1"},
			},
			key:  "flag",
			want: true,
		},
		{
			name: "Bool option 0",
			options: OptionsList{
				{Flag: "flag", Type: "bool", Value: "0"},
			},
			key:  "flag",
			want: false,
		},
		{
			name: "Bool option with pointer",
			options: OptionsList{
				{Flag: "flag", Type: "bool", Value: func() interface{} { b := true; return &b }()},
			},
			key:  "flag",
			want: true,
		},
		{
			name: "Default option",
			options: OptionsList{
				{Flag: "default", Type: "unknown", Value: 123},
			},
			key:  "default",
			want: 123,
		},
		{
			name:    "Option not found",
			options: OptionsList{},
			key:     "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.options.Get(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("OptionsList.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				if *(got.(*bool)) == tt.want.(bool) {
					return
				}
				t.Errorf("OptionsList.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
