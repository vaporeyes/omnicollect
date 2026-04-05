// ABOUTME: Theme generation using poline for perceptually uniform color palettes.
// ABOUTME: Supports dynamic anchors, numPoints, hue shifting, and light/dark modes.
import {Poline} from 'poline'

export interface ThemeAnchor {
  hue: number       // 0-360
  saturation: number // 0-1
  lightness: number  // 0-1
}

export interface ThemeConfig {
  lightAnchors: ThemeAnchor[]
  darkAnchors: ThemeAnchor[]
  numPoints: number    // palette size (6-20)
  hueShift: number     // degrees to shift all hues (-180 to 180)
  mode: 'light' | 'dark' | 'system'
}

export const DEFAULT_CONFIG: ThemeConfig = {
  lightAnchors: [
    {hue: 220, saturation: 0.15, lightness: 0.96},
    {hue: 200, saturation: 0.55, lightness: 0.45},
    {hue: 35, saturation: 0.65, lightness: 0.55},
  ],
  darkAnchors: [
    {hue: 220, saturation: 0.25, lightness: 0.12},
    {hue: 200, saturation: 0.50, lightness: 0.55},
    {hue: 35, saturation: 0.55, lightness: 0.65},
  ],
  numPoints: 12,
  hueShift: 0,
  mode: 'system',
}

// Preset palettes for quick selection
export const PRESETS: Record<string, Omit<ThemeConfig, 'mode'>> = {
  'Warm Teal': {
    lightAnchors: [
      {hue: 220, saturation: 0.15, lightness: 0.96},
      {hue: 200, saturation: 0.55, lightness: 0.45},
      {hue: 35, saturation: 0.65, lightness: 0.55},
    ],
    darkAnchors: [
      {hue: 220, saturation: 0.25, lightness: 0.12},
      {hue: 200, saturation: 0.50, lightness: 0.55},
      {hue: 35, saturation: 0.55, lightness: 0.65},
    ],
    numPoints: 12,
    hueShift: 0,
  },
  'Ocean Depth': {
    lightAnchors: [
      {hue: 210, saturation: 0.12, lightness: 0.97},
      {hue: 220, saturation: 0.60, lightness: 0.48},
      {hue: 260, saturation: 0.45, lightness: 0.55},
    ],
    darkAnchors: [
      {hue: 215, saturation: 0.30, lightness: 0.10},
      {hue: 220, saturation: 0.50, lightness: 0.55},
      {hue: 260, saturation: 0.40, lightness: 0.65},
    ],
    numPoints: 12,
    hueShift: 0,
  },
  'Forest': {
    lightAnchors: [
      {hue: 100, saturation: 0.10, lightness: 0.96},
      {hue: 155, saturation: 0.45, lightness: 0.40},
      {hue: 45, saturation: 0.55, lightness: 0.50},
    ],
    darkAnchors: [
      {hue: 120, saturation: 0.20, lightness: 0.11},
      {hue: 155, saturation: 0.40, lightness: 0.50},
      {hue: 45, saturation: 0.45, lightness: 0.60},
    ],
    numPoints: 12,
    hueShift: 0,
  },
  'Rose Gold': {
    lightAnchors: [
      {hue: 20, saturation: 0.12, lightness: 0.96},
      {hue: 350, saturation: 0.50, lightness: 0.50},
      {hue: 30, saturation: 0.60, lightness: 0.55},
    ],
    darkAnchors: [
      {hue: 10, saturation: 0.20, lightness: 0.12},
      {hue: 350, saturation: 0.45, lightness: 0.55},
      {hue: 30, saturation: 0.50, lightness: 0.60},
    ],
    numPoints: 12,
    hueShift: 0,
  },
  'Monochrome': {
    lightAnchors: [
      {hue: 0, saturation: 0.0, lightness: 0.96},
      {hue: 0, saturation: 0.0, lightness: 0.40},
      {hue: 0, saturation: 0.0, lightness: 0.60},
    ],
    darkAnchors: [
      {hue: 0, saturation: 0.0, lightness: 0.10},
      {hue: 0, saturation: 0.0, lightness: 0.55},
      {hue: 0, saturation: 0.0, lightness: 0.70},
    ],
    numPoints: 12,
    hueShift: 0,
  },
}

function anchorsToVectors(anchors: ThemeAnchor[], hueShift: number): [number, number, number][] {
  return anchors.map(a => [
    (a.hue + hueShift + 360) % 360,
    a.saturation,
    a.lightness,
  ])
}

function generatePalette(anchors: [number, number, number][], numPoints: number): string[] {
  const poline = new Poline({
    anchorColors: anchors,
    numPoints: Math.max(6, Math.min(20, numPoints)),
  })
  return poline.colorsCSS
}

function hslToComponents(hslStr: string): {h: number; s: number; l: number} | null {
  const match = hslStr.match(/hsl\((\d+(?:\.\d+)?),\s*(\d+(?:\.\d+)?)%,\s*(\d+(?:\.\d+)?)%\)/)
  if (!match) return null
  return {h: parseFloat(match[1]), s: parseFloat(match[2]), l: parseFloat(match[3])}
}

function hsl(h: number, s: number, l: number): string {
  return `hsl(${Math.round(h)}, ${Math.round(s)}%, ${Math.round(l)}%)`
}

function buildLightTokens(colors: string[]) {
  const bg = hslToComponents(colors[0])!
  const accent = hslToComponents(colors[Math.floor(colors.length / 2)])!

  return {
    '--bg-primary': hsl(bg.h, Math.max(bg.s - 5, 0), Math.min(bg.l + 2, 99)),
    '--bg-secondary': `hsla(${Math.round(bg.h)}, ${Math.round(bg.s)}%, ${Math.round(bg.l - 2)}%, 0.65)`,
    '--bg-tertiary': `hsla(${Math.round(bg.h)}, ${Math.round(bg.s + 2)}%, ${Math.round(bg.l - 6)}%, 0.55)`,
    '--bg-hover': `hsla(${Math.round(accent.h)}, ${Math.round(accent.s)}%, ${Math.round(accent.l)}%, 0.06)`,
    '--bg-input': hsl(bg.h, Math.max(bg.s - 3, 0), Math.min(bg.l + 1, 100)),
    '--text-primary': hsl(bg.h, 15, 18),
    '--text-secondary': hsl(bg.h, 10, 40),
    '--text-muted': hsl(bg.h, 8, 58),
    '--text-on-accent': hsl(bg.h, 5, 99),
    '--border-primary': `hsla(${Math.round(bg.h)}, ${Math.round(bg.s + 5)}%, ${Math.round(bg.l - 12)}%, 0.25)`,
    '--border-input': hsl(bg.h, bg.s + 3, bg.l - 18),
    '--accent-blue': hsl(accent.h, accent.s, accent.l),
    '--accent-blue-hover': hsl(accent.h, accent.s + 5, accent.l - 8),
    '--accent-blue-light': `hsla(${Math.round(accent.h)}, ${Math.round(accent.s)}%, ${Math.round(accent.l)}%, 0.1)`,
    '--btn-secondary-bg': hsl(bg.h, bg.s + 3, bg.l - 8),
    '--btn-secondary-bg-hover': hsl(bg.h, bg.s + 5, bg.l - 14),
    '--error-text': 'hsl(0, 65%, 45%)',
    '--error-bg': 'hsl(0, 70%, 95%)',
    '--error-border': 'hsl(0, 60%, 80%)',
    '--required-color': 'hsl(0, 65%, 50%)',
    '--success-text': 'hsl(150, 50%, 38%)',
    '--overlay-bg': `hsla(${Math.round(bg.h)}, 20%, 10%, 0.85)`,
    '--shadow-sm': `0 1px 3px hsla(${Math.round(bg.h)}, 10%, 20%, 0.08), 0 1px 2px hsla(${Math.round(bg.h)}, 10%, 20%, 0.06)`,
    '--shadow-md': `0 4px 12px hsla(${Math.round(bg.h)}, 10%, 20%, 0.12), 0 2px 4px hsla(${Math.round(bg.h)}, 10%, 20%, 0.08)`,
  }
}

function buildDarkTokens(colors: string[]) {
  const bg = hslToComponents(colors[0])!
  const accent = hslToComponents(colors[Math.floor(colors.length / 2)])!

  return {
    '--bg-primary': hsl(bg.h, bg.s, bg.l),
    '--bg-secondary': `hsla(${Math.round(bg.h)}, ${Math.round(bg.s - 3)}%, ${Math.round(bg.l + 5)}%, 0.65)`,
    '--bg-tertiary': `hsla(${Math.round(bg.h)}, ${Math.round(bg.s - 5)}%, ${Math.round(bg.l + 12)}%, 0.55)`,
    '--bg-hover': `hsla(${Math.round(accent.h)}, ${Math.round(accent.s)}%, ${Math.round(accent.l)}%, 0.08)`,
    '--bg-input': hsl(bg.h, bg.s - 2, bg.l + 6),
    '--text-primary': hsl(bg.h, 10, 90),
    '--text-secondary': hsl(bg.h, 8, 65),
    '--text-muted': hsl(bg.h, 6, 48),
    '--text-on-accent': hsl(bg.h, 5, 99),
    '--border-primary': `hsla(${Math.round(bg.h)}, ${Math.round(bg.s - 5)}%, ${Math.round(bg.l + 16)}%, 0.25)`,
    '--border-input': hsl(bg.h, bg.s - 5, bg.l + 20),
    '--accent-blue': hsl(accent.h, accent.s, accent.l),
    '--accent-blue-hover': hsl(accent.h, accent.s + 5, accent.l + 8),
    '--accent-blue-light': `hsla(${Math.round(accent.h)}, ${Math.round(accent.s)}%, ${Math.round(accent.l)}%, 0.12)`,
    '--btn-secondary-bg': hsl(bg.h, bg.s - 3, bg.l + 10),
    '--btn-secondary-bg-hover': hsl(bg.h, bg.s - 2, bg.l + 18),
    '--error-text': 'hsl(0, 60%, 65%)',
    '--error-bg': 'hsla(0, 60%, 50%, 0.12)',
    '--error-border': 'hsl(0, 50%, 50%)',
    '--required-color': 'hsl(0, 60%, 65%)',
    '--success-text': 'hsl(150, 45%, 55%)',
    '--overlay-bg': `hsla(${Math.round(bg.h)}, 30%, 5%, 0.92)`,
    '--shadow-sm': `0 1px 3px hsla(${Math.round(bg.h)}, 20%, 5%, 0.3), 0 1px 2px hsla(${Math.round(bg.h)}, 20%, 5%, 0.2)`,
    '--shadow-md': `0 4px 12px hsla(${Math.round(bg.h)}, 20%, 5%, 0.4), 0 2px 4px hsla(${Math.round(bg.h)}, 20%, 5%, 0.3)`,
  }
}

// Apply the poline-generated theme for the given mode and config.
export function applyPolineTheme(dark: boolean, config: ThemeConfig = DEFAULT_CONFIG) {
  const root = document.documentElement
  const anchors = dark ? config.darkAnchors : config.lightAnchors
  const vectors = anchorsToVectors(anchors, config.hueShift)
  const colors = generatePalette(vectors, config.numPoints)
  const tokens = dark ? buildDarkTokens(colors) : buildLightTokens(colors)

  for (const [prop, value] of Object.entries(tokens)) {
    root.style.setProperty(prop, value)
  }
  root.classList.toggle('dark-theme', dark)
}

// Get the generated palette colors as CSS strings (for preview swatches).
export function getPreviewColors(config: ThemeConfig, dark: boolean): string[] {
  const anchors = dark ? config.darkAnchors : config.lightAnchors
  const vectors = anchorsToVectors(anchors, config.hueShift)
  return generatePalette(vectors, config.numPoints)
}
