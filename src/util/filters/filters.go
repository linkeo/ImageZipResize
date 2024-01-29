package filters

import "os"

type StringFilter func(value string) bool

var PathIsDirectory StringFilter = func(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return stat.IsDir()
}

var PathIsRegularFile StringFilter = func(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return stat.Mode().IsRegular()
}

func Equal[T comparable](target T) func(value T) bool {
	return func(value T) bool {
		return value == target
	}
}
