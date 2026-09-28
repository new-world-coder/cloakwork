---
title: "Add cloakwork status --json"
labels:
  - help wanted
  - good first issue
---

`cloakwork status` prints a fixed text block. Tools wrapping the CLI need the same data as JSON, including when the daemon is stopped.

## Where to look

- `daemon/internal/service/service.go` defines `Status` and `FormatStatus`.
- `daemon/cmd/cloakwork/main.go` implements `cmdStatus`.

## Done when

- `cloakwork status --json` prints one JSON object with `version`, `daemonRunning`, `socketPath`, `profileId`, and `host`.
- The text output stays the default.
- When the daemon is stopped, the JSON still prints and the process still exits 1.
- A test covers both the running and stopped shapes. Use the unix socket the way `sock_test` does, or test a extracted formatter with a fake status.

## Out of scope

Do not change the native messaging `getStatus` response shape.
