---
title: "Contributing guidelines"
description: "The full contributor guidelines for Pelton: ways to help, the code of conduct, AI-assisted development rules, and commit conventions."
---

??? note "Editing this page"
    This is a generated copy of `CONTRIBUTING.md` from the repo root. To change
    its content, edit that file and run `make sync-docs` (CI does this
    automatically on push). Edits made directly to this file are
    overwritten on the next sync.

!!! tip "New here?"
    Optional but useful reading before opening your first PR. Once you have,
    see [Developing on Pelton](developing.md) or
    [Translating Pelton](translations.md) to get started.

<img width="2560" height="1440" alt="contr-pelton" src="../assets/contributing-banner.webp" />

<p align="center"><strong>First of all, thanks for taking the time to contribute! 🎉</strong></p>

*Before Contributing please read the Information below.*

**There are multiple ways you can Contribute to the Pelton Project:**

1. [Submit a Theme ↗](https://github.com/peltonapp/themes/issues/new?template=submit_theme.yml)
2. [Contribute to the Website (pelton.app) ↗](https://github.com/peltonapp/website)
3. [Submitting a Feature (Request)](https://github.com/peltonapp/Pelton/issues)
4. [Reporting a Bug](https://github.com/peltonapp/Pelton/issues)
5. [Requesting a Language or translating Pelton](https://github.com/peltonapp/Pelton/issues)
6. Submitting PRs
   
## Code of Conduct

Participation in Pelton is governed by the [Code of Conduct](code-of-conduct.md). Read it before opening an issue or PR, the short version at the top only takes a minute.

## AI-Assisted Development

Pelton has files to guide AI, starting with [AGENTS.md](https://github.com/peltonapp/Pelton/blob/main/AGENTS.md) as the entrypoint.

**Allowance**

You may use AI tools such as Claude Code, Codex, Aider, or similar. Fully autonomous systems that open PRs without human review are not allowed.

**Limits**

AI can be used for writing code, but architecture and the way things should work, look, and be implemented has to come from a human.

When working on the frontend, stick to Pelton's existing design, don't let AI introduce generic patterns like purple-blue gradients.

**Disclosure**

Contributors must state AI usage in the commit or PR description, e.g. `Assisted-by: Claude Code`. This is for transparency only and doesn't lead to different handling of the contribution. It's fully okay. 

*TLDR; Yes.*

## Developer Certificate of Origin (DCO)

### Full DCO text

The full, unmodified text of the Developer Certificate of Origin 1.1 is in [Developer Certificate of Origin](dco.md). When you sign off a commit, that's what you're attesting to.

### Why

Two things matter when you send code to Pelton:

1. You actually wrote it, or you have the right to submit it under the project's GPL-3.0 license. Code copied from a GPL project, code owned by your employer, or AI-generated code can't be accepted unless the permissions check out.
2. You're fine with it being GPL-3.0 from that point on. Once it's merged it stays under that license.

The DCO is how you confirm both, in one line, per commit.

### How to sign off

Set your name and email in git, once:

```bash
git config --global user.name "John S."
git config --global user.email "john@example.com"
```

Then pass `-s` when you commit:

```bash
git commit -s -m "fix: handle nil response in HTTP check"
```

That adds a line at the bottom of your commit message:

```text
Signed-off-by: John S. <john@example.com>
```

The DCO bot checks every PR. If a commit is missing the sign-off, it blocks the merge until you fix it.

> [!TIP]
> If using Claude Code or Codex, use a prompt like `Add -s to git commits.`.


## Commit Message Convention

Pelton uses [Conventional Commits](https://www.conventionalcommits.org/) for commit messages. This keeps the history readable and allows changelogs to be generated automatically.

Each commit message should follow this format:

```text
<type>(<optional scope>): <short description>

<optional body>

<optional footer(s)>
```

**Common types:**

- `feat`: a new feature
- `fix`: a bug fix
- `docs`: documentation-only changes
- `style`: formatting, missing semicolons, etc. (no code logic change)
- `refactor`: code change that neither fixes a bug nor adds a feature
- `perf`: performance improvement
- `test`: adding or correcting tests
- `chore`: build process, tooling, dependency updates

**Examples:**
```text
feat(oauth): add Google OAuth token refresh
fix(inbox): resolve crash when marking empty selection as read
docs: update setup instructions in README
```

Breaking changes should be indicated with a `!` after the type/scope, and explained in the footer:
```text
feat(api)!: change account sync response format

BREAKING CHANGE: sync endpoint now returns paginated results
```

Please keep the short description under ~72 characters, written in the imperative mood ("add" not "added" or "adds"). Don't forget to sign off your commits per the DCO requirement above.

## Pelton-specific

As Pelton is privacy focused you have a couple more rules:

- No telemetry or phone-home (self-explanatory)
- Any external request (except for IMAP/SMTP) must be off by default
  > If you add anything that even reaches out to any 3rd party server, it has to be off by default.
  > 
  > **Example:** Let's say you'd be adding a VirusTotal integration: Off by default and must be enabled first in settings.


- Most things should be customizable and toggleable in settings. Make sure to put the setting you're adding in a fitting place.
