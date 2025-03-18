# anicli-ru

Интерфейс командной строки для поиска и просмотра аниме, полностью написаный на Go.

<img src="https://go.dev/blog/gopher/header.jpg" alt="Gopher" width="300"/>

## Что делает?
https://github.com/user-attachments/assets/e498ea62-478f-4bc0-b496-847f985f3220

К лучшему это или к худшему, но cli графика была полностью написана с нуля, без использования фреймворков.
В будущем постараюсь добавить функционала и вынести в отдельный модуль, доступный для всех.

Ищет аниме через [yummyanime](https://yummy-anime.ru/) и [animego](https://animego.org/), воспроизводит видео через mpv. 
Прямые ссылки на видео для mpv запрашиваются у встроенных в сайт плееров.
Ссылки на видео группируются по озвучке и качеству видео, 
в каждой группе выбирается видео плеера высшего приоритета.
Постарался раздать приоритет плеерам в соответствии с их стабильностью.

Плеер sibnet и anilib могут игнорировать http запросы по несколько раз
прежде чем ответить, поэтому, если вы наткнетесь на тот редкий случай, когда видео 
доступно только в этом плеере, то видео может открываться очень долго.

Так как домены плееров и аниме сайтов часто банят, я реализовал автоматическое отбрасывание недоступных доменов. 
По этой причине первый запуск приложения может занять около 2-3 секунд.
Подробнее о том, какие домены используются для парсинга и, как настроить их самостоятельно, в разделе 
[конфиг](https://github.com/OMRIFIJI/anicli-ru?tab=readme-ov-file#%D0%BA%D0%BE%D0%BD%D1%84%D0%B8%D0%B3).

Дополнительный функционал anicli-ru:
```
--check-providers
    проверить, какие источники из конфига доступны
-c, --continue
    продолжить просмотр аниме
-d, --delete
    удалить запись из базы данных, просматриваемых аниме
--delete-all
    удалить все записи из базы данных, просматриваемых аниме
```

## Установка
Скачайте готовый бинарник из последнего [релиза](https://github.com/OMRIFIJI/anicli-ru/releases) или установите с помощью go.
```
go install github.com/OMRIFIJI/anicli-ru/cmd/anicli-ru@latest
```

Для воспроизведения видео используется плеер `mpv` и `ffmpeg`.

## Конфиг
Программа создаст стандартный конфиг по одному из следующих путей:
1. `$XDG_CONFIG_HOME/anicli-ru/config.toml`
2. `$HOME/anicli-ru/config.toml`

Стандартный конфиг будет иметь примерно следующий вид:
```
[Video]
dub = ''
quality = 1080

[Providers]
autoSync = true

[Providers.domainMap]
animego = 'animego.one'
yummyanime = 'yummy-anime.ru'
anilib = 'api2.mangalib.me'

[Players]
syncInterval = '3d'
domains = [
  'sovetromantica.com',
  'kodik.info',
  'vk.com',
  'aniboom.one',
  'alloha.yani.tv',
  'aksor.yani.tv',
  'video.sibnet.ru',
  'video1.anilib.me'
]
```

### Video

* `dub` - предпочитаемая озвучка. Регистр не важен, лучше писать полное название озвучки, например, "AniLibria" или "anidub".
* `quality` - предпочитаемое качество. Если выбранное качество недоступно, программа выберет ближайшее доступное к выбранному.

### Providers

* `autoSync` - автоматическая синхронизация источников аниме. 
Если включить, то один раз в день при запуске будет синхронизировать зеркала с моим файлом на 
[gist](https://gist.github.com/OMRIFIJI/aacb12102b3aff21c37d5273f2b76fa0).

### Providers.domainMap

* Пара значений `"название источника" = "доменное имя источника"`. 
Если, например, источник не доступен в вашем регионе, можете отключить его, убрав соответствующую ему строчку,
или подставить доменное имя доступного зеркала. В этом случае не забудьте отключить синхронизацию.

### Players
* `syncInterval` - интервал в днях, с которым программа будет проверять доступность плееров. 
Процесс синхронизации списка плееров занимает 2 секунды.
Чтобы синхронизация происходила вместе с каждым запуском приложения установите значение `"0d"`. 
Если хотите регулировать список плееров самостоятельно, уберите эту строчку из конфига.
* `domains` - список используемых доменов плееров. Можно убирать плееры из списка, 
добавлять новые не надо. Чтобы изменения не были стёрты в ходе синхронизации,
не забудьте убрать `syncInterval` из конфига.

## Windows
* Не забудьте добавить `mpv` в PATH.
* Рекомендую использовать `powershell`.
* Через `cmd` программа тоже будет работать, но новый [буфер экрана](https://learn.microsoft.com/ru-ru/windows/console/console-screen-buffers) она создавать не будет, вместо этого она будет захламлять ваш буфер.
* Конфиг находится в `%LOCALAPPDATA%\anicli-ru\config.toml`.
* Лог находится в `%LOCALAPPDATA%\anicli-ru\log.txt`.


## Проблемы и способы их решения
Если какие-то из источников не доступны в вашем регионе, то программа будет послушно ждать таймаут на
каждом запросе к этим источникам. В этом случае уберите недоступные источники из конфига
или замените их зеркалами.

Если встретитесь с какими-то другими проблемами, связанными с приложением, то можете заглянуть в лог.
Лог лежит в `$XDG_CONFIG_HOME/anicli-ru/log.txt`.

При сборке `ffmpeg` из исходников необходима зависимость `libxml2` для поддержки DASH.
