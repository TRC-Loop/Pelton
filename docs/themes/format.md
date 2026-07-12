# Theme format

A `.peltontheme` file is a zip container. The only fixed name inside is `manifest.json`; everything else is declared by the manifest, so you can organize files however you like. Extra files such as a README or LICENSE ride along and are ignored by the engine.

```text
Nordish.peltontheme
├── manifest.json          required, everything else is referenced from here
├── tokens/
│   ├── colors.json
│   └── type.json
├── css/
│   └── polish.css
├── assets/
│   ├── fonts/inter.woff2
│   └── preview.png
└── icons/
    └── palette.svg
```

## manifest.json

```json
{
  "manifestVersion": 1,
  "id": "nordish",
  "name": "Nordish",
  "author": "Arne K.",
  "version": "1.0.0",
  "description": "An arctic, north-bluish theme.",
  "license": "MIT",
  "base": "dark",
  "pelton": { "min": "1.0.8" },
  "tokens": ["tokens/colors.json", "tokens/type.json"],
  "css": ["css/polish.css"],
  "preview": "assets/preview.png",
  "icons": { "palette": "icons/palette.svg" }
}
```

| Field | Required | Meaning |
| ----- | -------- | ------- |
| `manifestVersion` | yes | Container format version, currently `1`. A file with a newer version than the app understands is refused, since an unknown format cannot degrade gracefully. |
| `id` | no | Stable lowercase slug (`a-z`, `0-9`, dashes). Names the install folder and keys update detection: importing a known `id` with a higher `version` offers an update instead of a duplicate. Defaults to a slug of `name`. |
| `name` | yes | Display name, free-form. |
| `author`, `version`, `description`, `homepage`, `license` | no | Shown in the gallery and import preview. |
| `base` | yes | `light` or `dark`. Which built-in token set fills everything the theme does not override, and what the OS-level color scheme reports. |
| `pelton` | no | `{ "min": "...", "max": "..." }`, both optional. The app version range the theme was made for. Outside the range is a warning badge, never a block. |
| `tokens` | no | List of JSON files merged in order (later wins), or one inline object. |
| `css` | no | List of CSS files, concatenated in this order into one stylesheet. |
| `preview` | no | Screenshot shown in the theme gallery and import preview. |
| `icons` | no | Map of icon name to SVG file, see below. |

## Tokens

Token files are flat maps of token name to CSS value:

```json
{
  "surface-base": "#2e3440",
  "accent": "#88c0d0",
  "radius-control": "7px",
  "font-ui": "\"Inter\", system-ui, sans-serif"
}
```

Names are the app's token names without the `--` prefix (a leading `--` is also accepted). Only allowlisted tokens are themeable; an unknown name fails the import loudly rather than being ignored, so typos surface immediately.

The themeable surface:

| Group | Tokens |
| ----- | ------ |
| Surfaces | `surface-base`, `surface-raised`, `surface-overlay`, `surface-sunken`, `surface-hover`, `selection-bg`, `selection-bg-strong` |
| Text | `text-primary`, `text-secondary`, `text-tertiary`, `text-inverse`, `link` |
| Borders | `border-subtle`, `border-default`, `border-strong`, `hairline` |
| Accent | `accent`, `accent-fg` |
| Semantic | `success`, `success-bg`, `warning`, `warning-bg`, `danger`, `danger-bg` |
| Radii | `radius-control`, `radius-card`, `radius-none` |
| Fonts | `font-ui`, `font-mono` |
| Type | `fz-meta`, `fz-label`, `fz-list`, `fz-body`, `fz-heading`, `fz-title`, `fw-regular`, `fw-medium`, `fw-semibold`, `fw-bold` |
| Elevation | `shadow-overlay` |

Spacing and density tokens are deliberately not themeable: density is a user setting, and a theme fighting it would break layouts.

Values go through a safety check: no semicolons, braces, `@` or `url()`. Everything colors, font stacks and shadows need still works, including `color-mix()`.

## CSS

CSS files are concatenated in manifest order into a single stylesheet that loads after the theme's token overrides, so "base file, then override file" layering behaves the way you would expect.

Rules for references:

- `url("assets/...")` pointing at a file inside the container is the intended way to use fonts and images. The file is inlined at apply time, so it never touches the network.
- Remote references (`url("https://...")`, any `@import`) are flagged at import. The import dialog lists each one and asks whether to keep or strip them; strip is the default. Whatever you choose is baked into the installed files.

Size caps: 20 MB per container, 1 MB of CSS total, 5 MB per inlined asset.

## Icons

The `icons` map replaces interface icons by name. Names are the [Tabler icon](https://tabler.io/icons) names the app uses, in lowercase kebab, without the `Icon` prefix: `IconPalette` becomes `palette`.

- SVGs should draw with `currentColor` so they follow the theme's text colors.
- SVGs are sanitized at import and again on every load: script elements, event handler attributes, `href`, `foreignObject` and external references are rejected with an error naming the file.
- Unknown names are accepted and ignored, so a theme can ship icons for future app versions.
- Coverage grows over time: the app routes icons through a themeable wrapper and converts call sites progressively.

## Versioning

Two independent fields, two behaviors:

- `manifestVersion` is the format version, for the engine. Newer than the app understands means the import is refused with a clear message.
- `pelton` is the app version range, for humans. Outside the range shows a warning at import and a badge in the gallery ("made for Pelton 1.0.8 or newer"), but the theme still installs and applies. This also covers the reverse case: when the app updates past a theme's `max`, the badge appears without breaking anything.
