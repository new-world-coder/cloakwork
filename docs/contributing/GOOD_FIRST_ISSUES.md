# Help wanted: good first issues

These tasks are sized for a first contribution. Each one has the labels **help wanted** and **good first issue**. Full write-ups are in [issues](issues).

| Issue | You will touch |
| --- | --- |
| [Add a version command](issues/01-version-command.md) | `daemon/cmd/cloakwork/main.go` |
| [Add a macOS reference cohort](issues/02-macos-cohort.md) | `daemon/internal/profile/cohorts/` |
| [Add GitHub Actions CI](issues/03-github-actions-ci.md) | `.github/workflows/ci.yml` |
| [Publish a JSON Schema for profiles](issues/04-profile-json-schema.md) | `daemon/internal/profile/` |
| [Cover Chrome Canary and Chromium host paths](issues/05-chrome-channel-paths.md) | `daemon/internal/install/` |
| [Document the native messaging examples](issues/06-protocol-examples.md) | `docs/protocol.md` |
| [Add status --json](issues/07-status-json.md) | `daemon/cmd/cloakwork/main.go` |

A maintainer files these on GitHub with `python3 scripts/file_github_issues.py` after `gh auth login`. Until then, open a pull request that links the issue filename.
