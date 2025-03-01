package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Providers []string
}

func getConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
    if err != nil {
        return "", fmt.Errorf("не удалось получить директорию для конфигов из переменных окружения среды. %s", err)
    }

    configPath := filepath.Join(dir, "anicli-ru", "config.toml")
	return configPath, nil
}

func newDefaultConfig(cfgPath string) (*Config, error) {
	cfg := Config{Providers: []string{"animego", "yummyanime"}}

	dir := filepath.Dir(cfgPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("не удалось создать директорию для конфига. %s", err)
	}

	cfgToml, err := toml.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(cfgPath, cfgToml, 0666)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Загружает конфиг. Если конфига не существует, то создаст стандартный и вернёт его.
func LoadConfig() (*Config, error) {
	cfgPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}
    
    cfgToml, err := os.ReadFile(cfgPath)
    if err != nil {
        return newDefaultConfig(cfgPath)
    }

    var cfg Config
    if err := toml.Unmarshal(cfgToml, &cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}
