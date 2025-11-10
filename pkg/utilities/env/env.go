package env

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var (
	ErrNoValidEnvTypeProvided = errors.New("no valid environment variable data type provided")
)

func Get(name string, val reflect.Value) error {
	if _, ok := os.LookupEnv(name); !ok {
		return fmt.Errorf("unable to find environment variable named %q", name)
	}

	varStrVal := os.Getenv(name)

	// add a case statment for every data type used by config.appConfig's fields
	// that is, a statment for every data type the environment variables will be casted to
	switch val.Interface().(type) {
	case uint8:
		varIntVal, err := strconv.ParseUint(varStrVal, 10, 8)
		if err != nil {
			return err
		}

		val.SetUint(varIntVal)
	case uint16:
		varIntVal, err := strconv.ParseUint(varStrVal, 10, 16)
		if err != nil {
			return err
		}

		val.SetUint(varIntVal)
	case uint32:
		varIntVal, err := strconv.ParseUint(varStrVal, 10, 32)
		if err != nil {
			return err
		}

		val.SetUint(varIntVal)
	case int:
		varIntVal, err := strconv.Atoi(varStrVal)
		if err != nil {
			return err
		}

		val.SetInt(int64(varIntVal))
	case string:
		val.SetString(varStrVal)
	case []string:
		splitVal := reflect.ValueOf(strings.Split(varStrVal, " "))
		val.Set(splitVal)
	default:
		return ErrNoValidEnvTypeProvided
	}

	return nil
}
