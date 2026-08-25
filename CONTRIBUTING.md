<img width="2560" height="1440" alt="contr-pelton" src="https://github.com/user-attachments/assets/e18a6bd6-ecbb-4b62-9585-09b55321dab0" />

<p align="center"><strong>First of all, thanks for taking the time to contribute! 🎉</strong></p>

*Before Contributing please read the Information below.*

**There are multiple ways you can Contribute to the Pelton Project:**
1. [Submit a Theme ↗](https://github.com/TRC-Loop/pelton-themes/issues/new?template=submit_theme.yml)
2. [Contribute to the Website (pelton.app) ↗](https://github.com/TRC-Loop/pelton.app)
3. [Submitting a Feature (Request)](https://github.com/TRC-Loop/Pelton/issues)
4. [Reporting a Bug](https://github.com/TRC-Loop/Pelton/issues)
5. [Requesting a Language or translating Pelton](https://github.com/TRC-Loop/Pelton/issues)
6. Submitting PRs
   
## Code of Conduct

The rules are short because they have to be remembered, not looked up. Be respectful. Don't be rude to people asking questions, filing their first issue, or making mistakes. If you'd be embarrassed saying it face-to-face, don't type it.

Assume good faith. When something someone wrote could be read two ways, pick the charitable reading. Non-native English speakers sometimes sound blunt when they're being neutral. New contributors miss things because they haven't read the docs yet, not because they're disrespecting you. Ask before you assume. Disagree on ideas, not people. "This approach has a race condition" is fine. "You clearly don't understand concurrency" is not. Code can be wrong. People shouldn't be attacked for writing it.

Zero tolerance for harassment or discrimination. Targeting someone for their race, gender, sexuality, religion, nationality, or disability, threatening, or doxxing anyone gets you banned from the project. No warnings, no debate.

Report problems to [me@arne.sh ↗](mailto:me@arne.sh). Reports are handled privately.

*TLDR; Be nice and responsible*

## AI-Assisted Development

Pelton has files to guide AI, starting with [AGENTS.md](https://github.com/TRC-Loop/Pelton/blob/main/AGENTS.md) as the entrypoint.

**Allowance**

You may use AI tools such as Claude Code, Codex, Aider, or similar. Fully autonomous systems that open PRs without human review are not allowed.

**Limits**

AI can be used for writing code, but architecture and the way things should work, look, and be implemented has to come from a human.

When working on the frontend, stick to Pelton's existing design — don't let AI introduce generic patterns like purple-blue gradients.

**Disclosure**

Contributers must state AI usage in the commit or PR description, e.g. `Assisted-by: Claude Code`. This is for transparency only and doesn't lead to different handling of the contribution. It's fully okay. 

*TLDR; Yes.*

## Developer Certificate of Origin (DCO)

### Full DCO text

The full, unmodified text of the Developer Certificate of Origin 1.1 is in [DCO.md](./DCO.md). When you sign off a commit, that's what you're attesting to.

### Why

Two things matter when you send code to Pelton:

1. You actually wrote it, or you're allowed to submit it. Code copied from a GPL project, code owned by your employer, or code an AI generated from copyrighted training data can't be accepted unless the permissions check out.
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

```
Signed-off-by: John S. <john@example.com>
```

The DCO bot checks every PR. If a commit is missing the sign-off, it blocks the merge until you fix it.

> [!TIP]
> If using Claude Code or Codex, use a prompt like `Add -s to git commits.`.


## Commit Message Convention

Pelton uses [Conventional Commits](https://www.conventionalcommits.org/) for commit messages. This keeps the history readable and allows changelogs to be generated automatically.

Each commit message should follow this format:

```
<type>(<optional scope>): <short description>

<optional body>

<optional footer(s)>
```

**Common types:**
- `feat` — a new feature
- `fix` — a bug fix
- `docs` — documentation only changes
- `style` — formatting, missing semicolons, etc. (no code logic change)
- `refactor` — code change that neither fixes a bug nor adds a feature
- `perf` — performance improvement
- `test` — adding or correcting tests
- `chore` — build process, tooling, dependency updates

**Examples:**
```
feat(oauth): add Google OAuth token refresh
fix(inbox): resolve crash when marking empty selection as read
docs: update setup instructions in README
```

Breaking changes should be indicated with a `!` after the type/scope, and explained in the footer:
```
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
