#!/bin/sh
# Stop/SubagentStop: не даёт агенту закончить ход, пока go vet или тесты красные.
input=$(cat)

# Повторный заход после блокировки — отпускаем, чтобы не зациклиться.
[ "$(printf '%s' "$input" | jq -r '.stop_hook_active // false')" = "true" ] && exit 0

cd "${CLAUDE_PROJECT_DIR:-.}" || exit 0

# Go-файлы не менялись — проверять нечего.
[ -z "$(git status --porcelain -- '*.go' go.mod go.sum)" ] && exit 0

if ! out=$(go vet ./... 2>&1 && go test -race ./... 2>&1); then
	printf 'go vet / go test -race упали, исправь перед завершением:\n%s\n' "$(printf '%s' "$out" | tail -n 60)" >&2
	exit 2
fi
