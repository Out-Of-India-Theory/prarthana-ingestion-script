package configuration

import (
	"fmt"
	"github.com/Out-Of-India-Theory/oit-go-commons/config"
	"github.com/spf13/viper"
	"strings"
)

var configuration *Configuration

type Configuration struct {
	ServerConfig config.AppConfig
	MongoConfig  MongoConfig
}

type MongoConfig struct {
	config.MongoConfig
	Database string
}

func addConfigPath(v *viper.Viper) {
	v.AddConfigPath(".")
	v.AddConfigPath("config")
}

func init() {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("json")
	addConfigPath(v)
	v.AutomaticEnv()
	var err error
	if err = v.ReadInConfig(); err != nil {
		fmt.Printf("error while reading config file, %v\n", err)
		panic(err)
	}
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err = v.Unmarshal(&configuration); err != nil {
		fmt.Printf("error while deserializing config, %v\n", err)
		panic(err)
	}
}

func GetConfig() *Configuration {
	return configuration
}
