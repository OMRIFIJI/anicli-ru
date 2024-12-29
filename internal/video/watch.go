package video

import (
	"anicliru/internal/api"
	"anicliru/internal/api/models"
	"anicliru/internal/api/player"
	"context"
	"errors"
	"fmt"
)

type AnimePlayer struct {
	api         *api.AnimeAPI
	anime       *models.Anime
	converter   *player.PlayerLinkConverter
	selector    *videoSelector
	player      *videoPlayer
	noDubErr    *noDubError
	replayVideo bool
}

func NewAnimePlayer(anime *models.Anime, apiPtr *api.AnimeAPI, options ...func(*AnimePlayer)) *AnimePlayer {
	ap := &AnimePlayer{
		anime:     anime,
		converter: api.NewPlayerLinkConverter(),
		api:       apiPtr,
		selector:  newSelector(),
		player:    newVideoPlayer(),
		noDubErr:  &noDubError{},
	}

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

func (ap *AnimePlayer) Play() error {
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

	err = ap.spinPlay()
	return err
}

func (ap *AnimePlayer) updateLink() error {
	ep, err := ap.anime.EpCtx.GetSelectedEp()
	if err != nil {
		return err
	}

	err = ap.api.SetEmbedLinks(ap.anime, ep)
	if err != nil {
		return err
	}

	videos, err := ap.converter.GetVideos(ep.EmbedLinks)
	if err != nil {
		return err
	}

	err = ap.player.SetVideos(videos)
	return err
}

func (ap *AnimePlayer) startMpvWrapped(ctx context.Context) error {
	videoTitle := fmt.Sprintf("Серия %d. %s.", ap.anime.EpCtx.Cur, ap.anime.Title)
	err := ap.player.StartMpv(videoTitle, ctx)
	return err
}

func (ap *AnimePlayer) spinPlay() error {
	backgroundCtx := context.Background()
	ctx, cancel := context.WithCancelCause(backgroundCtx)

	// Первое видео уже настроено
	ap.replayVideo = true
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				ap.showMpvAndMenu(ctx, cancel)
			}
		}
	}()

	select {
	case <-ctx.Done():
		err := context.Cause(ctx)
		if err == context.Canceled {
			return nil
		}
		return err
	}
}

func (ap *AnimePlayer) showMpvAndMenu(ctx context.Context, cancel context.CancelCauseFunc) {
	go func() {
		if ap.replayVideo {
			err := ap.startMpvWrapped(ctx)
			if err != nil {
				cancel(err)
			}
		}
	}()
	promptMessage := fmt.Sprintf("Серия %d. %s.", ap.anime.EpCtx.Cur, ap.anime.Title)
	menuOption, isExitOnQuit, err := ap.selector.selectMenuOption(promptMessage)
	if err != nil {
		cancel(err)
	}
	if isExitOnQuit {
		cancel(nil)
	}

	if menuOption == exitPlayer {
		cancel(nil)
	}

	ap.replayVideo, err = ap.handleVideoChange(menuOption)
	if err != nil {
		cancel(err)
	}
}

func (ap *AnimePlayer) handleVideoChange(menuOption string) (replayVideo bool, err error) {
	switch menuOption {
	case nextEpisode, previousEpisode:
		return ap.handleEpisodeSwitch(menuOption)
	case replay:
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
		if err := ap.anime.EpCtx.SelectNextEp(); err != nil {
			return false, err
		}
	case previousEpisode:
		if err := ap.anime.EpCtx.SelectPreviousEp(); err != nil {
			return false, nil
		}
	}

	err = ap.updateLink()
	if errors.As(err, &ap.noDubErr) {
		promptMessage := fmt.Sprintf("%s Выберите новую озвучку. %s.", ap.noDubErr.Error(), ap.anime.Title)
		isExitOnQuit, err := ap.selector.selectDub(promptMessage, ap.player)
		if err != nil {
			return false, err
		}
		if isExitOnQuit {
			return false, errors.New("Выход из меню обязательного выбора новой озвучки.")
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
