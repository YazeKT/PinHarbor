# Architecture

## Product boundary

PinHarbor is a local Go application with an embedded HTMX interface. The supported build stores job metadata in SQLite and results in CSV files under `webdata/`.

The upstream module path `github.com/gosom/google-maps-scraper` remains unchanged to avoid a high-risk import rewrite.

## Runtime flow

```text
Browser form
    │ POST /scrape
    ▼
web handlers ──► preset policy ──► SQLite job queue
                                      │
                                      ▼
runner/webrunner ──► gmaps jobs ──► resilient CSV writer
        │                 │                  │
        │                 └─ exact work counters
        │                                    │
        └──────── web progress service ◄─────┘
                          │
                          ▼
                 HTMX refresh every 2s
```

## Package responsibilities

- `gmaps/`: Google Maps browser/search jobs, parsing, scrolling, and enrichment.
- `exiter/`: thread-safe discovered/completed work counters and completion cancellation.
- `runner/`: shared CLI configuration and execution-mode selection.
- `runner/webrunner/`: sequential local job worker, browser setup, context lifetime, and resilient CSV output.
- `web/`: HTTP handlers, SQLite-backed service, preset policy, local-time presentation, realtime progress, live CSV previews, and safe deletion.
- `web/static/`: embedded templates, CSS themes, application JavaScript, OpenAPI spec, and PinHarbor assets.
- `scripts/windows/`: maintained persistent Windows launcher.

## Filesystem ownership

The main Go process owns SQLite and CSV files. Browser code communicates only through narrow HTTP/HTMX routes and never receives filesystem paths.

## Job lifecycle

`pending → working → ok`

An operator stop changes `pending` or `working` to `stopped` after the active worker has released the CSV. Flushed rows remain downloadable. Deleting a job is a separate destructive action that also removes its CSV.

A startup recovery pass changes abandoned `working` rows to `failed`. Existing CSV rows remain downloadable as a partial export.

## Realtime accuracy

- Result count: rows successfully readable from the flushed CSV.
- Search/place counters: live snapshots from `exiter`.
- Progress: completed work divided by currently discovered work.
- Speed: flushed rows divided by elapsed minutes.
- ETA: extrapolated from completed/discovered work and recalculated every refresh.

Because place work is discovered while searches run, early progress and ETA may move as the denominator grows. The interface labels ETA as estimated.
