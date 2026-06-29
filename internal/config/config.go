package config

import (
	"crypto/rsa"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gopkg.in/yaml.v3"
)

type Postgres struct {
	Name         string        `yaml:"dbname"`
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	User         string        `yaml:"user"`
	Password     string        `yaml:"password"`
	SSLMode      string        `yaml:"sslmode"`
	MaxOpenConns int           `yaml:"maxOpenConns"`
	MaxIdleConns int           `yaml:"maxIdleConns"`
	MaxIdleTime  time.Duration `yaml:"maxIdleTime"`
}

type HTTPServer struct {
	Address      string        `yaml:"address"`
	IdleTimeout  time.Duration `yaml:"idleTimeout"`
	ReadTimeout  time.Duration `yaml:"readTimeout"`
	WriteTimeout time.Duration `yaml:"writeTimeout"`
}

type PEM struct {
	PrivPath string `yaml:"privPath"`
	PubPath  string `yaml:"pubPath"`
}

type Log struct {
	LogPath string `yaml:"logPath"`
}

type TokenCleaner struct {
	Interval time.Duration `yaml:"interval"`
}

type Config struct {
	Postgres     `yaml:"postgres"`
	HTTPServer   `yaml:"httpServer"`
	PEM          `yaml:"pem"`
	Log          `yaml:"log"`
	TokenCleaner `yaml:"tokenCleaner"`
	Keys         KeysPair `yaml:"-"`
}

func LoadConfig() (*Config, error) {
	data, err := os.ReadFile("./config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	keys, err := LoadKeys(cfg.PEM.PrivPath, cfg.PEM.PubPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load keys: %w", err)
	}
	cfg.Keys = *keys

	return cfg, nil
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
