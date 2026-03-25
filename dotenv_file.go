package dotenvgo

import (
	"os"
	"strings"
)

// LoadDotEnv loads environment variables from a .env file.
// By default, it does NOT override existing environment variables.
// Pass true as the second argument to override existing variables.
//
// Examples:
//
//	LoadDotEnv(".env")        // doesn't override existing vars
//	LoadDotEnv(".env", false) // same as above
//	LoadDotEnv(".env", true)  // overrides existing vars
func LoadDotEnv(path string, override ...bool) error {
	shouldOverride := false
	if len(override) > 0 {
		shouldOverride = override[0]
	}
	return loadDotEnvInternal(path, shouldOverride)
}

func loadDotEnvInternal(path string, override bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	entries, err := parseDotEnvEntries(string(data))
	if err != nil {
		return err
	}

	resolved := make(map[string]string, len(entries))
	for _, entry := range entries {
		value := os.Expand(entry.value, func(name string) string {
			if resolvedValue, ok := resolved[name]; ok {
				return resolvedValue
			}
			envValue, _ := os.LookupEnv(name)
			return envValue
		})

		if override {
			_ = os.Setenv(entry.key, value)
			resolved[entry.key] = value
			continue
		}

		if existing, exists := os.LookupEnv(entry.key); exists {
			resolved[entry.key] = existing
			continue
		}

		_ = os.Setenv(entry.key, value)
		resolved[entry.key] = value
	}

	return nil
}

type dotenvEntry struct {
	key   string
	value string
}

func parseDotEnvEntries(data string) ([]dotenvEntry, error) {
	lines := strings.Split(data, "\n")
	entries := make([]dotenvEntry, 0, len(lines))

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		before, after, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key := strings.TrimSpace(before)
		if key == "" {
			continue
		}

		value, nextLine, err := parseDotEnvValue(lines, i, strings.TrimLeft(after, " \t"))
		if err != nil {
			return nil, err
		}

		entries = append(entries, dotenvEntry{key: key, value: value})
		i = nextLine
	}

	return entries, nil
}

func parseDotEnvValue(lines []string, start int, valuePart string) (string, int, error) {
	if valuePart == "" {
		return "", start, nil
	}

	quote := valuePart[0]
	if quote == '"' || quote == '\'' {
		return parseQuotedDotEnvValue(lines, start, valuePart, quote)
	}

	return parseUnquotedDotEnvValue(valuePart), start, nil
}

func parseQuotedDotEnvValue(lines []string, start int, valuePart string, quote byte) (string, int, error) {
	var b strings.Builder
	currentLine := start
	segment := valuePart[1:]

	for {
		for i := 0; i < len(segment); i++ {
			ch := segment[i]
			if quote == '"' && ch == '\\' && i+1 < len(segment) {
				i++
				switch segment[i] {
				case 'n':
					b.WriteByte('\n')
				case 'r':
					b.WriteByte('\r')
				case 't':
					b.WriteByte('\t')
				case '\\', '"', '$':
					b.WriteByte(segment[i])
				default:
					b.WriteByte(segment[i])
				}
				continue
			}

			if ch == quote {
				return b.String(), currentLine, nil
			}

			b.WriteByte(ch)
		}

		if currentLine+1 >= len(lines) {
			return b.String(), currentLine, nil
		}

		b.WriteByte('\n')
		currentLine++
		segment = lines[currentLine]
	}
}

func parseUnquotedDotEnvValue(valuePart string) string {
	value := valuePart
	for i := 0; i < len(valuePart); i++ {
		if valuePart[i] != '#' {
			continue
		}
		if i == 0 {
			return ""
		}
		if valuePart[i-1] == ' ' || valuePart[i-1] == '\t' {
			return strings.TrimSpace(valuePart[:i])
		}
	}

	return value
}

// MustLoadDotEnv loads a .env file or panics.
func MustLoadDotEnv(path string) {
	if err := LoadDotEnv(path); err != nil {
		panic(err)
	}
}
