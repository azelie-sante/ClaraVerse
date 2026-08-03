# Changelog

All notable changes to Clara Agent are documented here.

Clara Agent is a fork of [Pi](https://github.com/earendil-works/pi). This file
tracks Clara Agent's own releases only; for upstream history, see Pi's changelog.

## [0.1.0]

Initial release, forked from Pi 0.83.0.

### Added

- ClaraVerse account provider. `/login claraverse` runs a device-authorization
  flow against your ClaraVerse instance, and the agent then uses the model and
  provider credentials configured on that account.
- Config file at `~/.clara-agent/agent/claraverse.json` for the ClaraVerse
  server URL. Resolution order: `CLARAVERSE_URL` env var, the URL captured on
  the stored credential at login, this config file, then `http://localhost:3000`.
- `claracli` command, with `clara-agent` kept as an alias.
- `npm run setup` for a one-shot install: builds the packages the CLI needs and
  links the binary onto your PATH.

### Changed

- Rebranded from Pi: command name, config directory (`~/.clara-agent`),
  environment variable prefix, and provider attribution headers.

### Removed

- Install telemetry to the upstream project's servers.
- Self-update version checks against the upstream release channel; Clara Agent
  has no release channel of its own yet.
- Upstream-specific easter eggs and promotional startup content.
