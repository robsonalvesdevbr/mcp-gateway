package eval

import "fmt"

func volume(value any) string {
	source := fmt.Sprintf("%v", value)
	target := source
	if isWindowsPath(target) {
		target = toLinuxPath(target)
	}
	return fmt.Sprintf("%s:%s", source, target)
}
