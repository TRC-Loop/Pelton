# Themes

Pelton's whole interface is driven by design tokens: named values for every color, font, radius and shadow. A theme is a file that overrides some of those tokens, optionally adds CSS for things tokens cannot express, and can even swap interface icons. Anything a theme does not override falls back to the built-in light or dark look, so a theme is always complete.

Themes ship as single `.peltontheme` files (a zip container). Install one under **Settings, Themes, Import theme**, share one by sending the file.

<div class="grid cards" markdown>

- **[Theme format](format.md)**

    The full `.peltontheme` spec: container layout, manifest fields, the themeable token surface, CSS rules and icon overrides.

- **[Create a theme](create.md)**

    Build a working theme from scratch, test it live, and export it for sharing.

</div>

## Security model, in short

Themes are code-adjacent, so Pelton treats them with care:

- Before anything installs, you see the theme's metadata and the raw contents of every CSS file it ships.
- CSS that references the network (remote `url()`, `@import`) triggers an explicit warning listing every reference. You choose whether to keep them or have them stripped; stripping is the default. A theme that loads a remote resource can be used to track you, which is why well-made themes bundle fonts and images inside the file instead.
- Icon SVGs are checked at import: scripts, event handlers and external references are rejected outright.
- Token values are validated against an allowlist, so a theme cannot smuggle arbitrary CSS through a color field.
