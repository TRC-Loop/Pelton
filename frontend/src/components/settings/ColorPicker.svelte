<script lang="ts">
  // the theme editor's color control: a well that opens a saturation/value
  // area with hue and alpha sliders, plus typed hex and r/g/b fields. values
  // the color model cannot parse (color-mix, named colors) still round-trip:
  // the picker opens on a neutral gray and only overwrites the token once the
  // user actually picks something.
  import { createEventDispatcher, onDestroy } from 'svelte'
  import { parseColor, formatColor, toHex, rgbToHsv, hsvToRgb } from '../../theme/color'
  import type { RGBA } from '../../theme/color'
  import { t } from '../../lib/i18n'

  /** the token's current value, in any notation. */
  export let value = ''

  const dispatch = createEventDispatcher<{ change: string }>()

  const fallback: RGBA = { r: 136, g: 136, b: 136, a: 1 }

  let open = false
  let anchor: HTMLButtonElement
  let popover: HTMLDivElement
  let popTop = 0
  let popLeft = 0
  let flipped = false

  let hue = 0
  let sat = 0
  let val = 0.53
  let alpha = 1
  // the last value this picker emitted, so the incoming prop can be told
  // apart from an outside edit and the hsv state is not reset mid-drag.
  let emitted = ''

  $: syncFrom(value)

  function syncFrom(incoming: string): void {
    if (incoming === emitted) {
      return
    }
    const parsed = parseColor(incoming)
    if (!parsed) {
      return
    }
    const hsv = rgbToHsv(parsed)
    // a fully desaturated or black color carries no usable hue; keeping the
    // previous one stops the hue slider from jumping to red when the user
    // drags into a corner.
    if (hsv.s > 0 && hsv.v > 0) {
      hue = hsv.h
    }
    sat = hsv.s
    val = hsv.v
    alpha = parsed.a
  }

  $: current = { ...hsvToRgb({ h: hue, s: sat, v: val }), a: alpha }
  $: swatch = parseColor(value) ? formatColor(current) : 'transparent'
  $: hueColor = toHex(hsvToRgb({ h: hue, s: 1, v: 1 }))
  $: solid = toHex(current)

  function emit(): void {
    emitted = formatColor(current)
    dispatch('change', emitted)
  }

  function toggle(): void {
    open = !open
    if (open) {
      if (!parseColor(value)) {
        const hsv = rgbToHsv(fallback)
        hue = hsv.h
        sat = hsv.s
        val = hsv.v
        alpha = 1
      }
      place()
    }
  }

  // the popover is fixed-positioned: the settings modal scrolls its body, and
  // an absolutely positioned panel would be clipped by that overflow.
  function place(): void {
    const rect = anchor.getBoundingClientRect()
    const height = 268
    flipped = rect.bottom + height + 8 > window.innerHeight
    popTop = flipped ? rect.top - height - 6 : rect.bottom + 6
    popLeft = Math.max(8, Math.min(rect.left, window.innerWidth - 232))
  }

  function close(): void {
    open = false
  }

  function onDocPointerDown(event: PointerEvent): void {
    const target = event.target as Node
    if (open && !popover?.contains(target) && !anchor?.contains(target)) {
      close()
    }
  }

  function onDocKeydown(event: KeyboardEvent): void {
    if (open && event.key === 'Escape') {
      event.stopPropagation()
      close()
    }
  }

  // area and slider dragging share one pointer-capture pattern: the handler
  // runs on down and on every move while the pointer is held.
  function drag(node: HTMLElement, onMove: (x: number, y: number) => void) {
    function handle(event: PointerEvent): void {
      const rect = node.getBoundingClientRect()
      const x = Math.min(1, Math.max(0, (event.clientX - rect.left) / rect.width))
      const y = Math.min(1, Math.max(0, (event.clientY - rect.top) / rect.height))
      onMove(x, y)
    }
    function down(event: PointerEvent): void {
      node.setPointerCapture(event.pointerId)
      handle(event)
      event.preventDefault()
    }
    function move(event: PointerEvent): void {
      if (node.hasPointerCapture(event.pointerId)) {
        handle(event)
      }
    }
    node.addEventListener('pointerdown', down)
    node.addEventListener('pointermove', move)
    return {
      destroy(): void {
        node.removeEventListener('pointerdown', down)
        node.removeEventListener('pointermove', move)
      },
    }
  }

  function onArea(x: number, y: number): void {
    sat = x
    val = 1 - y
    emit()
  }

  function onHue(x: number): void {
    hue = x * 360
    emit()
  }

  function onAlpha(x: number): void {
    alpha = Math.round(x * 100) / 100
    emit()
  }

  function onHexInput(event: Event): void {
    const typed = (event.currentTarget as HTMLInputElement).value
    const parsed = parseColor(typed.startsWith('#') ? typed : '#' + typed)
    if (parsed) {
      const hsv = rgbToHsv(parsed)
      if (hsv.s > 0 && hsv.v > 0) {
        hue = hsv.h
      }
      sat = hsv.s
      val = hsv.v
      alpha = parsed.a
      emit()
    }
  }

  function onChannel(channel: 'r' | 'g' | 'b', event: Event): void {
    const n = parseInt((event.currentTarget as HTMLInputElement).value, 10)
    if (!Number.isFinite(n)) {
      return
    }
    const next: RGBA = { ...current, [channel]: Math.min(255, Math.max(0, n)) }
    const hsv = rgbToHsv(next)
    if (hsv.s > 0 && hsv.v > 0) {
      hue = hsv.h
    }
    sat = hsv.s
    val = hsv.v
    emit()
  }

  onDestroy(() => {
    open = false
  })
</script>

<svelte:window on:resize={() => open && place()} />
<svelte:document on:pointerdown={onDocPointerDown} on:keydown={onDocKeydown} />

<button
  type="button"
  class="well"
  bind:this={anchor}
  aria-label={$t('themeEditor.pickColor')}
  aria-expanded={open}
  on:click={toggle}
>
  <span class="swatch" style:background={swatch}></span>
</button>

{#if open}
  <div
    class="pop"
    class:flipped
    bind:this={popover}
    style:top={popTop + 'px'}
    style:left={popLeft + 'px'}
    role="dialog"
    aria-label={$t('themeEditor.pickColor')}
  >
    <div
      class="area"
      style:background="linear-gradient(to top, #000, transparent), linear-gradient(to right, #fff, transparent), {hueColor}"
      use:drag={onArea}
    >
      <span class="dot" style:left={sat * 100 + '%'} style:top={(1 - val) * 100 + '%'} style:background={solid}></span>
    </div>

    <div class="slider hue" use:drag={(x) => onHue(x)}>
      <span class="knob" style:left={(hue / 360) * 100 + '%'} style:background={hueColor}></span>
    </div>

    <div class="slider alpha" style:--solid={solid} use:drag={(x) => onAlpha(x)}>
      <span class="knob" style:left={alpha * 100 + '%'} style:background={solid}></span>
    </div>

    <div class="fields">
      <label class="f hex">
        <span>{$t('themeEditor.hex')}</span>
        <input type="text" class="mono" value={toHex(current)} spellcheck="false" on:input={onHexInput} />
      </label>
      <label class="f">
        <span>R</span>
        <input type="number" min="0" max="255" value={current.r} on:input={(e) => onChannel('r', e)} />
      </label>
      <label class="f">
        <span>G</span>
        <input type="number" min="0" max="255" value={current.g} on:input={(e) => onChannel('g', e)} />
      </label>
      <label class="f">
        <span>B</span>
        <input type="number" min="0" max="255" value={current.b} on:input={(e) => onChannel('b', e)} />
      </label>
    </div>
  </div>
{/if}

<style>
  .well {
    width: 28px;
    height: 24px;
    padding: 2px;
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    /* the checkerboard shows through a translucent token value */
    background-color: #ffffff;
    background-image:
      linear-gradient(45deg, #cccccc 25%, transparent 25%, transparent 75%, #cccccc 75%),
      linear-gradient(45deg, #cccccc 25%, transparent 25%, transparent 75%, #cccccc 75%);
    background-size: 8px 8px;
    background-position: 0 0, 4px 4px;
    cursor: pointer;
  }

  .swatch {
    display: block;
    width: 100%;
    height: 100%;
    border-radius: 3px;
  }

  .pop {
    position: fixed;
    z-index: 200;
    width: 224px;
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    padding: var(--space-3);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-card);
    background: var(--surface-overlay);
    box-shadow: var(--shadow-overlay);
  }

  .area {
    position: relative;
    height: 120px;
    border-radius: var(--radius-control);
    cursor: crosshair;
    touch-action: none;
  }

  .dot {
    position: absolute;
    width: 12px;
    height: 12px;
    margin: -6px 0 0 -6px;
    border: 2px solid #ffffff;
    border-radius: 50%;
    box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.4);
    pointer-events: none;
  }

  .slider {
    position: relative;
    height: 12px;
    border-radius: 999px;
    cursor: pointer;
    touch-action: none;
  }

  .hue {
    background: linear-gradient(to right, #f00, #ff0, #0f0, #0ff, #00f, #f0f, #f00);
  }

  .alpha {
    background-color: #ffffff;
    background-image:
      linear-gradient(to right, transparent, var(--solid)),
      linear-gradient(45deg, #cccccc 25%, transparent 25%, transparent 75%, #cccccc 75%),
      linear-gradient(45deg, #cccccc 25%, transparent 25%, transparent 75%, #cccccc 75%);
    background-size: 100% 100%, 8px 8px, 8px 8px;
    background-position: 0 0, 0 0, 4px 4px;
  }

  .knob {
    position: absolute;
    top: 50%;
    width: 14px;
    height: 14px;
    margin: -7px 0 0 -7px;
    border: 2px solid #ffffff;
    border-radius: 50%;
    box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.4);
    pointer-events: none;
  }

  .fields {
    display: grid;
    grid-template-columns: 1.6fr 1fr 1fr 1fr;
    gap: var(--space-1);
  }

  .f {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .f span {
    font-size: var(--fz-meta);
    color: var(--text-tertiary);
  }

  .f input {
    width: 100%;
    min-width: 0;
    height: 22px;
    padding: 0 var(--space-1);
    border: var(--hairline) solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    color: var(--text-primary);
    font-size: var(--fz-meta);
  }

  .mono {
    font-family: var(--font-mono);
  }
</style>
