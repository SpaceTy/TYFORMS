/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        'minecraft': {
          'green': '#44bd32',
          'earth': '#7e5835',
          'stone': '#7f8c8d',
          'deepslate': '#1e2a36',
          'water': '#60a5fa',
          'lava': '#fb923c',
          'gold': '#FACC15',
          'important-red': '#ef4444',
        },
        primary: {
          DEFAULT: '#60a5fa',
          50: '#eff6ff',
          100: '#dbeafe',
          200: '#bfdbfe',
          300: '#93c5fd',
          400: '#60a5fa',
          500: '#3b82f6',
          600: '#2563eb',
          700: '#1d4ed8',
          800: '#1e40af',
          900: '#1e3a8a',
        },
        neutral: {
          50: '#f8fafc',
          100: '#f1f5f9',
          200: '#e2e8f0',
          300: '#cbd5e1',
          400: '#94a3b8',
          500: '#64748b',
          600: '#475569',
          700: '#334155',
          800: '#1f2937',
          900: '#111827',
        }
      },
      fontFamily: {
        'minecraft': ['"Minecraft"', '"Press Start 2P"', 'monospace'],
        'pixel': ['"Press Start 2P"', 'monospace'],
        'sans': ['Inter', 'ui-sans-serif', 'system-ui', 'Segoe UI', 'Roboto', 'Apple Color Emoji', 'Segoe UI Emoji']
      },
      borderRadius: {
        'xl': '1rem',
        '2xl': '1.25rem'
      },
      boxShadow: {
        'soft': '0 8px 24px rgba(0,0,0,0.25)',
        'glass': '0 10px 30px rgba(0,0,0,0.25), inset 0 1px 0 rgba(255,255,255,0.05)'
      },
      animation: {
        'fade-in': 'fadeIn 0.5s ease-in-out',
        'slide-up': 'slideUp 0.5s ease-in-out',
        'pulse': 'pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        'float': 'float 6s ease-in-out infinite'
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideUp: {
          '0%': { transform: 'translateY(20px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        pulse: {
          '0%, 100%': { opacity: '1', transform: 'scale(1)' },
          '50%': { opacity: '.7', transform: 'scale(1.05)' },
        },
        float: {
          '0%, 100%': { transform: 'translateY(0)' },
          '50%': { transform: 'translateY(-6px)' }
        }
      },
    },
  },
  plugins: [],
}
