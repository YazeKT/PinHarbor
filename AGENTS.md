# PinHarbor development guide

## Product invariants

- PinHarbor is a Windows-local, single-operator application.
- The supported launcher is `Run PinHarbor.bat`; do not reintroduce Docker/WSL as the default path.
- Preserve HTMX routes and form parameter compatibility.
- Preserve the CSV order: `title,address,phone,complete_address,about,emails`.
- Auto Mode is the default and must remain completion-driven.
- Dark mode is the default; every UI change must be verified in dark and light themes.
- Never delete `webdata/` during troubleshooting.
- Keep the upstream Go module path unchanged unless an explicit migration is planned.

## Build and test

- `go test ./...`
- `go vet ./...`
- `gofmt -w <changed-go-files>`
- `git diff --check`
- `go test ./path/to/package` for focused work

`make test` is an upstream convenience target but `make` is not installed on this Windows host. Race detection is unavailable with the current `windows/386` Go toolchain.

## Code organization

- `gmaps/`: collection and parsing.
- `exiter/`: exact work completion counters.
- `runner/webrunner/`: local worker and resilient result writer.
- `web/presets.go`: preset policy.
- `web/progress.go`: realtime presentation.
- `web/static/`: PinHarbor UI and assets.
- `scripts/windows/`: persistent launcher.
- `docs/`: maintained current-build documentation.

## Go style

- Run `gofmt`.
- Standard library, third-party, then local imports.
- Wrap errors with `fmt.Errorf(... %w ...)`.
- Put `context.Context` first.
- Prefer early returns.
- Add godoc comments for exported types and functions.
- Avoid magic numbers and unexplained lint suppressions.
