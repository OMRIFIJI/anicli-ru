# anicli-ru

Интерфейс командной строки для поиска и просмотра аниме, написаный на Go.

<img src="https://go.dev/blog/gopher/header.jpg" alt="Gopher" width="300"/>

## Что умеет?
https://github.com/user-attachments/assets/e498ea62-478f-4bc0-b496-847f985f3220

Поиск аниме осуществляется парсингом [animego](https://animego.org/).
Прямые ссылки на видео для mpv запрашиваются у встроенных в animego плееров.
Некоторые плееры, в частности aniboom, предлагают DASH.
Для работы с ними mpv может использовать ffmpeg.
Ссылки на видео группируются по озвучке и качеству видео, 
в каждой группе выбирается ссылка приоритетного плеера.

Приоритет плееров задан следующим образом:
1. aniboom
2. kodik
3. vk
4. sibnet

Плеер sibnet может игнорировать http запросы по несколько раз
прежде чем ответить, поэтому видео от этого плеера в среднем будут запускаться 
намного медленее остальных.

В ближайших планах:
- [x] Динамическая отрисовка графики
- [x] Собственный парсинг аниме
- [ ] Опция для продолжения просмотра
- [ ] Искать аниме не только на animego
- [ ] Добавить поддержку VLC
- [ ] Сделать код читабельнее

## Установка
Просто скачайте готовый бинарник из последнего [релиза](https://github.com/OMRIFIJI/anicli-ru/releases). Для воспроизведения видео используется плеер `mpv` и `ffmpeg`.

## Проблемы и способы их решения
При сборке `ffmpeg` из исходников важно добавить зависимость `libxml2` для поддержки DASH.
