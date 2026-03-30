/*
Package aggreg8 реализует сервис сбора, хранения и отображения метрик.

Проект состоит из двух компонентов:

Сервер (cmd/server) принимает метрики по HTTP и хранит их в памяти или PostgreSQL.
Поддерживает эндпоинты:

  - POST /update/{type}/{name}/{value} — обновление метрики через URL.
  - POST /update — обновление метрики через JSON.
  - POST /updates — пакетное обновление метрик через JSON.
  - GET /value/{type}/{name} — получение значения метрики в текстовом формате.
  - POST /value — получение метрики в формате JSON.
  - GET / — отображение всех метрик в виде HTML-таблицы.
  - GET /ping — проверка доступности базы данных.

Агент (cmd/agent) периодически собирает runtime- и системные метрики
и отправляет их на сервер с поддержкой gzip-сжатия и HMAC-подписи.

Основные внутренние пакеты:

  - internal/handler — HTTP-обработчики эндпоинтов.
  - internal/service — интерфейс хранилища и бизнес-логика.
  - internal/service/memory — in-memory хранилище метрик.
  - internal/service/pg — хранилище на основе PostgreSQL.
  - internal/service/sync — обёртка с синхронной записью в файл.
  - internal/model — модели данных (Metrics, Counter, Gauge).
  - internal/repository — слой доступа к PostgreSQL.
  - internal/repository/file — файловое хранилище метрик.
  - internal/router — middleware (gzip, HMAC-подпись).
  - internal/logger — структурированное логирование.
  - internal/signature — HMAC-SHA256 подпись данных.
  - internal/audit — аудит событий обновления метрик.
  - internal/agent — агент сбора и отправки метрик.
  - internal/config — конфигурация сервера и агента.
*/
package aggreg8
