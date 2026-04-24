package config

import (
	"crypto/rsa"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

type Database struct {
	DBUrl string `mapstructure:"dbUrl"`
}

type HTTPServer struct {
	Address      string        `mapstructure:"address"`
	IdleTimeout  time.Duration `mapstructure:"idleTimeout"`
	ReadTimeout  time.Duration `mapstructure:"readTimeout"`
	WriteTimeout time.Duration `mapstructure:"writeTimeout"`
}

type PEM struct {
	PrivPath string `mapstructure:"privPath"`
	PubPath  string `mapstructure:"pubPath"`
}

type Config struct {
	Database   `mapstructure:"database"`
	HTTPServer `mapstructure:"httpServer"`
	PEM        `mapstructure:"pem"`
	Keys       *KeysPair `mapstructure:"-"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config into struct: %w", err)
	}

	keys, err := LoadKeys(cfg.PEM.PrivPath, cfg.PEM.PubPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load keys: %w", err)
	}
	cfg.Keys = keys

	return &cfg, nil
}

type KeysPair struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

func LoadKeys(privPath, pubPath string) (*KeysPair, error) {
	privKeyData, err := os.ReadFile(privPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key from %s: %w", privPath, err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privKeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	pubKeyData, err := os.ReadFile(pubPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key from %s: %w", pubPath, err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(pubKeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	// log.Printf("Keys to string:\n%v", string(pubKeyData))

	return &KeysPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}
