# jrpc

JSON-RPC 2.0 поверх Echo v5. Модуль `github.com/mrFokin/jrpc/v2`, Go 1.27. Ветка Echo v4 живёт в тегах `v1.1.x` — её не трогаем.

## Устройство

- `jrpc.go` — `Endpoint`, регистрация методов (`Handler`, `Method`), HTTP-обработчик, middleware, recover паники.
- `protocol.go` — разбор тела: batch/single, `Request`, `id` (string | number | null), лимит тела 1 MiB.
- `context.go` — `Context`: `Bind`, `Result`, `Request`, поле `Echo`.
- `error.go` — `JRPCError`, `NewError`, `NewErrorInvalidParams`, стандартные коды.
- `jrpc_test.go`, `example_test.go` — тесты и godoc-примеры.

## Правила

- Спецификация — https://www.jsonrpc.org/specification. Любое отклонение от неё — баг, если не описано в README как осознанное решение.
- Публичный API `/v2` не ломать. Если без этого никак — остановиться и спросить.
- Не добавлять то, что README явно исключает: `HandleMethod`, интерфейс `JRPC`, WebSocket, другие фреймворки.
- KISS: без новых пакетов, интерфейсов и обёрток, если задача их не требует. Новые зависимости — только с согласия.
- Тесты: testify (`assert`), HTTP через `httptest` + `e.ServeHTTP`, табличные где кейсов больше двух. Сравнивать тело ответа строкой JSON, как в существующих тестах.
- Каждое заметное изменение — строка в `CHANGELOG.md` в `[Unreleased]` (формат Keep a Changelog). Изменение поведения — ещё и README/godoc.
- Проверка перед завершением: `go vet ./... && go test -race ./...` (хук `Stop` запускает это сам).

## Git

- Не коммитить и не пушить без явного «да» от пользователя; одно «да» — один коммит.
- Сообщения коммитов — на русском, в повелительном наклонении, с точкой («Перейти на Echo v5 и путь модуля /v2.»).

## Агенты

- `spec-auditor` (Opus) — сверка кода со спецификацией, только чтение, выдаёт список расхождений.
- `go-tester` (Sonnet) — тесты, fuzz, бенчмарки; правит только `*_test.go`.
- Ревью диффа — встроенные `/code-review` и `/simplify`.
