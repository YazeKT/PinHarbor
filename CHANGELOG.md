# Changelog

All notable PinHarbor changes are documented here. The project follows the
structure of [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and plans
to use semantic versioning for public releases.

## [Unreleased]

### Added

- PinHarbor identity, logo, dark-first visual system, and readable light theme.
- Completion-driven Auto Mode plus Quick, Balanced, Deep, and Custom presets.
- Compact HTMX monitor with actual saved rows, work counters, speed, elapsed
  time, estimated completion, and recent results.
- Safe **Stop & keep partial** and non-destructive **Reset form** controls.
- Persistent Windows launcher with first-run browser setup, health checks, and
  automatic local-server recovery.
- Local documentation for architecture, operations, API use, testing, brand,
  security, licensing, attribution, and responsible use.

### Changed

- Focused CSV exports on the stable six-column PinHarbor contract.
- Hardened result writing and recovery so flushed partial results remain usable.
- Positioned the Windows-local workflow as the supported default.

### Security

- Kept local environment files, runtime databases, results, logs, and build
  output out of source control.
- Documented the unauthenticated localhost boundary and private reporting path.
