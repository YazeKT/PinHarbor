# Release and publication checklist

Use this checklist before publishing source or a distributable PinHarbor build.

## Source safety

- [ ] `.env`, `webdata/`, CSV exports, databases, browser profiles, logs, and
      local binaries are absent from the Git index.
- [ ] A repository-wide secret scan finds no real credentials.
- [ ] Upstream copyright, license, and contributor history remain intact.
- [ ] New third-party assets and dependencies have documented origins and
      compatible license terms.
- [ ] Documentation matches the behavior being published.

## Verification

```powershell
go test ./...
go vet ./...
git diff --check
```

- [ ] Changed Go files are formatted with `gofmt`.
- [ ] `Run PinHarbor.bat` reaches a healthy `/health` endpoint.
- [ ] Auto Mode is selected and remains completion-driven.
- [ ] Dark and light themes are readable.
- [ ] Stop preserves partial results; reset does not delete data.
- [ ] CSV headers match the exact six-column contract.

## GitHub evidence

- [ ] The pushed commit matches the locally verified commit.
- [ ] CI and CodeQL results are checked rather than assumed.
- [ ] GitHub detects the MIT license and renders the README assets.
- [ ] Repository description, topics, issue forms, and security reporting work.
- [ ] Release notes disclose known limitations and any unsigned executable.

## Binary distribution

Source publication does not automatically clear a binary for distribution.
Before attaching an executable, generate a dependency-license inventory,
include required notices, scan the artifact, compute a SHA-256 checksum, test a
clean-machine install or portable launch, and disclose the code-signing status.
