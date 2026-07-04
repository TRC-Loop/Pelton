# No slop

- No AI-sounding boilerplate in anything user-facing or in commit/PR text:
  no "leverage", "seamless", "robust solution", no em-dashes (never, in any
  UI copy), no filler adjectives, no unnecessary emoji.
- No unused exports, dead code paths, orphaned files, or unused
  params/imports left behind after a refactor. If it's unused, delete it,
  don't comment it out or gate it behind an unused flag.
- No new Go module or npm dependency without flagging it to the user first
  and explaining why an existing dependency or the standard library can't
  do it.
- No speculative abstractions, no half-finished features, no
  backwards-compatibility shims for code that hasn't shipped yet.
- Comments explain non-obvious *why*, never *what* (the code already says
  what). Don't reference the current task/issue/PR in comments, that
  context belongs in the commit message and rots in the code.
