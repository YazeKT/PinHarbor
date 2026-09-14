# Testing

## Required automated checks

```powershell
gofmt -w <changed-go-files>
go test ./...
go vet ./...
git diff --check
```

`make test` is the upstream convenience target, but `make` is not installed on the supported Windows workstation. The equivalent Go commands are authoritative here.

The current Go toolchain reports `windows/386`, where `go test -race` is unsupported.

## Critical regression coverage

- Six-column CSV contract.
- Nil enrichment results do not crash the CSV writer.
- Auto preset has no duration/depth cap and enables email discovery.
- Quick/Balanced/Deep values are server-authoritative.
- Legacy timed jobs remain Custom.
- Live work snapshots are race-safe.
- Local-time job display and completed progress.
- Interrupted jobs become failed.
- Delete cancels the worker before removing its files.

## Browser verification

For each UI release verify:

1. Default theme is dark.
2. Every label, input, status, metric, and action is readable.
3. Theme toggle produces a readable light theme.
4. Auto is selected and preset-controlled fields are clearly locked.
5. Opening Custom enables manual controls.
6. Jobs refresh every two seconds.
7. Desktop and mobile widths do not crop actions or metrics.
8. `/health` remains available after an idle interval.
