---
title: "Add a cloakwork version command"
labels:
  - help wanted
  - good first issue
---

`config.Version` exists, and `cloakwork` prints it inside the help text, but there is no `cloakwork version` command. Scripts and bug reports need a stable one-line version.

## Where to look

- `daemon/internal/config/config.go` holds `Version`.
- `daemon/cmd/cloakwork/main.go` dispatches subcommands.

## Done when

- `cloakwork version` prints the version string and exits 0.
- `cloakwork --version` and `cloakwork -v` do the same thing.
- A test covers the command. Build the binary in the test with `go test` and run it, or extract the command into a function that returns the line.
- Help text mentions the new command.

## Out of scope

Do not change the native messaging protocol or the profile schema.
