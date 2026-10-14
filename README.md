# Tetris Client

## Одиночная игра

Игра без сервера использует ту же логику фигур, столкновений, очистки строк
и подсчёта очков, что и сетевой режим. Общий движок находится в `lib/field`,
одиночная сессия — в `lib/single`. Сетевой режим запускается без аргументов.

```sh
cd client
go build -o tetrisClient .
./tetrisClient --single
```

Управление: `a`/`d` — влево/вправо, `s` — вниз, `q`/`e` — поворот,
`x`, Escape или Ctrl+C — выход. Поле рисуется ASCII-символами `#`.
Ввод идёт через `go-tty`, без ожидания Enter; режим терминала восстанавливается
при выходе. Нужен терминал с ANSI-командами перемещения курсора.

Следующая фигура показана рисунком справа от поля. Экран обновляется не чаще
30 раз в секунду и только изменившимися участками; неподвижная фигура у стены
не вызывает перерисовку при автоповторе клавиши. Ввод и таймер падения работают
независимо от частоты кадров. Внешняя очистка запускается только при входе в игру.

Очистка экрана остаётся внешней командой: `clear` в Linux/macOS,
`cmd /c cls` в Windows. Для OSel нужна Linux-сборка с `CGO_ENABLED=0`,
`/dev/tty` и исполняемая `/bin/clear` в `PATH`. Отсутствие команды очистки
даёт понятную ошибку при запуске одиночной игры.

Для воспроизводимого сквозного теста: `--single --seed 2 --trace`.
`--seed` фиксирует последовательность фигур; `--trace` печатает события
`TETRIS READY`, `TETRIS INPUT key=a changed=true`, `TETRIS TICK`, `TETRIS EXIT`.
`changed=true` означает реальное изменение состояния игры, включая форму фигуры
при повороте. Seed 2 начинает с L-фигуры, чтобы сквозной тест проверял поворот.
Тест передаёт настоящие нажатия в терминал; таймер продолжает двигать фигуры.

Проверки: `go test ./...` отдельно в `lib`, `client` и `server`.

## Сетевая игра

Сервер можно выбрать без пересборки клиента:

```sh
./tetrisClient --network --server 192.168.1.10:8080
./tetrisClient --network --server 192.168.1.10:8080 --create
./tetrisClient --network --server 192.168.1.10:8080 --session 123456789
```

Без аргументов открывается сетевое меню с прежним адресом сервера. В меню:
`c` — создать, `r` — обновить, `w`/`s` — выбрать, Enter — подключиться,
`x` или `e` — выйти. В игре `x`, Escape или Ctrl+C возвращают управление.
Оба поля рисуются ASCII, помещаются в терминале шириной 80 символов и
обновляются до 30 кадров/с. Недоступный сервер даёт ошибку вместо паники;
сетевые запросы и подключение ограничены пятью секундами. Протокол — HTTP
и WebSocket без TLS. `--trace` выводит подключение, поля, ввод, pong и выход.

Локальный сервер из этого репозитория:

```sh
cd server
go run . -addr 0.0.0.0:8080
```

В OSel: `tetris --network --server IP:8080`. Одиночный запуск `tetris`
сохраняет прежнее поведение. Ядро должно поддерживать TCP и системные вызовы
сокетов Linux. Для QEMU локальный сервер хоста доступен как `10.0.2.2`.

## Demo

### Scoring tetris:
![Tetris.gif](./gifs/Tetris.gif)

### Full Game:
![DrawGame.gif](./gifs/Game.gif)

## Guide

### Controls
Controls are realised by buttons:
- a = move piece left
- s = move piece down
- d = move piece right
- q = rotate piece left
- e = rotate piece right

### Launch
Show help menu
```
./tetrisClient help
```

Launch menu
```
./tetrisClient
```

Create new session:
```
./tetrisClient --network --create
```

Connect to existing session:
```
./tetrisClient --network --session 123456789
```

The network menu lists active sessions; press `r` to refresh.
