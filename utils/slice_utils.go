package utils

import "golang.org/x/exp/constraints"

func InsertUnique[T constraints.Ordered](slice []T, element T) []T {
	for _, v := range slice {
		if v == element {
			return slice
		}
	}
	return append(slice, element)
}
