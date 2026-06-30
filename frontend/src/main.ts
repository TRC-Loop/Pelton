// bundled ui and mono fonts via @fontsource (no runtime cdn). weights cover the
// regular/medium/semibold/bold the ui uses and the mono weights for mail bodies.
import '@fontsource/familjen-grotesk/400.css'
import '@fontsource/familjen-grotesk/500.css'
import '@fontsource/familjen-grotesk/600.css'
import '@fontsource/familjen-grotesk/700.css'

import '@fontsource/spline-sans-mono/300.css'
import '@fontsource/spline-sans-mono/400.css'
import '@fontsource/spline-sans-mono/500.css'
import '@fontsource/spline-sans-mono/600.css'
import '@fontsource/spline-sans-mono/700.css'

// theme tokens must load before the base styles that reference them.
import './theme/tokens.css'
import './style.css'

import App from './App.svelte'

const target = document.getElementById('app')
if (!target) {
  throw new Error('pelton: #app mount point missing')
}

const app = new App({ target })

// remove the pre-mount splash now that svelte has taken over.
document.getElementById('splash')?.remove()

export default app
