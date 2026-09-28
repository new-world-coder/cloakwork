---
title: "Install the native host for Chrome Canary and Chromium"
labels:
  - help wanted
  - good first issue
---

`cloakwork install-host` writes the manifest only for stable Google Chrome:

`~/Library/Application Support/Google/Chrome/NativeMessagingHosts/`

People who develop the extension on Chrome Canary or Chromium never see the host. Teach the installer about those locations.

## Where to look

- `daemon/internal/install/install.go`
- `install/install-host-macos.sh`
- Chrome's per-user host directories on macOS:
  - `Google/Chrome`
  - `Google/Chrome Canary`
  - `Chromium`

## Done when

- By default, install still writes the stable Chrome manifest.
- A flag or subcommand argument selects `canary` or `chromium` and writes the same manifest into that channel's `NativeMessagingHosts` directory.
- A unit test uses `install.Options` and a temp directory. It must not write into the real home Library folder.
- The README install section lists the three channels.

## Out of scope

Linux and Windows host paths are a different task. Do not add a TCP listener.
