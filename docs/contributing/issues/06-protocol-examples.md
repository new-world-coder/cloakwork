---
title: "Document native messaging with a worked example"
labels:
  - help wanted
  - good first issue
---

The framing rules live in `daemon/internal/protocol/frame.go`. A new contributor should be able to read one page and speak to `cloakwork host` without reading the server.

## Done when

- `docs/protocol.md` explains the 4-byte little-endian length prefix and the 1 MiB limit.
- It shows request and response JSON for `ping`, `getStatus`, and `getProfile`.
- It states that stdout is the message stream and logs go to stderr.
- It states that another uid on the unix socket receives `{"type":"error","error":"unauthorized"}`.
- It links the page from `README.md`.

## Out of scope

Do not document Sprint 2 injection, time zone writes, or WebRTC policy changes. Those commands are rejected on purpose today (`unknown message type`).
