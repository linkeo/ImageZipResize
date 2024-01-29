package slices

func Filter[T comparable](slice []T, predict func(value T) bool) (result []T) {
	for _, value := range slice {
		if predict(value) {
			result = append(result, value)
		}
	}
	return result
}

func Any[T comparable](slice []T, predict func(value T) bool) bool {
	for _, value := range slice {
		if predict(value) {
			return true
		}
	}
	return false
}

func Every[T comparable](slice []T, predict func(value T) bool) bool {
	for _, value := range slice {
		if !predict(value) {
			return false
		}
	}
	return true
}

func ForEach[T any](slice []T, iterate func(value T)) {
	for _, value := range slice {
		iterate(value)
	}
}

func Find[T comparable](slice []T, predict func(value T) bool) (value T, index int, found bool) {
	for i, value := range slice {
		if predict(value) {
			return value, i, true
		}
	}
	return
}

func FindLast[T comparable](slice []T, predict func(value T) bool) (value T, index int, found bool) {
	if len(slice) == 0 {
		return
	}
	for i := len(slice) - 1; i >= 0; i-- {
		if predict(slice[i]) {
			return slice[i], i, true
		}
	}
	return
}
