package utils

import (
	"reflect"
	"testing"
)

func TestInsertUniqueString(t *testing.T) {
	tests := []struct {
		name    string
		slice   []string
		element string
		want    []string
	}{
		{
			name:    "Append to nil slice",
			slice:   nil,
			element: "a",
			want:    []string{"a"},
		},
		{
			name:    "Append to empty slice",
			slice:   []string{},
			element: "a",
			want:    []string{"a"},
		},
		{
			name:    "Append new element",
			slice:   []string{"a", "b"},
			element: "c",
			want:    []string{"a", "b", "c"},
		},
		{
			name:    "Existing element is not appended",
			slice:   []string{"a", "b"},
			element: "b",
			want:    []string{"a", "b"},
		},
		{
			name:    "Existing first element is not appended",
			slice:   []string{"a", "b"},
			element: "a",
			want:    []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InsertUnique(tt.slice, tt.element)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InsertUnique() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInsertUniqueInt(t *testing.T) {
	got := InsertUnique([]int{1, 2, 3}, 2)
	if !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Errorf("InsertUnique() = %v, want %v", got, []int{1, 2, 3})
	}

	got = InsertUnique([]int{1, 2, 3}, 4)
	if !reflect.DeepEqual(got, []int{1, 2, 3, 4}) {
		t.Errorf("InsertUnique() = %v, want %v", got, []int{1, 2, 3, 4})
	}
}
