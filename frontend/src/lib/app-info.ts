// app-info.ts holds the static facts shown in the settings about section. the
// version is hardcoded for now; wiring it to the build (ldflags or wails.json)
// is a small follow-up.
// TODO(build): inject version at build time instead of hardcoding here.
export const APP = {
  name: 'Pelton',
  version: '0.1.0',
  tagline: 'An open-source desktop mail client.',
  repo: 'https://github.com/TRC-Loop/Pelton',
  issues: 'https://github.com/TRC-Loop/Pelton/issues',
  contributors: 'https://github.com/TRC-Loop/Pelton/graphs/contributors',
  license: 'GPL-3.0',
  licenseUrl: 'https://www.gnu.org/licenses/gpl-3.0.html',
  author: 'Arne K.',
} as const
