# Create a theme

This walks you through a small but complete dark theme called Nordish. You need a text editor and a way to make a zip file, nothing else.

## 1. Lay out the folder

```text
nordish/
├── manifest.json
├── tokens/
│   └── colors.json
└── css/
    └── polish.css
```

`manifest.json`:

```json
{
  "manifestVersion": 1,
  "id": "nordish",
  "name": "Nordish",
  "author": "you",
  "version": "1.0.0",
  "description": "An arctic, north-bluish theme.",
  "base": "dark",
  "pelton": { "min": "1.0.8" },
  "tokens": ["tokens/colors.json"],
  "css": ["css/polish.css"]
}
```

`tokens/colors.json`, the actual look. Start from the surfaces and text, then the accent:

```json
{
  "surface-base": "#2e3440",
  "surface-raised": "#3b4252",
  "surface-overlay": "#434c5e",
  "surface-sunken": "#292e39",
  "surface-hover": "#434c5e",
  "text-primary": "#eceff4",
  "text-secondary": "#d8dee9",
  "text-tertiary": "#9aa5b8",
  "accent": "#88c0d0",
  "accent-fg": "#2e3440"
}
```

`css/polish.css` is optional. Use it only for what tokens cannot express:

```css
::-webkit-scrollbar-thumb {
  background: #4c566a;
  background-clip: padding-box;
  border: 3px solid transparent;
}
```

## 2. Zip it

The manifest must sit at the root of the archive, not inside a subfolder:

```sh
cd nordish
zip -r ../Nordish.peltontheme manifest.json tokens css
```

## 3. Import and iterate

Open **Settings, Themes, Import theme** and pick the file. You will see the metadata and the raw CSS before anything installs. After installing, click the Nordish card to apply it.

For quick iteration you do not need to re-zip after every change. Installed themes live extracted in your data folder:

| Platform | Path |
| -------- | ---- |
| macOS | `~/Library/Application Support/Pelton/themes/nordish/` |
| Linux | `~/.config/Pelton/themes/nordish/` |
| Windows | `%AppData%\Pelton\themes\nordish\` |

Edit the files there, press **Reload** in **Settings, Themes**, and re-activate the theme to see the change. The same validation as at import runs on every load, so a broken edit tells you what is wrong instead of half-applying.

## 4. Bundle fonts and images the right way

Want a custom font? Put the file in the container and reference it relatively:

```css
@font-face {
  font-family: "Inter";
  src: url("assets/fonts/inter.woff2") format("woff2");
}
```

```json
{ "font-ui": "\"Inter\", system-ui, sans-serif" }
```

Bundled references are inlined when the theme applies, so they work offline and make no network requests. Remote URLs would trigger the tracking warning at import, and users are told to distrust them; just bundle instead.

## 5. Export and share

**Settings, Themes** has an export button on every installed theme. It zips the installed folder (including any hand edits) back into a `Name.peltontheme` you can send to anyone.

A few finishing touches worth doing before sharing:

- Add a `preview` screenshot to the manifest so the gallery card shows your theme instead of a plain swatch.
- Set `pelton.min` to the app version you tested against.
- Bump `version` on every release; users importing a newer version of an installed theme get an update prompt instead of a duplicate.
