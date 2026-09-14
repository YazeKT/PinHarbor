# Local API

Base URL: `http://localhost:8080`

## Health

`GET /health`

Returns the application name, readiness, and current local server time.

## Jobs

- `POST /api/v1/jobs`: create a job.
- `GET /api/v1/jobs`: list jobs.
- `GET /api/v1/jobs/{id}`: inspect a job.
- `POST /api/v1/jobs/{id}/stop`: stop collection and preserve its job and partial CSV.
- `DELETE /api/v1/jobs/{id}`: cancel/delete a job and its CSV.
- `GET /api/v1/jobs/{id}/download`: download its CSV.

Interactive schema documentation is available at `/api/docs`.

## Auto Mode example

```powershell
$body = @{
  name = "Eastern Cape dentists"
  run_mode = "auto"
  keywords = @("dentists in East London", "dentists in Gqeberha")
  lang = "en"
  zoom = 15
  radius = 10000
} | ConvertTo-Json

Invoke-RestMethod -Method Post -Uri http://localhost:8080/api/v1/jobs -ContentType application/json -Body $body
```

For API durations, `max_time` remains seconds for compatibility. It is ignored by managed presets and required for `custom`.

Job statuses are `pending`, `working`, `ok`, `failed`, and `stopped`.

## CSV output

`title,address,phone,complete_address,about,emails`
