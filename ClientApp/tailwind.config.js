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
          'deepslate': '#2c3e50',
          'water': '#3498db',
          'lava': '#e67e22',
          'gold': '#FFD700',
          'important-red': '#ff5252',
        },
      },
      fontFamily: {
        'minecraft': ['"Minecraft"', '"Press Start 2P"', 'monospace'],
        'pixel': ['"Press Start 2P"', 'monospace'],
      },
      animation: {
        'fade-in': 'fadeIn 0.5s ease-in-out',
        'slide-up': 'slideUp 0.5s ease-in-out',
        'pulse': 'pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite',
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
      },
    },
  },
  plugins: [],
} 