---
title: "Add a coherent macOS reference cohort"
labels:
  - help wanted
  - good first issue
---

Sprint 1 ships one Windows cohort, `cohort-win11-chrome-en-us`. A second cohort is the smallest way to learn the schema: add a macOS profile that `profile.Parse` accepts, and show that a mixed profile is still rejected.

## Where to look

- `daemon/internal/profile/profile.go` lists the platform, font, WebGL, and time zone rules.
- `daemon/internal/profile/cohorts/cohort-win11-chrome-en-us.json` is the pattern.
- `daemon/internal/profile/profile_test.go` shows rejection cases.

## Constraints

- Platform `MacIntel`, client-hints platform `macOS`, a Macintosh user agent, an Apple WebGL renderer, and a font list that includes `Helvetica`.
- Use `America/Los_Angeles` with an offset the validator already allows (420 or 480).
- `webRtcPolicy` stays `disable_non_proxied_udp`.
- Do not add canvas noise, random values, or a `HeadlessChrome` user agent.
- Leave `profile.DefaultID` on the Windows cohort so current callers stay stable.

## Done when

- The new JSON file loads through `LoadEmbedded`.
- `cloakwork profile list` prints both ids.
- A test rejects that macOS file if its renderer is rewritten to a Direct3D string.

## Out of scope

Do not inject the profile into Chrome. That is Sprint 2.
