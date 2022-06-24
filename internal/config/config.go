package config

import (
	"strings"

	"github.com/spf13/viper"
)

var (
	// the Config for etcd adapter
	Config config
)

// Init load and unmarshal config file
func Init(configFile string) error {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}
	viper.AddConfigPath("config")

	// read configuration file
	viper.SetEnvPrefix("EA")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	// parse configuration
	err = viper.Unmarshal(&Config)
	if err != nil {
		return err
	}

	return nil
}
