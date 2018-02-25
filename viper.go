package hermes

import (
	"github.com/spf13/viper"
)

func newViper() *viper.Viper {
	config := viper.New()
	config.SetConfigFile("./res/config/general.yml")
	err := config.ReadInConfig()
	if err != nil {
		panic(err)
	}
	return config
}
