# UI style and design tokens

Read this before touching anything under `frontend/src/components/`,
`frontend/src/theme/`, or `frontend/src/style.css`.

## Style

- Minimal, native-feeling desktop app: low chrome, no gradients or heavy
  shadows, hairline borders (`--hairline`), thin single-sided accent borders
  rather than boxes.
- Dense and keyboard/power-user oriented: compact list rows, vim mode in
  compose, shortcuts throughout. Favor information density over whitespace
  unless the user's chosen density setting says otherwise.
- Fonts are bundled locally via `@fontsource` packages, no runtime CDN
  fetches (that would also violate the privacy/offline stance, see
  [privacy.md](privacy.md)).

## Tokens: the hard rule

Every color, spacing, radius, font, and size is a CSS custom property defined
in `frontend/src/theme/tokens.css`. **Components reference tokens, never
literal hex/px values.** A hardcoded color or pixel value in a component is a
bug, not a style choice. A theme is nothing more than another named token set
in that same shape.

Token contract (the complete surface every theme must define):

| Group     | Tokens |
|-----------|--------|
| surfaces  | `--surface-base/raised/overlay/sunken/hover`, `--selection-bg` |
| text      | `--text-primary/secondary/tertiary/inverse`, `--link` |
| borders   | `--border-subtle/default/strong`, `--hairline` (width) |
| accent    | `--accent` (runtime), `--accent-fg` (runtime), `--selection-bg`, `--link` |
| semantic  | `--success/-bg`, `--warning/-bg`, `--danger/-bg` |
| radii     | `--radius-control`, `--radius-card`, `--radius-none` |
| fonts     | `--font-ui`, `--font-mono` |
| type      | `--fz-list/body/label/meta/heading`, `--fw-*` weights |
| spacing   | `--space-1..6` (fixed) plus density-driven row/pane spacing |
| elevation | `--shadow-overlay` |

- `--accent` is injected at runtime from the user's chosen hex (default
  `#465AF2`); `--accent-fg` (black/white) is computed from luminance in
  `accent.ts` so text on an accent surface stays legible for any color the
  user picks. `--selection-bg` and `--link` are derived from `--accent` via
  `color-mix()` so the user only ever picks one color. Don't add a second
  user-facing color picker without discussing it first.
- Every token needs both a light and dark value (`:root[data-theme="light"]`
  / `:root[data-theme="dark"]`), so the theme toggle never leaves a gap.
- Density variants (`compact` / `medium` / `luxe`, set via
  `:root[data-density=...]`) only ever swap spacing and line-height tokens.
  Components must stay density-agnostic: build against `--row-pad-y`,
  `--pane-pad`, `--control-height`, etc., not fixed numbers, so they adapt
  automatically.
- Extend `tokens.css` if a new token is genuinely needed, in both themes and
  across density variants where relevant, rather than reaching for a literal
  value or a one-off local variable.
