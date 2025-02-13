package conf

import (
	"log"

	"github.com/spf13/viper"
)

func Config(path string) *viper.Viper {
	viper.SetConfigName("conf") // Name without extension
	viper.SetConfigType("yaml") // File type
	viper.AddConfigPath(path)   // Look for config in the current directory

	// Read configuration file
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	return viper.GetViper()
}
