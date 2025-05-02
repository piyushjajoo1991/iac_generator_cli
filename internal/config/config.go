package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// CfgFile holds the config file path
var CfgFile string

// Config holds configuration values
type Config struct {
	LogLevel    string `mapstructure:"log_level"`
	OutputDir   string `mapstructure:"output_dir"`
	DefaultType string `mapstructure:"default_type"`
}

// AppConfig holds the application config
var AppConfig Config

// InitConfig reads in config file and ENV variables if set
func InitConfig() {
	if CfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(CfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".iacgen" (without extension)
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName(".iacgen")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// Set defaults
	viper.SetDefault("log_level", "info")
	viper.SetDefault("output_dir", ".")
	viper.SetDefault("default_type", "terraform")

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	// Unmarshal config into AppConfig
	if err := viper.Unmarshal(&AppConfig); err != nil {
		fmt.Printf("Unable to decode config into struct: %v\n", err)
	}
}

// SaveConfig saves the current configuration to file
func SaveConfig() error {
	configDir := filepath.Dir(viper.ConfigFileUsed())
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configDir = home
	}

	// Ensure directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Write config
	return viper.WriteConfig()
}
