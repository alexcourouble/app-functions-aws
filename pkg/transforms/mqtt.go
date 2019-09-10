package transforms

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/appsdk"
	sdkTransforms "github.com/edgexfoundry/app-functions-sdk-go/pkg/transforms"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

const (
	awsIoTMQTTHost           = "AwsIoTMQTTHost"
	awsIoTMQTTPort           = 8883
	awsIoTThingName          = "awsIoTThingName"
	awsIoTRootCAFilename     = "CaCertPath"
	awsIoTCertFilename       = "MQTTCert"
	awsIoTPrivateKeyFilename = "MQTTKey"
	user                     = "someUser"
)

var log logger.LoggingClient

type certPair struct {
	Cert string `json:"cert,omitempty"`
	Key  string `json:"key,omitempty"`
}

// AWSMQTTConfig holds AWS IoT specific information
type AWSMQTTConfig struct {
	MQTTConfig  *sdkTransforms.MqttConfig
	IoTHost     string
	IoTDevice   string
	KeyCertPair *sdkTransforms.KeyCertPair
}

func getNewClient(skipVerify bool) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}

	return &http.Client{Timeout: 10 * time.Second, Transport: tr}
}

func getAppSetting(settings map[string]string, name string) string {
	value, ok := settings[name]

	if ok {
		return value
	}
	log.Error(fmt.Sprintf("ApplicationName application setting %s not found", name))
	return ""

}

// LoadAWSMQTTConfig Loads the mqtt configuration necessary to connect to AWS cloud
func LoadAWSMQTTConfig(sdk *appsdk.AppFunctionsSDK) (*AWSMQTTConfig, error) {
	if sdk == nil {
		return nil, errors.New("Invalid AppFunctionsSDK")
	}

	log = sdk.LoggingClient

	var ioTHost, iotDevice, mqttCert, mqttKey string

	appSettings := sdk.ApplicationSettings()
	if appSettings != nil {
		ioTHost = getAppSetting(appSettings, awsIoTMQTTHost)
		iotDevice = getAppSetting(appSettings, awsIoTThingName)
		mqttCert = getAppSetting(appSettings, awsIoTCertFilename)
		mqttKey = getAppSetting(appSettings, awsIoTPrivateKeyFilename)
	} else {
		return nil, errors.New("No application-specific settings found")
	}

	config := AWSMQTTConfig{}

	config.IoTHost = ioTHost
	config.IoTDevice = iotDevice
	config.MQTTConfig = sdkTransforms.NewMqttConfig()

	pair := &sdkTransforms.KeyCertPair{
		KeyFile:  mqttKey,
		CertFile: mqttCert,
	}

	config.KeyCertPair = pair

	return &config, nil
}

// NewAWSMQTTSender return a mqtt sender capable of sending the event's value to the given MQTT broker
func NewAWSMQTTSender(logging logger.LoggingClient, config *AWSMQTTConfig) *sdkTransforms.MQTTSender {

	// TODO: configurable topic?

	topic := fmt.Sprintf("thing/%s/messages/", config.IoTDevice)

	addressable := models.Addressable{
		Address:   awsIoTMQTTHost,
		Port:      awsIoTMQTTPort,
		Protocol:  "tls",
		Publisher: awsIoTThingName,
		User:      "",
		Password:  "",
		Topic:     topic,
	}

	mqttSender := sdkTransforms.NewMQTTSender(logging, addressable, config.KeyCertPair, config.MQTTConfig)

	return mqttSender
}
