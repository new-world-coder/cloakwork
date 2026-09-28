---
title: "Run go test on GitHub Actions"
labels:
  - help wanted
  - good first issue
---

Contributors need a check that runs before review. Add a workflow that runs the Sprint 1 test suite.

## Done when

- `.github/workflows/ci.yml` runs on pull requests and on pushes to `main`.
- The job uses a current Go 1.26.x release and runs `go test ./...` in `daemon/`.
- The workflow checks out the repo and does not need secrets.
- `README.md` or `CONTRIBUTING.md` mentions the check in one sentence.

## Notes

The peer-credential acceptance test skips itself off macOS. The rejection test injects a fake uid and should pass on the Ubuntu runner. Do not disable the whole `sock` package to make CI green.

## Out of scope

Do not add lint, release, or deploy jobs in this change.
