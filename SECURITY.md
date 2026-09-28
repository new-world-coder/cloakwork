# Security policy

CloakWork is a local privacy daemon. The native messaging host and the unix socket are the trust boundary.

## Report a problem

Do not open a public issue for a vulnerability that is still unfixed. Until the GitHub repository has private security advisories, contact the maintainers directly and include the version, the command you ran, and the impact.

## In scope

- Another local user reading or writing the daemon socket
- The native host logging protocol data to stdout, or accepting messages Chrome did not deliver
- A shipped cohort that contradicts itself and still loads
- The installer writing the host manifest outside the current user's Chrome directory

## Out of scope

- Claims that a cohort is still identifiable by a commercial fingerprint vendor
- Bypassing a website's bot detection
- Time zone changes, which this sprint does not perform
