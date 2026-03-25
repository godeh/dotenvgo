package dotenvgo

import (
	"reflect"
	"strings"
)

func parseSliceValue(sliceType reflect.Type, value, sep string, parseElem func(string) (reflect.Value, error)) (reflect.Value, error) {
	if sep == "" {
		sep = ","
	}

	parts := strings.Split(value, sep)
	slice := reflect.MakeSlice(sliceType, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		elem, err := parseElem(part)
		if err != nil {
			return reflect.Value{}, err
		}
		slice = reflect.Append(slice, elem)
	}

	return slice, nil
}
