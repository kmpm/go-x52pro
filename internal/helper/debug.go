package helper

import (
	"os"
	"strings"
)

func cleanName(name string) string {
	return strings.Trim(strings.ToLower(name), " ")
}

// HasDebug returns true if the module has debug enabled.
// That is done by setting the module name in the environment variable DEBUG.
// DEBUG=module1,module2,module3
func HasDebug(module string) bool {
	module = cleanName(module)
	parts := strings.Split(os.Getenv("DEBUG"), ",")
	for _, part := range parts {
		part = cleanName(part)
		if part == "*" || part == module {
			return true
		}
	}
	return false
}
