package config

import (
	"strconv"
	"strings"
)

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
