package config

import (
	"anicliru/internal/api/player/common"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/adrg/xdg"
	"github.com/pelletier/go-toml/v2"
)

type ConverterCfg struct {
	SyncInterval string   `toml:"syncInterval"`
	Domains      []string `toml:"domains,omitempty"`
}

type Config struct {
	cfgPath   string            `toml:"-"`
	Providers map[string]string `toml:"Providers,omitempty"`
	Players   ConverterCfg
}

func (cfg *Config) Write() error {
	cfgToml, err := prettyMarshal(cfg)
	if err != nil {
		return err
	}

	err = os.WriteFile(cfg.cfgPath, cfgToml, 0666)
	if err != nil {
		return err
	}

	return nil
}

func getConfigPath() (string, error) {
	configHome, err := xdg.ConfigFile("anicli-ru")
	if err != nil {
		return "", nil
	}

	configPath := filepath.Join(configHome, "config.toml")
	return configPath, nil
}

func newDefaultConfig(cfgPath string) (*Config, error) {
	defaultSyncInterval := "3d"
	playerOriginMap := common.NewPlayerOriginMap()
	var domains []string

	for key := range playerOriginMap {
		domains = append(domains, key)
	}

	cfg := Config{
		cfgPath: cfgPath,
		Providers: map[string]string{
			"animego":    "animego.club",
			"yummyanime": "yummy-anime.ru",
		},
		Players: ConverterCfg{
			Domains:      domains,
			SyncInterval: defaultSyncInterval,
		},
	}

	dir := filepath.Dir(cfgPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("не удалось создать директорию для конфига. %s", err)
	}

	cfgToml, err := prettyMarshal(&cfg)
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

	cfg := Config{cfgPath: cfgPath}
	if err := toml.Unmarshal(cfgToml, &cfg); err != nil {
		return nil, errors.New("не удалось загрузить конфиг")
	}

	if len(cfg.Providers) == 0 {
		return nil, errors.New("все источники отключены в конфиге")
	}

	if len(cfg.Players.Domains) == 0 {
		return nil, errors.New("все плееры отключены в конфиге")
	}

	if !isDayInterval(cfg.Players.SyncInterval) {
		return nil, errors.New("некорректная дата обновления в конфиге")
	}

	return &cfg, nil
}

func prettyMarshal(cfg *Config) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	encoder := toml.NewEncoder(buf)
	encoder.SetArraysMultiline(true)

	err := encoder.Encode(cfg)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func isDayInterval(syncInterval string) bool {
	// Пустую строку допускаем - отключает синхронизацию
	if len(syncInterval) == 0 {
		return true
	}
	// Проверяем является ли "{положительное число}d"
	before, after, found := strings.Cut(syncInterval, "d")
	if !found {
		return false
	}
	if len(after) != 0 {
		return false
	}
	daysCount, err := strconv.Atoi(before)
	if err != nil {
		return false
	}
	if daysCount <= 0 {
		return false
	}
	return true
}
