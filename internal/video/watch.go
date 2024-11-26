package video

import (
	"anicliru/internal/api"
	"anicliru/internal/api/models"
	"anicliru/internal/api/player"
	"errors"
	"fmt"
)

type AnimePlayer struct {
	anime     *models.Anime
	converter *player.PlayerLinkConverter
	selector  *videoSelector
	player    *videoPlayer
	noDubErr  *noDubError
}

func NewAnimePlayer(anime *models.Anime, options ...func(*AnimePlayer)) *AnimePlayer {
	ap := &AnimePlayer{anime: anime}
	ap.converter = api.NewPlayerLinkConverter()
	ap.selector = newSelector()
	ap.player = newVideoPlayer()
	ap.noDubErr = &noDubError{}

	for _, o := range options {
		o(ap)
	}

	return ap
}

// Для продолжения просмотра
func WithVideoPlayerConfig(cfg videoPlayerConfig) func(*AnimePlayer) {
	return func(ap *AnimePlayer) {
		ap.player = &videoPlayer{cfg: cfg}
	}
}

func (ap *AnimePlayer) SpinWatch() error {
	err := ap.updateLink()
	if err != nil && !errors.As(err, &ap.noDubErr) {
		return err
	}

	if ap.player.cfg.isEmpty() {
		promptMessage := "Выберите озвучку. " + ap.anime.Title
		isExitOnQuit, err := ap.selector.selectDub(promptMessage, ap.player)
		if err != nil {
			return err
		}
		if isExitOnQuit {
			return nil
		}
	}

	// Просмотр первой выбранной серии
	if err = ap.wrappedStartMpv(); err != nil {
		return err
	}

	// Интерфейс для запуска остальных серий и изменения опций просмотра
	err = ap.spinWatchWithOptions()
	return err
}

func (ap *AnimePlayer) updateLink() error {
	ep, err := ap.anime.GetSelectedEp()
	if err != nil {
		return err
	}

	api.GetEmbedLink(ep)

	videoLink, err := ap.converter.GetVideoLink(ep.EmbedLink)
	if err != nil {
		return err
	}

	err = ap.player.SetLinks(videoLink)
	return err
}

func (ap *AnimePlayer) wrappedStartMpv() error {
	videoTitle := fmt.Sprintf("Серия %d. %s.", ap.anime.GetSelectedEpKey(), ap.anime.Title)
	err := ap.player.StartMpv(videoTitle)
	return err
}

func (ap *AnimePlayer) spinWatchWithOptions() error {
	for {
		promptMessage := fmt.Sprintf("Серия %d. %s.", ap.anime.GetSelectedEpKey(), ap.anime.Title)
		menuOption, isExitOnQuit, err := ap.selector.selectMenuOption(promptMessage)
		if err != nil {
			return err
		}
		if isExitOnQuit {
			return nil
		}

		if menuOption == exitPlayer {
			return nil
		}

		replayVideo, err := ap.handleVideoChange(menuOption)
		if err != nil {
			return err
		}
		if replayVideo {
			ap.wrappedStartMpv()
		}

	}
}

func (ap *AnimePlayer) handleVideoChange(menuOption string) (replayVideo bool, err error) {
	switch menuOption {
	case nextEpisode, previousEpisode:
		return ap.handleEpisodeSwitch(menuOption)
	case replay:
		err = ap.player.KillMpv()
		if err != nil {
			return false, err
		}
		return true, nil
	case changeDub:
		return ap.handleChangeDub()
	case changeQuality:
		return ap.handleChangeQuality()
	}
	return false, errors.New("Опцию меню не удалось обработать")
}

func (ap *AnimePlayer) handleEpisodeSwitch(menuOption string) (replayVideo bool, err error) {
	switch menuOption {
	case nextEpisode:
		if err := ap.anime.SelectNextEp(); err != nil {
			return false, err
		}
	case previousEpisode:
		if err := ap.anime.SelectPreviousEp(); err != nil {
			return false, nil
		}
	}

	err = ap.updateLink()
	if errors.As(err, &ap.noDubErr) {
		promptMessage := ap.noDubErr.Error() + ap.anime.Title
		isExitOnQuit, err := ap.selector.selectDub(promptMessage, ap.player)
		if err != nil {
			return false, err
		}
		if isExitOnQuit {
			return false, errors.New("Не была выбрана новая озвучка.")
		}
		return true, nil
	}

	if err != nil {
		return false, err
	}
	return true, nil
}

func (ap *AnimePlayer) handleChangeDub() (replayVideo bool, err error) {
	promptMessage := fmt.Sprintf("Выберите новую озвучку. Сейчас выбрана %s. %s.", ap.player.cfg.CurrentDub, ap.anime.Title)
	isExitOnQuit, err := ap.selector.selectDub(promptMessage, ap.player)
	if err != nil {
		return false, err
	}
	if isExitOnQuit {
		return false, nil
	}
	return true, nil
}

func (ap *AnimePlayer) handleChangeQuality() (replayVideo bool, err error) {
	promptMessage := fmt.Sprintf("Выберите новое качество видео. Сейчас выбрано %d. %s.", ap.player.cfg.CurrentQuality, ap.anime.Title)
	isExitOnQuit, err := ap.selector.selectQuality(promptMessage, ap.player)
	if err != nil {
		return false, err
	}
	if isExitOnQuit {
		return false, nil
	}
	return true, nil
}
