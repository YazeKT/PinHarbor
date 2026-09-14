# PinHarbor on macOS and Linux

Windows is the primary supported environment for this local fork. A direct Go supervisor is provided for macOS/Linux as a convenience.

## Requirements

- Go
- `curl`
- A desktop opener such as `xdg-open` on Linux

## Start

```bash
chmod +x start-host.sh
./start-host.sh
```

The script builds `bin/pinharbor`, installs Playwright browser assets on first use, starts `http://localhost:8080`, and restarts the process after an unexpected exit.

There is no idle shutdown timer. Keep the terminal open and press `Ctrl+C` to stop intentionally.

## Verify

```bash
curl http://127.0.0.1:8080/health
```

The maintained Windows procedure and recovery details are in [docs/OPERATIONS.md](docs/OPERATIONS.md).
