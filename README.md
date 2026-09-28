# CloakWork

CloakWork is an open-source privacy suite. One coherent browser cohort is the source of truth for a desktop daemon, a Chrome extension, and a later automation SDK. The cohort mimics a common browser. It does not randomize itself per page.

This repository is Sprint 1: the Go daemon and the Chrome native messaging bridge.

CloakWork reduces cross-signal contradiction, such as a browser time zone that disagrees with the claimed locale. It does not defeat a determined tracker, and it is not a fraud or multi-account toolkit. See [CONTRIBUTING.md](CONTRIBUTING.md) for how that limit applies to contributions.

Licensed under the [MIT License](LICENSE).

## Sprint 1

The daemon speaks three messages: `ping`, `getStatus`, and `getProfile`. It does not change the system time zone and it does not inject anything into a page.

| Piece | Role |
| --- | --- |
| `cloakwork daemon` | Listens on a per-user unix socket. Another uid is rejected. |
| `cloakwork host` | Chrome's native messaging host. Stdio only. Chrome starts it. |
| `cloakwork status` | Reports whether the daemon is running, and which cohort is active. |
| `cloakwork profile` | Prints the shipped cohort. |
| `cloakwork install-host` | Writes the Chrome host manifest for this user. |

There is no localhost WebSocket. The extension is authenticated because Chrome launches the host and the manifest pins `allowed_origins`. The CLI and a future SDK use the unix socket. On macOS the daemon reads the peer uid with `LOCAL_PEERCRED` and drops any connection from another user.

## Build

Go 1.26 or newer:

```sh
make test
make build
./daemon/bin/cloakwork profile show
./daemon/bin/cloakwork status
```

`status` exits 1 until the daemon is running:

```sh
./daemon/bin/cloakwork daemon
```

The socket path is `$TMPDIR/cloakwork.sock`, mode `0600`. On macOS that directory is already private to your user.

## Install the Chrome host

```sh
make install-host
```

That writes:

- `~/Library/Application Support/CloakWork/cloakwork`
- `~/Library/Application Support/CloakWork/cloakwork-host`
- `~/Library/Application Support/Google/Chrome/NativeMessagingHosts/com.cloakwork.daemon.json`

The committed template is [`install/com.cloakwork.daemon.json`](install/com.cloakwork.daemon.json). `allowed_origins` is `chrome-extension://EXTENSION_ID_PLACEHOLDER/` until the extension exists. The manifest `path` must be the absolute path of the wrapper, which `install-host` fills in.

Native messaging frames are a 4-byte little-endian length plus JSON, up to 1 MiB. The host must not log to stdout.

## Profiles

Shipped cohorts live in `daemon/internal/profile/cohorts/`. A profile is rejected when its signals disagree: the user agent, client hints, WebGL renderer, fonts, locale, and time zone offset have to describe the same machine. The reference cohort is Windows 11, Chrome 131, `en-US`, `America/New_York`.

## Contributing

New contributors should start with a [help wanted / good first issue](docs/contributing/GOOD_FIRST_ISSUES.md). Those tasks are written up in the repo. Filing them on GitHub takes `gh auth login` and then `python3 scripts/file_github_issues.py`.

Please read [CONTRIBUTING.md](CONTRIBUTING.md) and the [Code of Conduct](CODE_OF_CONDUCT.md).
