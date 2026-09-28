# Contributing to CloakWork

CloakWork is open source under the [MIT License](LICENSE). New contributors are welcome. The starter tasks are labeled **help wanted** and **good first issue**.

Read the [Code of Conduct](CODE_OF_CONDUCT.md) before you start.

## What this project is

CloakWork keeps one coherent browser cohort across a desktop daemon, a Chrome extension, and a later automation SDK. Sprint 1 is the Go daemon and the Chrome native messaging host. The extension UI, OS time zone helper, leak canary, and Playwright SDK are later sprints. Issues for those later pieces say so in the task.

CloakWork reduces cross-signal contradiction. It does not claim to be undetectable, and it is not a toolkit for fraud or multi-account abuse.

## Setup

Go 1.26 or newer is required. Sprint 1 is developed on macOS.

```sh
make test
make build
./daemon/bin/cloakwork status
./daemon/bin/cloakwork profile show
```

`status` exits 1 until `cloakwork daemon` is running. That is expected.

`make install-host` registers the native host for the current user at:

`~/Library/Application Support/Google/Chrome/NativeMessagingHosts/com.cloakwork.daemon.json`

The manifest ships with `chrome-extension://EXTENSION_ID_PLACEHOLDER/`. Replace that origin when the extension id exists. Do not open a localhost WebSocket instead. Chrome authenticates the host by launching it.

## Good first issues

The tracker is not attached yet (GitHub login on the maintainer machine is expired). The issues already exist in the repo so you can pick one up now:

- [Index](docs/contributing/GOOD_FIRST_ISSUES.md)
- Individual write-ups in [docs/contributing/issues](docs/contributing/issues)

Each issue has a title, the labels to apply, the files to touch, and a definition of done. Comment on the issue file's pull request, or on the GitHub issue once it is filed, before you start so two people do not take the same task.

After `gh auth login`, a maintainer can open them on GitHub:

```sh
python3 scripts/file_github_issues.py
```

That creates the `help wanted` and `good first issue` labels if they are missing, then opens one issue per file.

## Pull requests

- One issue per pull request when you are picking up a starter task.
- Run `make test` before you ask for review.
- Keep the native messaging host quiet on stdout. Chrome reads stdout as the protocol stream. Logs belong on stderr.
- Do not add a TCP or WebSocket listener. The extension uses native messaging. The CLI uses the per-user unix socket.
- Profiles that contradict themselves must fail `profile.Parse`. Add a test next to the change.
- Do not commit `daemon/bin/` or a live `cloakwork.sock`.

## Commit messages

Use a short sentence that says why the change exists. Example: `Reject cohorts whose WebGL renderer disagrees with the claimed OS.`
