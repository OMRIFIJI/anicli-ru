package video

import (
	"context"
	"errors"
	"fmt"

	"github.com/OMRIFIJI/anicli-ru/internal/animeapi"
	"github.com/OMRIFIJI/anicli-ru/internal/animeapi/models"
	"github.com/OMRIFIJI/anicli-ru/internal/animefmt"
	"github.com/OMRIFIJI/anicli-ru/internal/app/config"
)

type AnimePlayer struct {
	api         *animeapi.AnimeAPI
	anime       *models.Anime
	selector    *videoSelector
	player      *videoPlayer
	noDubErr    *noDubError
	replayVideo bool
}

func NewAnimePlayer(anime *models.Anime, api *animeapi.AnimeAPI, cfg *config.VideoCfg) *AnimePlayer {
	ap := &AnimePlayer{
		anime:    anime,
		api:      api,
		selector: newSelector(),
		player:   newVideoPlayer(cfg),
		noDubErr: &noDubError{},
	}

	return ap
}

func (ap *AnimePlayer) Play() error {
	err := ap.updateLink()
	if err != nil {
		// Обычная ошибка
		if !errors.As(err, &ap.noDubErr) {
			return err
		}

		// Не нашлась озвучка из конфига
		promptMessage := "Выберите озвучку. " + ap.anime.Title
		isExitOnQuit, err := ap.selector.selectDub(promptMessage, ap.player)
		if err != nil {
			return err
		}
		if isExitOnQuit {
			return nil
		}
	}

	if err := ap.spin(); err != nil {
		return err
	}

	return nil
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

	videos, err := ap.api.Converter.Convert(ep.EmbedLinks)
	if err != nil {
		return err
	}

	err = ap.player.SetVideos(videos)
	return err
}

func (ap *AnimePlayer) startMpvWrapped(ctx context.Context) error {
	videoTitle := animefmt.VideoTitle(ap.anime.EpCtx.Cur, ap.player.ResolvedDub, ap.anime.Title)
	err := ap.player.StartMpv(videoTitle, ctx)
	return err
}

func (ap *AnimePlayer) spin() error {
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
	promptMessage := animefmt.PlayerMenuHeader(ap.anime)
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
	return false, errors.New("опцию меню не удалось обработать")
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
			return false, errors.New("выход из меню обязательного выбора новой озвучки")
		}
		return true, nil
	}

	if err != nil {
		return false, err
	}
	return true, nil
}

func (ap *AnimePlayer) handleChangeDub() (replayVideo bool, err error) {
	promptMessage := fmt.Sprintf("Выберите новую озвучку. Сейчас выбрана %s. %s.", ap.player.cfg.Dub, ap.anime.Title)
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
	promptMessage := fmt.Sprintf("Выберите новое качество видео. Сейчас выбрано %d. %s.", ap.player.cfg.Quality, ap.anime.Title)
	isExitOnQuit, err := ap.selector.selectQuality(promptMessage, ap.player)
	if err != nil {
		return false, err
	}
	if isExitOnQuit {
		return false, nil
	}
	return true, nil
}
