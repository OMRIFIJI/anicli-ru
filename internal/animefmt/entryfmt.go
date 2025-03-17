package animefmt

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/models"
)

// Возвращает названия аниме и количество вышедших серий из общего количества серий.
func WrapAnimeTitlesAired(animes []models.Anime) []string {
	wrappedTitles := make([]string, 0, len(animes))
	for _, anime := range animes {
		wrappedTitle := wrapAnimeTitleAired(anime)
		wrappedTitles = append(wrappedTitles, wrappedTitle)
	}
	return wrappedTitles
}

// Возвращает названия аниме и количество просмотренных серий из количества вышедших серий.
func WrapAnimeTitlesWatched(animes []models.Anime) []string {
	wrappedTitles := make([]string, 0, len(animes))
	for _, anime := range animes {
		wrappedTitle := wrapAnimeTitleWatched(anime)
		wrappedTitles = append(wrappedTitles, wrappedTitle)
	}
	return wrappedTitles
}

// Возвращает []string с номерами эпизодов из epCtx.
func EpisodeEntries(epCtx models.EpisodesContext) []string {
	epEntries := make([]string, epCtx.AiredEpCount)
	for i := 0; i < epCtx.AiredEpCount; i++ {
		epEntries[i] = strconv.Itoa(i + 1)
	}

	return epEntries
}

// Возвращает строку вида "Серия 'номер текущей' из 'количество доступных'. 'Название аниме'.".
func PlayerMenuHeader(anime *models.Anime) string {
	return fmt.Sprintf("Серия %d из %d. %s.", anime.EpCtx.Cur, anime.EpCtx.AiredEpCount, anime.Title)
}

// Возвращает строку вида "Серия current, lowerDubName. title.".
// Если dubName начинается со слова озвучка или субтитры, то lowerDubName - 
// dubName с маленькой буквы, в противном случае lowerDubName = dubName.
func VideoTitle(current int, dubName, title string) string {
    // Первый символ к нижнему регистру и убрать пробелы
    r, n := utf8.DecodeRuneInString(strings.TrimSpace(dubName))
    lowerDubName := string(unicode.ToLower(r)) + dubName[n:]

    // Заменить при необходимости
    if strings.HasPrefix(lowerDubName, "озвучка") || strings.HasPrefix(lowerDubName, "субтитры") {
        dubName = lowerDubName
    }

    return fmt.Sprintf("Серия %d, %s. %s.", current, dubName, title)
}

func wrapSeries(title string, available, total int) string {
	if total == -1 {
		return fmt.Sprintf("%s (%d из ??? серий)", title, available)
	}

	if available == total {
		if total == 1 {
			return fmt.Sprintf("%s (%d серия)", title, total)
		}
		if total < 5 {
			return fmt.Sprintf("%s (%d серии)", title, total)
		}
		return fmt.Sprintf("%s (%d серий)", title, total)
	}

	return fmt.Sprintf("%s (%d из %d серий)", title, available, total)
}

func wrapAnimeTitleAired(anime models.Anime) string {
	if anime.MediaType == "фильм" {
		return fmt.Sprintf("%s (фильм)", anime.Title)
	}

	return wrapSeries(anime.Title, anime.EpCtx.AiredEpCount, anime.EpCtx.TotalEpCount)
}

func wrapAnimeTitleWatched(anime models.Anime) string {
	if anime.Provider == "" {
		return fmt.Sprintf("%s (не доступно, нажмите, чтобы найти в другом источнике)", anime.Title)
	}
	if anime.MediaType == "фильм" {
		return fmt.Sprintf("%s (фильм)", anime.Title)
	}

	return wrapSeries(anime.Title, anime.EpCtx.Cur, anime.EpCtx.AiredEpCount)
}
