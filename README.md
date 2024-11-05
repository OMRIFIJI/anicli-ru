# anicli-ru
Интерфейс командной строки для поиска и просмотра аниме, написаный на Go.

<img src="https://go.dev/blog/gopher/header.jpg" alt="Gopher" width="300"/>

## Зависимости
Видео проигрываются с помощью [mpv](https://github.com/mpv-player/mpv).

## Что умеет?
На данный момент представляет из себя минимальный вариант.
Находит и достаёт видео с помощью 
[Anilibria API](https://github.com/anilibria/docs/blob/master/api_v3.md).

В ближайших планах:
- [ ] Сохранять историю просмотра
- [ ] Опция для продолжения просмотра
- [ ] Добавить поддержку VLC
- [ ] Больше источников для просмотра
- [ ] Подправить костыли в коде

## Сборка
1. `git clone https://github.com/OMRIFIJI/anicli-ru.git`
2. `cd anicli-ru`
3. `go build -ldflags "-s -w" -o anicli-ru ./cmd/main.go`
4. По желанию копируем в одну из директорий **PATH**.\
Например: `sudo cp anicli-ru /usr/bin`
