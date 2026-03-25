package dotenvgo

import (
	"os"
	"strings"
)

// Export returns all environment variables as a map.
func Export() map[string]string {
	result := make(map[string]string)
	for _, env := range os.Environ() {
		idx := strings.Index(env, "=")
		if idx != -1 {
			result[env[:idx]] = env[idx+1:]
		}
	}
	return result
}

// ExportWithPrefix returns environment variables matching a prefix.
func ExportWithPrefix(prefix string) map[string]string {
	result := make(map[string]string)
	prefixUpper := strings.ToUpper(prefix)
	if !strings.HasSuffix(prefixUpper, "_") {
		prefixUpper += "_"
	}

	for _, env := range os.Environ() {
		idx := strings.Index(env, "=")
		if idx != -1 {
			key := env[:idx]
			if strings.HasPrefix(strings.ToUpper(key), prefixUpper) {
				result[key] = env[idx+1:]
			}
		}
	}
	return result
}

// Set sets an environment variable.
func Set(key, value string) {
	_ = os.Setenv(key, value)
}

// Unset removes an environment variable.
func Unset(key string) {
	_ = os.Unsetenv(key)
}
