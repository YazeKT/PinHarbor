# Operations

## Supported Windows start

Double-click `Run PinHarbor.bat`.

It delegates to `scripts/windows/run-pinharbor.ps1`, which:

1. Reuses an already healthy PinHarbor instance.
2. Builds `bin/pinharbor.exe` from current source.
3. Installs Playwright Chromium once.
4. starts the local server.
5. Waits for `/health` before opening the browser.
6. Restarts the server three seconds after an unexpected exit.

Docker Desktop and WSL are not used.

## Keep localhost running

Keep the PinHarbor launcher window open. The supervisor restarts the Go process if it crashes. The HTTP server has no idle shutdown timer, so leaving the page unused does not stop localhost.

To stop PinHarbor intentionally, focus the launcher window and press `Ctrl+C`, or end `pinharbor.exe` in Task Manager.

## Health check

```powershell
Invoke-RestMethod http://127.0.0.1:8080/health
```

Expected fields:

```json
{"app":"PinHarbor","status":"ok","time":"2026-08-31T20:00:00+02:00"}
```

## Data

- `webdata/jobs.db`: SQLite job metadata.
- `webdata/<job-id>.csv`: result rows.
- `webdata/.playwright-ready`: first-run browser marker.

Back up `webdata/` only while PinHarbor is stopped, or use a SQLite-aware backup process.

## Recovery

### Port 8080 already in use

Check the listener:

```powershell
Get-NetTCPConnection -LocalPort 8080 -State Listen
```

If `/health` returns `app: PinHarbor`, use the existing instance. Do not start a duplicate worker against the same `webdata` folder.

### Interrupted job

On restart, an abandoned working job is marked failed. Any flushed CSV rows remain available through **Download partial**.

### Stop or reset

- **Stop & keep partial** cancels active browser work, waits for the CSV writer to close, marks the job `stopped`, and keeps every flushed row available for download.
- **Reset form** only restores the input form to Auto complete defaults. It never stops a running job and never deletes stored data.

### Browser engine problem

Delete only `webdata/.playwright-ready` and start PinHarbor again. The launcher will reinstall the Playwright browser.

### Build problem

Run:

```powershell
go build -o bin\pinharbor-check.exe .
go test ./...
```

Never delete `webdata/` as a troubleshooting shortcut; it contains user results.
