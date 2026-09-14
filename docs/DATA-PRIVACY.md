# Data and privacy

## Local storage

The supported PinHarbor workflow stores operator data under `webdata/`:

- `jobs.db` contains job metadata;
- `<job-id>.csv` files contain collected results;
- `.playwright-ready` records that browser setup has completed.

These files are excluded from Git. PinHarbor does not intentionally upload the
job database or finished CSV exports to a PinHarbor service.

## Network activity

Local-first does not mean offline. Collection requests Google Maps. Email
discovery visits public business websites, and configured proxies can observe
traffic and receive embedded proxy credentials. The HTMX library is currently
loaded from cdnjs when the interface opens.

## Retention and deletion

Keep results only as long as needed for a defined purpose. The **Delete** action
removes the selected job and its CSV; **Stop & keep partial** does not. **Reset
form** only clears unsent form values and never removes stored jobs or results.

Back up `webdata/` only while PinHarbor is stopped or with a SQLite-aware backup
process. Protect backups to the same standard as the live data.

## Safe sharing

Before sharing logs, screenshots, issues, or example exports, remove:

- search terms and private job names;
- personal names, direct phone numbers, and email addresses;
- proxy URLs, usernames, passwords, tokens, and environment values;
- job identifiers and local filesystem paths when they reveal private context.

Use synthetic examples in public bug reports whenever possible.

## Operator obligations

The operator decides why data is collected, how it is used, who receives it,
and when it is deleted. Review [`LEGAL.md`](../LEGAL.md) for responsible-use and
accuracy limitations. PinHarbor does not provide consent management, outreach
suppression lists, or jurisdiction-specific compliance automation.
