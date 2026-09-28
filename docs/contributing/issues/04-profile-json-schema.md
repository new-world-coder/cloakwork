---
title: "Publish a JSON Schema for cohort files"
labels:
  - help wanted
  - good first issue
---

`profile.Parse` rejects unknown fields and contradictory signals, but editors cannot see that until a cohort is loaded by the daemon. Publish a JSON Schema document that matches the Go struct.

## Where to look

- `daemon/internal/profile/profile.go` for the fields and enums.
- `daemon/internal/profile/cohorts/cohort-win11-chrome-en-us.json` must validate against the schema.

## Done when

- A schema file lives at `daemon/internal/profile/cohort.schema.json`.
- A Go test loads the shipped cohort and checks it against the schema with a small dependency, or with a focused set of assertions if you want to avoid a new module dependency. Say which you chose in the pull request.
- The schema marks the known fields as required and sets `additionalProperties` to false.
- Enum fields (`platform`, `webRtcPolicy`) list only the values `Validate` accepts.

## Out of scope

The schema does not have to encode every cross-field rule (Windows plus Apple GPU). Those stay in `Validate`. Mention that limit in a schema `$comment` or in the pull request.
