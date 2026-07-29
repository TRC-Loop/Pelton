# Security policy

## Reporting a vulnerability

Report security vulnerabilities privately, not as a public issue, not in
Discord, and not in a pull request.

Two ways, either is fine:

- [Open a private security advisory](https://github.com/TRC-Loop/Pelton/security/advisories/new)
  on GitHub.
- Email <pelton@arne.sh> with `security` in the subject.

Please include, as far as you can:

- what the issue is and which part of Pelton it affects
  (IMAP/SMTP handling, credential storage, MCP, rendering of remote content,
  themes, the update check, the build pipeline)
- the Pelton version and your operating system
- steps to reproduce, or a proof of concept
- what an attacker gains, and what they need beforehand

We look at every report and will keep you updated while we work on it. We do
not promise a response time, and there is no bug bounty. Please give us a
reasonable chance to ship a fix before you publish details, and let us know if
you plan to disclose on a fixed date.

You are welcome to be credited in the release notes, or to stay anonymous,
whichever you prefer.

## Supported versions

Pelton is under active development and fixes go into the next release. Only the
latest release is supported; there are no backports to older versions.

## Scope

In scope is anything in this repository and the code it ships: the Go backend,
the Svelte frontend, the packaging in `.github/nfpm` and `build/`, and the
release workflows.

Out of scope are the security practices of the email providers you connect to,
issues that require an attacker to already have full access to your unlocked
device, and the content of third-party themes or MCP servers you install
yourself. Reports about the website belong in the
[pelton.app repository](https://github.com/TRC-Loop/pelton.app), same contact.

## What Pelton does not do

Pelton has no telemetry, no analytics and no servers of ours in the data path,
so there is no backend of ours to attack. Your mail, credentials and settings
stay on your device, and Pelton talks only to the providers you configure and
to GitHub when you ask it to check for updates.

Warranty and liability are covered in [DISCLAIMER.md](DISCLAIMER.md).
