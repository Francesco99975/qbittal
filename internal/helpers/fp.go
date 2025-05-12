package helpers

import "sort"

func FilteredSlice[T any](arr []T, test func(T) bool) []T {
	var result []T

	for _, item := range arr {
		if test(item) {
			result = append(result, item)
		}
	}

	return result
}

func SortSlice[T any](arr []T, less func(a, b T) bool) {
	sort.Slice(arr, func(i, j int) bool {
		return less(arr[i], arr[j])
	})
}
