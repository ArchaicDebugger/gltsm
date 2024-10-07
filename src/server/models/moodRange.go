package models

import "time"

type MoodRange struct {
	Date      time.Time
	StartTime time.Time
	EndTime   time.Time
	HasValue  bool
}

func Chunk[T any](m []T, chunkSize int) [][]T {
	l := len(m)

	var result [][]T

	for i := 0; i < l; i += chunkSize {
		var slice []T
		for j := i; j < i+chunkSize && j < l; j++ {
			slice = append(slice, (m)[j])
		}
		result = append(result, slice)
	}

	return result
}

func MapToArray[T any](m *map[string]T) []T {
	var result []T

	for _, value := range *m {
		result = append(result, value)
	}

	return result
}
