package config

import (
	"reflect"
	"sync"

	"github.com/rodney-b/swish-test-publisher/pkg/utilities/env"
)

type ConfigProvider interface {
	IsDevelopment() bool
	GetAppName() string
	GetAppMode() int
	GetConsoleAccessKey() string
	GetConsoleSecretKey() string
	GetDataObjectPath() string
	GetDataSourceBucket() string
	GetDataSourceClientCA() []byte
	GetDataSourceClientCert() []byte
	GetDataSourceClientCertKey() []byte
	GetDataSourceURL() string
	GetHealthcheckPort() string
	GetHealthcheckServicePrefix() string
	GetMessageQueueClientCA() []byte
	GetMessageQueueClientCert() []byte
	GetMessageQueueClientCertKey() []byte
	GetMessageQueueTopic() string
	GetMessageQueueURL() string
	GetMessageRate() int
	GetOTelHTTPReceiverURL() string
	GetOtelStdoutExporterEnabled() bool
	GetStage() string
}

// appCofnig implements ConfigProvider. It "provides" all its values from environment variables.
// each data type must have a case statement in utilities.env.Get()
// note: all types that can be casted to from int, are already covered
type appConfig struct {
	appMode                   int    `envname:"APP_MODE"`
	appName                   string `envname:"APP_NAME"`
	consoleAccessKey          string `envname:"CONSOLE_ACCESS_KEY"`
	consoleSecretKey          string `envname:"CONSOLE_SECRET_KEY"`
	dataSetOne                string `envname:"DATA_SET_ONE"`
	dataSetTwo                string `envname:"DATA_SET_TWO"`
	dataSourceBucket          string `envname:"DATA_SOURCE_BUCKET"`
	dataSourceClientCA        string `envname:"DATA_SOURCE_CA"`
	dataSourceClientCert      string `envname:"DATA_SOURCE_CRT"`
	dataSourceClientCertKey   string `envname:"DATA_SOURCE_KEY"`
	dataSourceURL             string `envname:"DATA_SOURCE_URL"`
	healthcheckPort           string `envname:"HEALTHCHECK_PORT"`
	healthcheckServicePrefix  string `envname:"HEALTHCHECK_SERVICE_PREFIX"`
	messageQueueClientCA      string `envname:"MESSAGE_QUEUE_CA"`
	messageQueueClientCert    string `envname:"MESSAGE_QUEUE_CRT"`
	messageQueueClientCertKey string `envname:"MESSAGE_QUEUE_KEY"`
	messageQueueTopicOne      string `envname:"MESSAGE_QUEUE_TOPIC_ONE"`
	messageQueueTopicTwo      string `envname:"MESSAGE_QUEUE_TOPIC_TWO"`
	messageQueueURL           string `envname:"MESSAGE_QUEUE_URL"`
	messageRateOne            uint8  `envname:"MESSAGE_RATE_ONE"`
	messageRateTwo            uint8  `envname:"MESSAGE_RATE_TWO"`
	otelHTTPReceiverURL       string `envname:"OTEL_HTTP_RECEIVER_URL"`
	otelStdoutExporterEnabled string `envname:"OTEL_STDOUT_EXPORTER_ENABLED"`
	stage                     string `envname:"STAGE"`
}

func (ac *appConfig) GetAppName() string {
	return ac.appName
}

func (ac *appConfig) GetAppMode() int {
	return ac.appMode
}

func (ac *appConfig) GetDataSourceClientCA() []byte {
	return []byte(ac.dataSourceClientCA)
}

func (ac *appConfig) GetDataSourceClientCert() []byte {
	return []byte(ac.dataSourceClientCert)
}

func (ac *appConfig) GetDataSourceClientCertKey() []byte {
	return []byte(ac.dataSourceClientCertKey)
}

func (ac *appConfig) GetConsoleAccessKey() string {
	return ac.consoleAccessKey
}

func (ac *appConfig) GetConsoleSecretKey() string {
	return ac.consoleSecretKey
}

func (ac *appConfig) GetDataObjectPath() string {
	funcOnce := sync.OnceValue(func() string {
		if ac.appMode == 1 {
			return ac.dataSetOne
		}

		return ac.dataSetTwo
	})

	return funcOnce()
}

func (ac *appConfig) GetDataSourceBucket() string {
	return ac.dataSourceBucket
}

func (ac *appConfig) GetDataSourceURL() string {
	return ac.dataSourceURL
}

func (ac *appConfig) GetHealthcheckPort() string {
	return ac.healthcheckPort
}

func (ac *appConfig) GetHealthcheckServicePrefix() string {
	return ac.healthcheckServicePrefix
}

func (ac *appConfig) GetMessageQueueClientCA() []byte {
	return []byte(ac.messageQueueClientCA)
}

func (ac *appConfig) GetMessageQueueClientCert() []byte {
	return []byte(ac.messageQueueClientCert)
}

func (ac *appConfig) GetMessageQueueClientCertKey() []byte {
	return []byte(ac.messageQueueClientCertKey)
}

func (ac *appConfig) GetMessageQueueTopic() string {
	funcOnce := sync.OnceValue(func() string {
		if ac.appMode == 1 {
			return ac.messageQueueTopicOne
		}

		return ac.messageQueueTopicTwo
	})

	return funcOnce()
}

func (ac *appConfig) GetMessageQueueURL() string {
	return ac.messageQueueURL
}

func (ac *appConfig) GetMessageRate() int {
	funcOnce := sync.OnceValue(func() int {
		if ac.appMode == 1 {
			return int(ac.messageRateOne)
		}

		return int(ac.messageRateTwo)
	})

	return funcOnce()
}

func (ac *appConfig) GetOTelHTTPReceiverURL() string {
	return ac.otelHTTPReceiverURL
}

func (ac *appConfig) GetOtelStdoutExporterEnabled() bool {
	funcOnce := sync.OnceValue(func() bool {
		if ac.otelStdoutExporterEnabled == "true" {
			return true
		}

		return false
	})

	return funcOnce()
}

func (ac *appConfig) GetStage() string {
	return ac.stage
}

// helper funcs
func (ac *appConfig) IsDevelopment() bool {
	return ac.GetStage() != Production && ac.GetStage() != Staging
}

// These are for config values the app shouldn't start without.
var initAppConfig = sync.OnceValues(func() (*appConfig, error) {
	ac := appConfig{}
	appConfVal := reflect.ValueOf(&ac).Elem()
	appConfType := reflect.TypeOf(ac)

	for i := range appConfVal.NumField() {
		fieldInfo := appConfType.Field(i)
		fieldTag := fieldInfo.Tag
		fieldValue := appConfVal.Field(i)
		fieldPtr := fieldValue.Addr().UnsafePointer()
		unsafeFieldValue := reflect.NewAt(fieldValue.Type(), fieldPtr).Elem()

		err := env.Get(fieldTag.Get("envname"), unsafeFieldValue)
		if err != nil {
			return nil, err
		}
	}

	return &ac, nil
})

func InitAppConfig() (*appConfig, error) {
	return initAppConfig()
}
