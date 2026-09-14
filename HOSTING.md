# PinHarbor local hosting

The supported build is local-only and does not require Docker, WSL, PostgreSQL, or a public host.

Start it by double-clicking `Run PinHarbor.bat`.

For startup, health, persistence, data, backup, and recovery instructions, read [docs/OPERATIONS.md](docs/OPERATIONS.md).

Do not expose port 8080 to the internet without adding authentication, TLS, request limits, and a reviewed deployment configuration.
