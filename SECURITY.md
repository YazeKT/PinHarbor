# Security policy

## Supported versions

PinHarbor is currently maintained on the `main` branch. Security fixes are not
backported to older snapshots unless a release note explicitly says otherwise.

## Local security model

PinHarbor is intended for one operator on a trusted computer.

- The local server listens on port `8080` and does not provide authentication.
- Do not port-forward, tunnel, or publicly expose the server.
- Job metadata and results live under `webdata/`.
- Proxy URLs may contain credentials and must not be committed or shared.
- `.env`, databases, CSV results, browser profiles, and logs must stay out of
  source control.
- Browser code has no direct filesystem access.
- Job identifiers are validated before result files are opened or removed.
- Removing an active job first cancels and waits for its worker.

## External requests

Collection contacts Google Maps. Email discovery also visits public business
websites. Proxy use sends traffic and credentials to the configured provider.

## Report a vulnerability

Do not open a public issue for a security vulnerability or attach private CSV
data, search terms, proxy credentials, environment files, or browser profiles.

Use GitHub's **Report a vulnerability** option in the Security tab to send the
maintainer a private report. Include:

- the affected commit or version;
- clear reproduction steps;
- the expected security impact;
- a minimal proof of concept with secrets and personal data removed;
- any mitigation or fix you have already identified.

For a vulnerability that also affects the unchanged upstream engine, report it
to the upstream project as well. Please allow reasonable time for investigation
and coordinated disclosure.
