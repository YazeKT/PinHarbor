# Auto Mode and presets

## Auto complete (default)

Auto Mode is completion-driven:

- `run_mode=auto`
- `depth=0`, interpreted as scroll until the result feed stops growing.
- No maximum clock duration.
- Standard browser mode.
- Email discovery enabled.
- Extra reviews disabled.
- The worker exits when every search seed and discovered place/enrichment task completes.
- The scraper retains its three-minute inactivity guard for stalled browser work.

This is the recommended default when completeness matters more than a fixed finish time.

## Timed presets

| Mode | Maximum | Depth | Email | Extra reviews |
| --- | ---: | ---: | --- | --- |
| `quick` | 15m | 25 | No | No |
| `balanced` | 1h | 100 | Yes | No |
| `deep` | 3h | 250 | Yes | Yes |

## Custom mode

`custom` accepts manual depth, Fast Mode, email, extra-review, and maximum-time values. The maximum is mandatory and must be at least three minutes.

## Server authority

The browser displays preset values, but `web.ApplyRunPreset` is authoritative. This prevents disabled inputs, stale tabs, or direct form edits from changing preset guarantees.

## API compatibility

API requests without `run_mode` use:

- `custom` when a positive `max_time` is supplied (legacy behavior).
- `auto` when no maximum is supplied.

Existing stored timed jobs are presented as legacy Custom runs.
