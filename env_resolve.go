package dotenvgo

import "os"

type resolvedEnvValue struct {
	raw    string
	value  string
	exists bool
}

func resolveEnvValue(key string, required bool) (resolvedEnvValue, error) {
	raw, exists := os.LookupEnv(key)
	if !exists {
		if required {
			return resolvedEnvValue{}, &RequiredError{Key: key}
		}
		return resolvedEnvValue{}, nil
	}

	return resolvedEnvValue{
		raw:    raw,
		value:  os.ExpandEnv(raw),
		exists: true,
	}, nil
}

func resolveFieldValue(key, defaultValue string, required bool) (string, bool, error) {
	resolved, err := resolveEnvValue(key, required)
	if err != nil {
		return "", false, err
	}
	if resolved.exists {
		return resolved.value, true, nil
	}

	if defaultValue == "" {
		return "", false, nil
	}

	return os.ExpandEnv(defaultValue), true, nil
}
