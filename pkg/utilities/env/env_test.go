package env_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/rodney-b/swish-test-publisher/pkg/utilities/env"
)

type testAppConfig struct {
	appName        string `envname:"APP_NAME"`
	timeoutSeconds uint8  `envname:"TIMEOUT_SECONDS"`
}

func TestGet(t *testing.T) {
	expected := testAppConfig{
		appName:        "env-test",
		timeoutSeconds: 10,
	}

	t.Setenv("APP_NAME", expected.appName)
	t.Setenv("TIMEOUT_SECONDS", strconv.Itoa(int(expected.timeoutSeconds)))
	tc := testAppConfig{}

	testVal := reflect.ValueOf(&tc).Elem()
	testType := reflect.TypeOf(tc)

	// mimicking the logic used by confg to retrieve env values
	for i := range testVal.NumField() {
		fieldInfo := testType.Field(i)
		fieldTag := fieldInfo.Tag
		fieldValue := testVal.Field(i)
		envVarName := fieldTag.Get("envname")

		fieldPtr := fieldValue.Addr().UnsafePointer()
		unsafeFieldValue := reflect.NewAt(fieldValue.Type(), fieldPtr).Elem()

		err := env.Get(envVarName, unsafeFieldValue)
		if err != nil {
			t.Fatalf("error getting env var %s: %v", envVarName, err)
		}
	}

	errBadValue := "invalid value for %s - expected %v but got %v"

	if tc.appName != expected.appName {
		t.Fatalf(errBadValue, "appName", expected.appName, tc.appName)
	}

	if tc.timeoutSeconds != expected.timeoutSeconds {
		t.Fatalf(errBadValue, "timeoutSeconds", expected.timeoutSeconds, tc.timeoutSeconds)
	}
}
