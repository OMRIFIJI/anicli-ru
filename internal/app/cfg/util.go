package config

import (
	"anicliru/internal/db"
	"bytes"
	"strconv"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
)

const providersSyncInterval = 1

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
	if daysCount < 0 {
		return false
	}
	return true
}

func isInSlice(el string, s []string) bool {
	for _, elS := range s {
		if el == elS {
			return true
		}
	}

	return false
}

func isTimeToSyncPlayers(syncInterval string, dbh *db.DBHandler, currentTime time.Time) bool {
    lastSyncTime, err := dbh.GetLastSyncTime("players")
	if err != nil {
		return true
	}
	diff := currentTime.Sub(*lastSyncTime)
	days := int(diff.Hours() / 24)

	syncIntervalInt, err := strconv.Atoi(syncInterval[:len(syncInterval)-1])

	return days >= syncIntervalInt
}

func isTimeToSyncProviders(dbh *db.DBHandler, currentTime time.Time) bool {
    lastSyncTime, err := dbh.GetLastSyncTime("providers")
	if err != nil {
		return true
	}
	diff := currentTime.Sub(*lastSyncTime)
	days := int(diff.Hours() / 24)

	return days >= providersSyncInterval
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
