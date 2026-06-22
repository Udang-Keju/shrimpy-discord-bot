// dashboard/lib/theme.ts

export type Theme = 'dark' | 'light';

export function getSystemTheme(): Theme {
  if (typeof window === 'undefined') return 'dark';
  return window.matchMedia('(prefers-color-scheme: dark)').matches
    ? 'dark'
    : 'light';
}

export function applyTheme(theme: Theme) {
  document.documentElement.setAttribute('data-theme', theme);
  localStorage.setItem('Shrimpy-theme', theme);
}

export function getSavedTheme(): Theme {
  if (typeof window === 'undefined') return 'dark';
  const saved = localStorage.getItem('Shrimpy-theme') as Theme | null;
  return saved ?? getSystemTheme();
}
