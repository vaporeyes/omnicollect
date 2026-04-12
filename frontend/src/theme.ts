// ABOUTME: Theme system with fixed light ("Archive") and dark ("Vault") palettes.
// ABOUTME: Supports light/dark/system mode switching via CSS custom properties.

export interface ThemeConfig {
  mode: 'light' | 'dark' | 'system'
}

export const DEFAULT_CONFIG: ThemeConfig = {
  mode: 'system',
}

const LIGHT_TOKENS: Record<string, string> = {
  '--bg-primary': '#F4F1EA',
  '--bg-secondary': '#FFFFFF',
  '--bg-tertiary': '#EBE8E1',
  '--bg-hover': 'rgba(0, 47, 167, 0.06)',
  '--bg-input': '#FFFFFF',
  '--text-primary': '#121212',
  '--text-secondary': '#3A3A3A',
  '--text-muted': '#6B6B6B',
  '--text-on-accent': '#FFFFFF',
  '--border-primary': '#121212',
  '--border-input': '#121212',
  '--accent-blue': '#002FA7',
  '--accent-blue-hover': '#001E6E',
  '--accent-blue-light': 'rgba(0, 47, 167, 0.08)',
  '--btn-secondary-bg': '#E2DFD8',
  '--btn-secondary-bg-hover': '#D5D2CB',
  '--error-text': '#B91C1C',
  '--error-bg': '#FEF2F2',
  '--error-border': '#DC2626',
  '--required-color': '#DC2626',
  '--success-text': '#166534',
  '--overlay-bg': 'rgba(18, 18, 18, 0.88)',
  '--shadow-sm': '2px 2px 0px #121212',
  '--shadow-md': '4px 4px 0px #121212',
}

const DARK_TOKENS: Record<string, string> = {
  '--bg-primary': '#090A0C',
  '--bg-secondary': '#16181D',
  '--bg-tertiary': '#1E2027',
  '--bg-hover': 'rgba(227, 66, 52, 0.08)',
  '--bg-input': '#16181D',
  '--text-primary': '#E8E6E1',
  '--text-secondary': '#9B9A97',
  '--text-muted': '#5C5B58',
  '--text-on-accent': '#FFFFFF',
  '--border-primary': '#2A2D35',
  '--border-input': '#2A2D35',
  '--accent-blue': '#E34234',
  '--accent-blue-hover': '#C7352A',
  '--accent-blue-light': 'rgba(227, 66, 52, 0.12)',
  '--btn-secondary-bg': '#1E2027',
  '--btn-secondary-bg-hover': '#282A32',
  '--error-text': '#F87171',
  '--error-bg': 'rgba(220, 38, 38, 0.12)',
  '--error-border': '#DC2626',
  '--required-color': '#F87171',
  '--success-text': '#4ADE80',
  '--overlay-bg': 'rgba(0, 0, 0, 0.92)',
  '--shadow-sm': '2px 2px 0px #2A2D35',
  '--shadow-md': '4px 4px 0px #2A2D35',
}

// Apply the theme tokens for the given mode.
export function applyTheme(dark: boolean) {
  const root = document.documentElement
  const tokens = dark ? DARK_TOKENS : LIGHT_TOKENS
  for (const [prop, value] of Object.entries(tokens)) {
    root.style.setProperty(prop, value)
  }
  root.classList.toggle('dark-theme', dark)
}
