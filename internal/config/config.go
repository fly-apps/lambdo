package config

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
)

type LambdoConfig struct {
	Environment        string `mapstructure:"env"`
	SQSLongPollSeconds int    `mapstructure:"sqs_long_poll_seconds"`
	SQSQueueUrl        string `mapstructure:"sqs_queue_url"`
	EventsPerMachine   int    `mapstructure:"events_per_machine"`
	FlyApp             string `mapstructure:"fly_app"`
	FlyRegion          string `mapstructure:"fly_region"`
	FlyToken           string `mapstructure:"fly_token"`
}

var lambdoConfig *LambdoConfig

func Configure() error {
	v := viper.New()
	v.SetEnvPrefix("lambdo")
	v.AutomaticEnv()

	v.BindEnv("env")
	v.BindEnv("sqs_long_poll_seconds")
	v.BindEnv("sqs_queue_url")
	v.BindEnv("events_per_machine")
	v.BindEnv("fly_app")
	v.BindEnv("fly_region")
	v.BindEnv("fly_token")

	v.SetDefault("env", "local")
	v.SetDefault("sqs_long_poll_seconds", 10)
	v.SetDefault("events_per_machine", 5)

	config := &LambdoConfig{}
	err := v.Unmarshal(&config)

	if err != nil {
		return err
	}

	// Pick these up from Fly runtime environment variables
	// if not set explicitly
	if len(config.FlyApp) == 0 {
		config.FlyApp = os.Getenv("FLY_APP_NAME")
	}

	if len(config.FlyRegion) == 0 {
		config.FlyRegion = os.Getenv("FLY_REGION")
	}

	// Some quick validation
	if len(config.FlyToken) == 0 {
		return fmt.Errorf("config LAMBDO_FLY_TOKEN must be set")
	}

	if len(config.FlyApp) == 0 {
		return fmt.Errorf("No values found for LAMDBO_FLY_APP nor FLY_APP_NAME")
	}

	if len(config.FlyRegion) == 0 {
		return fmt.Errorf("No values found for LAMBDO_FLY_REGION nor FLY_REGION")
	}

	if config.EventsPerMachine > 10 {
		log.Println("config events_per_machine set higher than 10, using value 10")
		config.EventsPerMachine = 10
	}

	lambdoConfig = config

	return nil
}

func GetConfig() *LambdoConfig {
	return lambdoConfig
}
