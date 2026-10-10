import { useEffect, useState } from 'react';

export interface ThemeColors {
  primary: string;
  text: string;
  textSecondary: string;
  separator: string;
  card: string;
  fill: string;
}

function read(): ThemeColors {
  const css = getComputedStyle(document.documentElement);
  const rgb = (name: string) => `rgb(${css.getPropertyValue(`--${name}`).trim()})`;
  return {
    primary: rgb('primary'),
    text: rgb('text'),
    textSecondary: css.getPropertyValue('--text-secondary').trim(),
    separator: rgb('separator'),
    card: rgb('card'),
    fill: css.getPropertyValue('--fill').trim(),
  };
}

/** Resolves design tokens to concrete colors for SVG charts; updates on light/dark switch. */
export function useThemeColors(): ThemeColors {
  const [colors, setColors] = useState(read);
  useEffect(() => {
    const mq = window.matchMedia('(prefers-color-scheme: dark)');
    const update = () => setColors(read());
    mq.addEventListener('change', update);
    return () => mq.removeEventListener('change', update);
  }, []);
  return colors;
}
