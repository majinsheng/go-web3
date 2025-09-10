package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Ethereum EthereumConfig
}

// ServerConfig holds configuration for the REST API server
type ServerConfig struct {
	Port string
	Host string
}

// EthereumConfig holds configuration for ethereum connection
type EthereumConfig struct {
	Provider   string
	ChainID    int64
	PrivateKey string
}

// LoadConfig loads the configuration from file and environment variables
func LoadConfig() (*Config, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Error loading .env file:", err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// Set default values
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("ethereum.provider", "http://localhost:8545")
	viper.SetDefault("ethereum.chainID", 1)

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// Override with environment variables
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Use environment variables for sensitive data
	privateKey := os.Getenv("PRIVATE_KEY")
	if privateKey != "" {
		viper.Set("ethereum.privateKey", privateKey)
	}

	// If INFURA_API_KEY is provided, use it to set the provider
	infuraKey := os.Getenv("INFURA_API_KEY")
	if infuraKey != "" {
		viper.Set("ethereum.provider", "https://mainnet.infura.io/v3/"+infuraKey)
	}

	// Unmarshal config
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
