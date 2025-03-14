package config

import (
	"anicliru/internal/api/player/common"
	"anicliru/internal/api/providers"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/pelletier/go-toml/v2"
)

type VideoCfg struct {
	Dub     string `toml:"dub"`
	Quality int    `toml:"quality"`
}

type converterCfg struct {
	SyncInterval string   `toml:"syncInterval"`
	Domains      []string `toml:"domains,omitempty"`
}

type providersCfg struct {
	AutoSync  bool              `toml:"autoSync"`
	DomainMap map[string]string `toml:"domainMap,omitempty"`
}

type Config struct {
	cfgPath   string `toml:"-"`
	Video     VideoCfg
	Providers providersCfg
	Players   converterCfg
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
		Providers: providersCfg{
			AutoSync:  true,
			DomainMap: make(map[string]string),
		},
		Players: converterCfg{
			Domains:      domains,
			SyncInterval: defaultSyncInterval,
		},
		Video: VideoCfg{
			Dub:     "",
			Quality: 1080,
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

	if err := cfg.check(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (cfg *Config) check() error {
	if len(cfg.Providers.DomainMap) == 0 && !cfg.Providers.AutoSync {
		return errors.New("все источники отключены в конфиге")
	}

	availableProviders := providers.GetProviders()
	for provider := range cfg.Providers.DomainMap {
		if !isInSlice(provider, availableProviders) {
			return fmt.Errorf("в конфиге указан не существующий источник %s", provider)
		}
	}

	availablePlayers := common.GetPlayerDomains()
	for _, provider := range cfg.Players.Domains {
		if !isInSlice(provider, availablePlayers) {
			return fmt.Errorf("в конфиге указан домен не существующего плеера %s", provider)
		}
	}

	if len(cfg.Players.Domains) == 0 {
		return errors.New("все плееры отключены в конфиге")
	}

	if !isDayInterval(cfg.Players.SyncInterval) {
		return errors.New("некорректная дата обновления в конфиге")
	}

	if cfg.Video.Quality <= 0 {
		return errors.New("неверное значение качества видео в конфиге")
	}

	return nil
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
