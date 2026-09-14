# Third-party notices

PinHarbor includes and depends on third-party open-source software. Those
components remain governed by their own copyright notices and license terms.

## Upstream foundation

- **Project:** `gosom/google-maps-scraper`
- **Source:** <https://github.com/gosom/google-maps-scraper>
- **License:** MIT
- **Copyright:** Copyright (c) 2023 Georgios Komninos

The upstream MIT notice is preserved in [`LICENSE`](LICENSE). Git history also
retains the attribution of upstream contributors.

## Go dependencies

Direct and transitive Go dependencies are declared in [`go.mod`](go.mod) and
locked by [`go.sum`](go.sum). Each package's license applies to that package.
Before distributing a compiled binary, review the license files for the exact
resolved dependency versions and include any notices those licenses require.

## Browser runtime and hosted assets

PinHarbor uses Playwright tooling and a Chromium browser runtime. The user
interface currently loads HTMX from cdnjs. Their respective licenses and terms
apply. A downstream distributor is responsible for the notices required by the
exact runtime and assets it ships.

This file is an attribution aid, not a complete legal opinion or a replacement
for an automated dependency-license review before binary distribution.
