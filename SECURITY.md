# Security Policy

## Supported versions

| Version | Supported |
|---|---|
| latest (`main`) | ✅ |
| older releases | ❌ (please upgrade) |

## Reporting a vulnerability

**Please do not open a public GitHub issue for security vulnerabilities.**

Use one of these private channels instead:

1. **GitHub private advisory** — [Create a security advisory](https://github.com/anmolnagpal/aiswitch/security/advisories/new) (preferred)
2. **Email** — Contact the maintainer via their GitHub profile

You will receive a response within 72 hours. If the issue is confirmed, a patch will be released as soon as possible (typically within 7 days for critical issues).

## Scope

Things we consider in scope:

- Secrets (API keys, tokens) being written to world-readable files
- Privilege escalation via the binary or its config files
- Credential exfiltration through the network
- Path traversal in `.aiswitch` file parsing

## Out of scope

- Issues in dependencies that do not affect aiswitch's security posture
- Missing rate-limiting (aiswitch is a CLI tool, not a server)
- Social engineering attacks
