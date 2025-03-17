package configs

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
)

type AppConfig struct {
	ProjectName string `json:"projectName"`
	Environment string `json:"environment"`
	RabbitMq    struct {
		Host           string `json:"host"`
		Port           int    `json:"port"`
		User           string `json:"user"`
		Password       string `json:"password"`
		ConnectionName string `json:"connectionName"`
	} `json:"rabbitMq"`
}

var appConfig *AppConfig

func LoadConfig() *AppConfig {
	environment := os.Getenv("APP_ENV")
	if environment == "" {
		environment = "development"
	}

	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath("./configs")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Error reading config file, %s\n", err)
		panic(err)
	}

	viper.SetConfigName("config." + environment)
	err = viper.MergeInConfig()
	if err != nil {
		fmt.Printf("Failed to load config.%s.json %s\n", environment, err.Error())
		panic(err)
	}

	appConfig = &AppConfig{}
	err = viper.Unmarshal(&appConfig)
	if err != nil {
		fmt.Printf("Failed to unmarshal config %s\n", err.Error())
		panic(err)
	}

	appConfig.Environment = environment
	return appConfig
}

func GetConfig() *AppConfig {
	return appConfig
}

func (a *AppConfig) GetRabbitMqConnectionStr() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d",
		a.RabbitMq.User,
		a.RabbitMq.Password,
		a.RabbitMq.Host,
		a.RabbitMq.Port)
}
