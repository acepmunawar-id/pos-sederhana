import type { Config } from 'tailwindcss';

/**
 * Design tokens shared with the mobile app (iOS system palette).
 * Colour values live as CSS variables in src/index.css so Light/Dark mode
 * switch automatically; components only use these token names.
 */
const rgb = (name: string) => `rgb(var(--${name}) / <alpha-value>)`;

export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  darkMode: 'media',
  theme: {
    extend: {
      colors: {
        primary: rgb('primary'),
        success: rgb('success'),
        danger: rgb('danger'),
        warning: rgb('warning'),
        yellow: rgb('yellow'),
        background: rgb('background'),
        card: rgb('card'),
        gray: rgb('gray'),
        separator: rgb('separator'),
        text: rgb('text'),
        // Already translucent per the iOS spec, so no alpha modifier.
        fill: 'var(--fill)',
        'text-secondary': 'var(--text-secondary)',
      },
      borderRadius: {
        sm: '8px',
        md: '10px',
        lg: '12px',
        xl: '14px',
      },
      spacing: {
        sidebar: '248px',
        topbar: '60px',
        'touch': '44px',
      },
      fontFamily: {
        sans: ['-apple-system', 'BlinkMacSystemFont', '"SF Pro Text"', '"SF Pro Display"', '"Segoe UI"', 'Roboto',
          '"Helvetica Neue"', 'Arial', 'sans-serif'],
      },
      fontSize: {
        // iOS type scale (size / line-height)
        'large-title': ['34px', { lineHeight: '41px', fontWeight: '700' }],
        title: ['28px', { lineHeight: '34px', fontWeight: '700' }],
        headline: ['17px', { lineHeight: '22px', fontWeight: '600' }],
        body: ['17px', '22px'],
        subheadline: ['15px', '20px'],
        footnote: ['13px', '18px'],
        caption: ['12px', '16px'],
      },
      boxShadow: {
        card: '0 1px 2px rgba(0,0,0,0.04), 0 1px 1px rgba(0,0,0,0.03)',
        modal: '0 10px 40px rgba(0,0,0,0.18)',
      },
      keyframes: {
        'fade-in': { from: { opacity: '0' }, to: { opacity: '1' } },
        'slide-up': { from: { opacity: '0', transform: 'translateY(8px)' }, to: { opacity: '1', transform: 'translateY(0)' } },
      },
      animation: {
        'fade-in': 'fade-in 150ms ease-out',
        'slide-up': 'slide-up 180ms ease-out',
      },
    },
  },
  plugins: [],
} satisfies Config;
