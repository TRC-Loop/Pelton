---
title: Create a theme
description: Write, test, and share a Pelton theme, from a manifest to a pull request on pelton-themes.
---

# Create a theme

!!! warning "Don't use the built-in editor for a theme you plan to share"
    Settings > Themes has a **New theme...** button, it's a quick palette
    editor for personal tweaks, not for building a theme you intend to
    give to someone else. It's structurally limited compared to hand-writing
    a theme file:

    - Only one CSS file, at a fixed path, hand-authored themes can ship as
      many as they want.
    - No way to add icon overrides at all.
    - No `preview` screenshot field.
    - No `homepage`, `license`, or `pelton` compatibility-range fields.
    - No way to bundle a font or image, the underlying format supports it,
      the editor's UI just doesn't expose it.
    - Any remote CSS reference is silently stripped, with no warning shown
      (unlike importing a file, which always warns you first).
    - Once a theme has any CSS attached to it, saved through the editor or
      imported from a file, the editor's pencil/edit button disappears for
      it permanently. From then on you can only change it by hand-editing
      the container.

    Use the editor to preview a quick palette swap for yourself. For
    anything you want to publish, follow the rest of this page instead.

## Checklist
<div class="checklist" markdown>
- [ ] Write the manifest and token files
- [ ] Add CSS, icons, or a bundled font, if you want them
- [ ] Zip it into a `.peltontheme` file
- [ ] Test it by importing it into Pelton
- [ ] Share it on pelton-themes
</div>

## 1. Clone the template

Start from [peltonapp/pelton-theme-template](https://github.com/peltonapp/pelton-theme-template), a ready-to-edit skeleton. You'll need [git](https://git-scm.com) installed to clone it:

```bash
git clone https://github.com/peltonapp/pelton-theme-template.git my-theme
cd my-theme
```

Fresh out of the clone, it looks like this:

```text
my-theme/
├── README.md              # your description
├── LICENSE
└── source/                # what you actually edit, not shipped as-is
    ├── manifest.json
    └── tokens/
        └── colors.json
```

You'll add a `source/css/` folder yourself if you want CSS overrides, see the next few steps.

Don't want to use git, or building a theme only for yourself? You don't need the template repo at all, just a `manifest.json` and a `tokens/` folder, laid out the same way, anywhere on disk.

`source/manifest.json`, metadata plus paths to the files that hold the actual overrides:

```json
{
  "manifestVersion": 1,
  "id": "my-theme",
  "name": "My Theme",
  "author": "you",
  "version": "1.0.0",
  "description": "An arctic, north-bluish theme.",
  "base": "dark",
  "pelton": { "min": "1.0.8" },
  "tokens": ["tokens/colors.json"],
  "css": ["css/polish.css"]
}
```

`source/tokens/colors.json`, a plain `{token: value}` object, always keep tokens in their own file like this rather than inlining them into the manifest, see the [full token list](format.md#tokens) for every name you can use here:

```json
{
  "surface-base": "#242933",
  "surface-raised": "#2e3542",
  "text-primary": "#e6e9ef",
  "accent": "#88c0d0"
}
```

`source/css/polish.css`, for anything the token list doesn't cover:

```css
.md-message-preview {
  border-radius: var(--radius-card);
}
```

## 2. Bundle a font, if you want one

Reference it with a normal relative `url()` in your CSS and include the font file in the archive next to it:

```css
@font-face {
  font-family: "Your Font";
  src: url("./fonts/your-font.woff2") format("woff2");
}

:root {
  --font-ui: "Your Font", sans-serif;
}
```

Pelton inlines bundled assets as `data:` URIs when the theme is applied, so there's no separate install step, just make sure the file is inside the archive and under the 5 MB per-asset limit.

!!! warning "Don't link to a remote font or image"
    A theme can't make external requests. A `url()` or `@import` in your CSS pointing anywhere outside the theme file itself, a remote font, a remote image, gets flagged: Pelton warns whoever imports it and blocks it by default. Bundle the file instead, as shown above.

## 3. Add icon overrides, if you want them

Map a Tabler icon name to an SVG file in `manifest.json`:

```json
{
  "icons": {
    "compose": "icons/compose.svg"
  }
}
```

Draw with `currentColor` so the icon follows the surrounding text color.

!!! warning "Only plain vector markup survives"
    Every icon SVG is sanitized before use: no `<script>`, no event-handler attributes, no `javascript:` links, no `<iframe>`/`<object>`/`<use>`/`<animate>` elements, no `href`, no `url()`/`@import`. Keep it to plain shapes and paths, see [Theme format](format.md#icons) for the full list of what gets stripped.

## 4. Zip it

From inside `source/`, zip the contents, not the folder itself, so `manifest.json` ends up at the archive root:

```bash
cd source
zip -r ../my-theme.peltontheme manifest.json tokens css assets
```

`zip` skips a folder that doesn't exist (no `assets/` yet, no icons), so it's fine to always list `tokens css assets` even before you've added all three.

Prefer clicking around in the app instead? **Settings > Themes**' export (upload) icon on an installed theme, including one you built with **New theme...**, saves a ready `.peltontheme` file too. Either route produces the same file.

!!! note "Size limits"
    A container tops out at 20 MB compressed, 1 MB of CSS total, 256 KB per icon SVG, 5 MB per bundled asset, and 512 files, generous for a theme, small enough that nothing can balloon Pelton's memory or disk. See [Theme format](format.md#size-limits) for the full table if you're bundling a lot of fonts or images.

## 5. Test it

Import your own theme the same way anyone else would:

1. **Settings > Themes > Import theme...**, pick the `.peltontheme` file you just built.
2. Check the preview: metadata looks right, every CSS file you expect shows up in the read-only viewer, and there's no unexpected remote-reference warning (if you bundled a font correctly, this shouldn't appear).
3. Confirm the import, then click the theme's card to apply it and click through the app.

While you're iterating, it's faster to skip re-zipping every time: drop your working files straight into the themes folder (**Settings > Themes > Open folder**, paths are listed in [Installing a theme](install.md#where-themes-live)) as a regular folder, then use **Reload** in Settings after each change.

For screenshots, launch Pelton with [`--potatoes-are-nice`](../cli-flags.md#-potatoes-are-nice) to get realistic-looking fake accounts and mail without touching anything real, then switch to your theme and take screenshots against that. It doesn't preview a theme automatically, it's just the standard way to get clean, private-looking screenshots for a `preview` image or a README.

```bash
pelton --potatoes-are-nice
```

## 6. Share it

Publishing a theme is a manual step, there's no in-app upload button, you submit the file to the community repo yourself. Before submitting:

- Bump `version` in your manifest, and set `pelton.min` to the Pelton version you actually tested against.
- Add a `preview.png` screenshot if you don't already have one, and write your theme's own `README.md`, replacing the template's instructions.
- Keep the `LICENSE` file (the template's is CC0, swap it for your own if you want a different license for the theme itself).

Your theme's folder (`README.md`, `LICENSE`, `my-theme.peltontheme`, `preview.png`) needs to land inside a `themes/my-theme/` folder in a clone of [peltonapp/themes](https://github.com/peltonapp/themes), the gallery repo, alongside `source/`:

```bash
git clone https://github.com/peltonapp/themes.git
cp -r ../my-theme pelton-themes/themes/my-theme
cd pelton-themes
python3 scripts/validate_theme.py themes/my-theme
```

`validate_theme.py` runs the same checks CI does (token allowlist, no remote references, size caps, required files), fix anything it reports before submitting. You can also run it with no argument to validate every theme in the repo.

From there, pick one of two ways in:

- **Pull request** (preferred if you're comfortable with git): commit `themes/my-theme/` and open a PR. CI validates it again and a maintainer reviews, then generates a metadata header (version, compatibility, author, license) above your README content automatically.
- **Issue form**: open a ["Submit a theme"](https://github.com/peltonapp/themes/issues/new?template=submit_theme.yml) issue and attach your files, a maintainer turns it into a PR for you. GitHub won't let you attach a `.peltontheme` file directly to an issue, since it's already a zip, rename it to `.zip` before attaching (or drop it inside a new `.zip`), the repo's CI renames it back automatically once it lands.

Once merged, your theme shows up on [themes.pelton.app](https://themes.pelton.app) for anyone to install.

!!! tip "Before you submit"
    - Keep body text readable on every surface you touch, aim for WCAG AA contrast.
    - Set both `accent` and `accent-fg` so text on an accent-colored background stays legible.
    - Don't try to override spacing or density, it isn't themeable, design with the built-in spacing in mind.
    - Bundle fonts and images rather than linking them remotely, bundled references get inlined and work offline.
    - Credit your source if you're porting an existing palette.

## Need help?

See [Support](../support.md).
