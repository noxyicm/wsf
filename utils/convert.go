package utils

import (
	"reflect"
	"strconv"
	"wsf/errors"
)

// InterfaceToString tryes to convert interface to string
func InterfaceToString(val interface{}) (string, error) {
	switch v := val.(type) {
	case string:
		return v, nil

	case int:
		return strconv.Itoa(int(v)), nil
	case int8:
		return strconv.Itoa(int(v)), nil
	case int16:
		return strconv.Itoa(int(v)), nil
	case int32:
		return strconv.Itoa(int(v)), nil
	case int64:
		return strconv.Itoa(int(v)), nil

	case float32:
		return strconv.FormatFloat(float64(v), 'f', 10, 32), nil

	case float64:
		return strconv.FormatFloat(v, 'f', 10, 32), nil

	case []byte:
		return string(v), nil
	}

	return "", errors.Errorf("Value of type '%s' can not be converted to string", reflect.TypeOf(val))
}

// InterfaceToInt tryes to convert interface to int
func InterfaceToInt(val interface{}) (int, error) {
	switch v := val.(type) {
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			return 0, errors.Errorf("Value '%s' can not be converted to integer", v)
		}
		return i, nil

	case int:
		return int(v), nil
	case int8:
		return int(v), nil
	case int16:
		return int(v), nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float32:
		return int(v), nil
	case float64:
		return int(v), nil
	}

	return 0, errors.Errorf("Value of type '%s' can not be converted to integer", reflect.TypeOf(val))
}
