package hermes

import (
	"github.com/spf13/viper"
)

func newViper(path string) *viper.Viper {
	config := viper.New()
	// config.SetConfigFile(os.Args[1])
	config.SetConfigFile(path)
	err := config.ReadInConfig()
	if err != nil {
		panic(err)
	}
	return config
}
