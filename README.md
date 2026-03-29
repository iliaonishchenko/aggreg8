# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

## Профилирование памяти (pprof)

Оптимизация: добавлен `sync.Pool` для переиспользования `gzip.Writer` и `gzip.Reader` в middleware сжатия (`internal/router/gzip.go`).

Результат сравнения профилей (`pprof -top -diff_base=profiles/base.pprof profiles/result.pprof`):

```
File: server
Type: alloc_space
Showing nodes accounting for -51.03MB, 83.26% of 61.29MB total
      flat  flat%   sum%        cum   cum%
  -17.54MB 28.61% 28.61%   -17.54MB 28.61%  compress/flate.(*dictDecoder).init (inline)
      -5MB  8.16% 36.77%       -5MB  8.16%  database/sql.driverArgsConnLocked
   -3.01MB  4.92% 41.69%   -20.55MB 33.53%  compress/flate.NewReader
   -2.64MB  4.31% 46.00%    -5.55MB  9.05%  compress/flate.NewWriter (inline)
   -2.33MB  3.81% 49.81%    -2.90MB  4.74%  compress/flate.(*compressor).init
   -1.51MB  2.46% 52.27%    -1.51MB  2.46%  bufio.NewReaderSize (inline)
   -1.50MB  2.45% 54.72%    -1.50MB  2.45%  sync.(*Pool).pinSlow
   -1.50MB  2.45% 57.17%       -3MB  4.90%  github.com/jackc/pgx/v5/pgconn.ParseConfigWithOptions
   -1.50MB  2.45% 59.62%    -1.50MB  2.45%  context.(*cancelCtx).Done
   -1.50MB  2.45% 62.07%    -1.50MB  2.45%  github.com/jackc/pgx/v5/pgconn.(*ResultReader).Read
   -1.16MB  1.89% 63.95%    -1.16MB  1.89%  runtime/pprof.StartCPUProfile
      -1MB  1.63% 65.58%       -1MB  1.63%  github.com/jackc/pgx/v5/internal/stmtcache.(*LRUCache).RemoveInvalidated
      -1MB  1.63% 67.22%       -1MB  1.63%  database/sql.(*Tx).grabConn
   -0.57MB  0.93% 65.30%    -0.57MB  0.93%  compress/flate.newDeflateFast (inline)
   -0.50MB  0.82% 66.12%    -0.50MB  0.82%  github.com/jackc/pgx/v5/internal/iobufpool.init.0.func1
   -0.50MB  0.82% 66.94%    -2.24MB  3.65%  runtime/pprof.(*profileBuilder).emitLocation
   -0.50MB  0.82% 67.75%    -0.50MB  0.82%  golang.org/x/text/internal/language.map.init.1
   -0.50MB  0.82% 68.57%    -0.50MB  0.82%  io.ReadAll
   -0.50MB  0.82% 69.39%    -0.50MB  0.82%  encoding/json.(*Decoder).refill
   -0.50MB  0.82% 70.20%       -1MB  1.64%  github.com/jackc/pgx/v5/pgconn.connectOne
   -0.50MB  0.82% 71.02%    -2.51MB  4.09%  github.com/jackc/pgx/v5.connect
   -0.50MB  0.82% 71.83%    -0.50MB  0.82%  maps.Copy[...]
   -0.50MB  0.82% 72.65%    -0.50MB  0.82%  github.com/jackc/pgx/v5/pgproto3.(*Describe).Encode
   -0.50MB  0.82% 73.47%    -0.50MB  0.82%  github.com/jackc/pgx/v5/pgconn.configTLS
   -0.50MB  0.82% 74.28%    -0.50MB  0.82%  net/http.(*Request).WithContext (inline)
   -0.50MB  0.82% 75.10%       -1MB  1.63%  net/http.readRequest
   -0.50MB  0.82% 75.91%    -0.50MB  0.82%  net/http.(*Request).SetPathValue (inline)
   -0.50MB  0.82% 76.73%       -1MB  1.63%  github.com/jackc/pgx/v5.(*Conn).Prepare
   -0.50MB  0.82% 77.55%    -0.50MB  0.82%  net.(*Dialer).DialContext
   -0.50MB  0.82% 78.36%    -0.50MB  0.82%  reflect.growslice
   -0.50MB  0.82% 79.18%    -0.50MB  0.82%  net.cgoLookupHostIP
   -0.50MB  0.82% 79.99%    -0.50MB  0.82%  github.com/go-chi/chi/v5.(*node).FindRoute
   -0.50MB  0.82% 80.81%       -4MB  6.53%  github.com/jackc/pgx/v5/stdlib.(*Conn).ExecContext
   -0.50MB  0.82% 81.63%    -0.50MB  0.82%  reflect.unsafe_New
   -0.50MB  0.82% 82.44%    -0.50MB  0.82%  github.com/jackc/pgx/v5/pgtype.NewMap
   -0.50MB  0.82% 83.26%   -16.51MB 26.93%  github.com/iliaonishchenko/aggreg8/internal/repository.(*MetricsRepository).BatchUpdate.func1
   -0.50MB  0.82% 83.26%    -0.50MB  0.82%  github.com/jackc/pgx/v5/pgconn.(*PgConn).makeCommandTag (inline)
   -0.50MB  0.82% 83.26%    -0.50MB  0.82%  github.com/jackc/pgx/v5/pgproto3.(*ParameterDescription).Decode
   -0.50MB  0.82% 83.26%    -0.50MB  0.82%  net/netip.Addr.AsSlice (inline)
   -0.50MB  0.82% 83.26%    -0.50MB  0.82%  strings.genSplit
```
