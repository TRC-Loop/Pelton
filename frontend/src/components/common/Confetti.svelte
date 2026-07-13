<script lang="ts">
  // a one-shot celebratory confetti burst on a full-viewport canvas. it draws
  // itself with requestAnimationFrame, respects the device pixel ratio, and
  // pulls its palette from the theme tokens so it matches the chosen accent. it
  // stops on its own once every piece has fallen away, then cleans up. purely
  // decorative; pointer-events are off so it never blocks the ui.
  import { onMount } from 'svelte'

  // run controls a single burst. toggle it false->true to fire again.
  export let run = true

  let canvas: HTMLCanvasElement
  let raf = 0

  // purely decorative, so it skips entirely under reduced motion (the in-app
  // setting marks the root; the media query covers the os preference).
  function motionReduced(): boolean {
    return (
      document.documentElement.hasAttribute('data-reduce-motion') ||
      window.matchMedia('(prefers-reduced-motion: reduce)').matches
    )
  }

  interface Piece {
    x: number
    y: number
    vx: number
    vy: number
    size: number
    rot: number
    vrot: number
    color: string
    life: number
  }

  function palette(): string[] {
    const root = getComputedStyle(document.documentElement)
    const tokens = ['--accent', '--success', '--warning', '--link', '--text-primary']
    const colors = tokens.map((t) => root.getPropertyValue(t).trim()).filter(Boolean)
    return colors.length > 0 ? colors : ['#465AF2']
  }

  function spawn(width: number, height: number): Piece[] {
    const colors = palette()
    const pieces: Piece[] = []
    // two angled fountains from the lower corners give a fuller, livelier burst
    // than a single source.
    const sources = [
      { x: width * 0.2, y: height + 10, dir: -1 },
      { x: width * 0.8, y: height + 10, dir: 1 },
    ]
    for (const s of sources) {
      for (let i = 0; i < 90; i++) {
        const angle = -Math.PI / 2 + s.dir * (Math.random() * 0.6 - 0.1)
        const speed = 9 + Math.random() * 9
        pieces.push({
          x: s.x,
          y: s.y,
          vx: Math.cos(angle) * speed,
          vy: Math.sin(angle) * speed,
          size: 5 + Math.random() * 6,
          rot: Math.random() * Math.PI,
          vrot: (Math.random() - 0.5) * 0.4,
          color: colors[Math.floor(Math.random() * colors.length)],
          life: 1,
        })
      }
    }
    return pieces
  }

  function start(): void {
    if (!canvas || motionReduced()) {
      return
    }
    const ctx = canvas.getContext('2d')
    if (!ctx) {
      return
    }
    const dpr = Math.min(window.devicePixelRatio || 1, 2)
    const width = window.innerWidth
    const height = window.innerHeight
    canvas.width = width * dpr
    canvas.height = height * dpr
    ctx.scale(dpr, dpr)

    let pieces = spawn(width, height)
    const gravity = 0.28
    const drag = 0.992

    const frame = (): void => {
      ctx.clearRect(0, 0, width, height)
      let alive = 0
      for (const p of pieces) {
        p.vy += gravity
        p.vx *= drag
        p.x += p.vx
        p.y += p.vy
        p.rot += p.vrot
        // start fading once a piece is past its peak and heading down.
        if (p.vy > 2) {
          p.life -= 0.012
        }
        if (p.life > 0 && p.y < height + 40) {
          alive++
          ctx.save()
          ctx.globalAlpha = Math.max(0, p.life)
          ctx.translate(p.x, p.y)
          ctx.rotate(p.rot)
          ctx.fillStyle = p.color
          ctx.fillRect(-p.size / 2, -p.size / 2, p.size, p.size * 0.6)
          ctx.restore()
        }
      }
      if (alive > 0) {
        raf = requestAnimationFrame(frame)
      } else {
        ctx.clearRect(0, 0, width, height)
      }
    }
    cancelAnimationFrame(raf)
    raf = requestAnimationFrame(frame)
  }

  // refire whenever run flips to true.
  let last = false
  $: if (run && !last && canvas) {
    start()
  }
  $: last = run

  onMount(() => {
    if (run) {
      start()
    }
    return () => cancelAnimationFrame(raf)
  })
</script>

<canvas bind:this={canvas} class="confetti" aria-hidden="true"></canvas>

<style>
  .confetti {
    position: fixed;
    inset: 0;
    width: 100%;
    height: 100%;
    z-index: 300;
    pointer-events: none;
  }
</style>
