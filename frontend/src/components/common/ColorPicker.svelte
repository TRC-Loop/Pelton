<script lang="ts">
  // a small self-contained color picker: a saturation/value square, a hue slider
  // and a hex field, all in one panel. it exists because the native color input
  // is inconsistent (and poor on macos). it speaks hex in and out through the
  // value prop and the change event, so callers stay simple. the spectrum
  // gradients here are intrinsic to a color picker, not theme chrome.
  import { createEventDispatcher } from 'svelte'
  import { isValidHex, normalizeHex } from '../../theme/accent'
  import { t } from '../../lib/i18n'

  export let value = '#465AF2'

  const dispatch = createEventDispatcher<{ change: string }>()

  let h = 0
  let s = 100
  let v = 100
  let hexText = value
  // lastHex guards against the external value prop fighting our own emits.
  let lastHex = ''
  let svEl: HTMLDivElement
  let hueEl: HTMLDivElement
  let dragging: 'sv' | 'hue' | null = null

  function clamp(n: number, lo: number, hi: number): number {
    return Math.min(hi, Math.max(lo, n))
  }

  function hexToRgb(hex: string): { r: number; g: number; b: number } {
    const clean = hex.replace('#', '')
    const full = clean.length === 3 ? clean.split('').map((c) => c + c).join('') : clean
    const num = parseInt(full, 16)
    return { r: (num >> 16) & 255, g: (num >> 8) & 255, b: num & 255 }
  }

  function rgbToHex(r: number, g: number, b: number): string {
    const to = (n: number): string => clamp(Math.round(n), 0, 255).toString(16).padStart(2, '0')
    return `#${to(r)}${to(g)}${to(b)}`
  }

  function rgbToHsv(r: number, g: number, b: number): { h: number; s: number; v: number } {
    r /= 255
    g /= 255
    b /= 255
    const max = Math.max(r, g, b)
    const min = Math.min(r, g, b)
    const d = max - min
    let hue = 0
    if (d) {
      if (max === r) hue = ((g - b) / d) % 6
      else if (max === g) hue = (b - r) / d + 2
      else hue = (r - g) / d + 4
      hue *= 60
      if (hue < 0) hue += 360
    }
    return { h: hue, s: max === 0 ? 0 : (d / max) * 100, v: max * 100 }
  }

  function hsvToRgb(hue: number, sat: number, val: number): { r: number; g: number; b: number } {
    const sn = sat / 100
    const vn = val / 100
    const c = vn * sn
    const x = c * (1 - Math.abs(((hue / 60) % 2) - 1))
    const m = vn - c
    let r = 0
    let g = 0
    let b = 0
    if (hue < 60) {
      r = c
      g = x
    } else if (hue < 120) {
      r = x
      g = c
    } else if (hue < 180) {
      g = c
      b = x
    } else if (hue < 240) {
      g = x
      b = c
    } else if (hue < 300) {
      r = x
      b = c
    } else {
      r = c
      b = x
    }
    return { r: (r + m) * 255, g: (g + m) * 255, b: (b + m) * 255 }
  }

  function currentHex(): string {
    const { r, g, b } = hsvToRgb(h, s, v)
    return rgbToHex(r, g, b)
  }

  // sync internal hsv when the value prop changes from outside (not our own emit).
  $: syncFromValue(value)
  function syncFromValue(hex: string): void {
    if (hex === lastHex || !isValidHex(hex)) {
      return
    }
    const { r, g, b } = hexToRgb(normalizeHex(hex))
    const hsv = rgbToHsv(r, g, b)
    h = hsv.h
    s = hsv.s
    v = hsv.v
    hexText = normalizeHex(hex)
  }

  function emit(): void {
    const hex = currentHex()
    lastHex = hex
    hexText = hex
    dispatch('change', hex)
  }

  function onHexInput(): void {
    if (isValidHex(hexText)) {
      const { r, g, b } = hexToRgb(normalizeHex(hexText))
      const hsv = rgbToHsv(r, g, b)
      h = hsv.h
      s = hsv.s
      v = hsv.v
      emit()
    }
  }

  function pointSV(event: PointerEvent): void {
    const rect = svEl.getBoundingClientRect()
    s = clamp((event.clientX - rect.left) / rect.width, 0, 1) * 100
    v = (1 - clamp((event.clientY - rect.top) / rect.height, 0, 1)) * 100
    emit()
  }

  function pointHue(event: PointerEvent): void {
    const rect = hueEl.getBoundingClientRect()
    h = clamp((event.clientX - rect.left) / rect.width, 0, 1) * 360
    emit()
  }

  function svDown(event: PointerEvent): void {
    dragging = 'sv'
    svEl.setPointerCapture(event.pointerId)
    pointSV(event)
  }

  function hueDown(event: PointerEvent): void {
    dragging = 'hue'
    hueEl.setPointerCapture(event.pointerId)
    pointHue(event)
  }
</script>

<div class="picker">
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div
    class="sv"
    bind:this={svEl}
    style={`--hue:${h}`}
    on:pointerdown={svDown}
    on:pointermove={(e) => dragging === 'sv' && pointSV(e)}
    on:pointerup={() => (dragging = null)}
  >
    <span class="sv-thumb" style={`left:${s}%;top:${100 - v}%`}></span>
  </div>

  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div
    class="hue"
    bind:this={hueEl}
    on:pointerdown={hueDown}
    on:pointermove={(e) => dragging === 'hue' && pointHue(e)}
    on:pointerup={() => (dragging = null)}
  >
    <span class="hue-thumb" style={`left:${(h / 360) * 100}%`}></span>
  </div>

  <div class="hex">
    <span class="preview" style={`background:${currentHex()}`}></span>
    <input
      type="text"
      bind:value={hexText}
      on:input={onHexInput}
      spellcheck="false"
      aria-label={$t('common.colorPicker.hexLabel')}
      placeholder="#RRGGBB"
    />
  </div>
</div>

<style>
  .picker {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    padding: var(--space-3);
    width: 240px;
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-overlay);
    box-shadow: var(--shadow-overlay);
  }

  /* saturation/value square: hue base with white->transparent (saturation) and
     transparent->black (value) overlays. */
  .sv {
    position: relative;
    height: 140px;
    border-radius: var(--radius-control);
    background: hsl(var(--hue) 100% 50%);
    cursor: crosshair;
    touch-action: none;
  }

  .sv::before,
  .sv::after {
    content: '';
    position: absolute;
    inset: 0;
    border-radius: var(--radius-control);
  }

  .sv::before {
    background: linear-gradient(to right, #fff, transparent);
  }

  .sv::after {
    background: linear-gradient(to top, #000, transparent);
  }

  .sv-thumb,
  .hue-thumb {
    position: absolute;
    width: 12px;
    height: 12px;
    border-radius: 999px;
    border: 2px solid #fff;
    box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.4);
    transform: translate(-50%, -50%);
    pointer-events: none;
    z-index: 1;
  }

  .hue {
    position: relative;
    height: 12px;
    border-radius: 999px;
    cursor: pointer;
    touch-action: none;
    background: linear-gradient(
      to right,
      #f00 0%,
      #ff0 17%,
      #0f0 33%,
      #0ff 50%,
      #00f 67%,
      #f0f 83%,
      #f00 100%
    );
  }

  .hue-thumb {
    top: 50%;
  }

  .hex {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }

  .preview {
    width: 24px;
    height: 24px;
    border-radius: var(--radius-control);
    border: var(--hairline) solid var(--border-subtle);
    flex-shrink: 0;
  }

  .hex input {
    flex: 1;
    min-width: 0;
    height: var(--control-height);
    padding: 0 var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-sunken);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: var(--fz-label);
    outline: none;
  }

  .hex input:focus {
    border-color: var(--accent);
  }
</style>
