package eval

import (
	"fmt"
	"reflect"
)

// isEmptyOrInvalid detecta valores que não devem virar volumes Docker.
// Retorna true para: nil, arrays vazios, slices vazios, maps vazios,
// e strings que representam valores vazios ("[]", "{}", "<nil>", "null").
func isEmptyOrInvalid(value any) bool {
	if value == nil {
		return true
	}

	// Use reflection para detectar valores vazios
	v := reflect.ValueOf(value)

	// Arrays/slices vazios
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		return v.Len() == 0
	}

	// Maps vazios
	if v.Kind() == reflect.Map {
		return v.Len() == 0
	}

	// String conversions - detectar representações de valores vazios
	str := fmt.Sprintf("%v", value)

	// Strings vazias ou representações de valores vazios
	return str == "" ||
		str == "[]" ||
		str == "{}" ||
		str == "<nil>" ||
		str == "null"
}

func volume(value any) string {
	if isEmptyOrInvalid(value) {
		return "" // Skip volume mount inválido
	}

	source := fmt.Sprintf("%v", value)
	target := source
	if isWindowsPath(target) {
		target = toLinuxPath(target)
	}
	return fmt.Sprintf("%s:%s", source, target)
}
