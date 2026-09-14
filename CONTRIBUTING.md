# Contributing to PinHarbor

Thanks for helping improve PinHarbor. The project welcomes focused bug fixes,
documentation improvements, and changes that preserve its local-first product
contract.

## Before you start

For a substantial change, open an issue first so the approach can be discussed.
Security vulnerabilities must be reported privately according to
[`SECURITY.md`](SECURITY.md).

By submitting a contribution, you confirm that you have the right to submit it
and agree that it may be distributed under the project's MIT License.

## Product contracts

A contribution must preserve these behaviors unless an approved migration says
otherwise:

- `Run PinHarbor.bat` is the supported Windows launcher.
- Auto Mode remains the completion-driven default.
- HTMX routes and existing form parameter names remain compatible.
- CSV columns remain `title,address,phone,complete_address,about,emails` in that
  exact order.
- `webdata/` is never deleted as a troubleshooting step.
- dark and light themes both remain readable.
- the upstream Go module path is not rewritten casually.

## Development workflow

1. Create a focused branch from `main`.
2. Make the smallest coherent change.
3. Format every changed Go file with `gofmt`.
4. Add or update tests for changed behavior.
5. Run the required checks.

```powershell
go test ./...
go vet ./...
git diff --check
```

For UI changes, run PinHarbor locally and verify both themes at desktop and
mobile widths. Do not include files from `webdata/`, `.env`, logs, credentials,
or generated binaries in a pull request.

## Pull requests

Explain the problem, the chosen solution, tests performed, and any user-visible
change. Keep unrelated formatting or dependency churn out of the same pull
request. Update documentation when behavior, configuration, or output changes.
