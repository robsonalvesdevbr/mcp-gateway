package eval

import "fmt"

func volumeTarget(value any) string {
	if isEmptyOrInvalid(value) {
		return "" // Skip volume target inválido
	}

	path := fmt.Sprintf("%v", value)
	if isWindowsPath(path) {
		return toLinuxPath(path)
	}
	return path
}
