package hermes

import (
	"github.com/spf13/viper"
)

func newViper() *viper.Viper {
	config := viper.New()
	// config.SetConfigFile(os.Args[1])
	config.SetConfigFile("./_res/config/hermes.yml")
	err := config.ReadInConfig()
	if err != nil {
		panic(err)
	}
	return config
}
