const colors = require('tailwindcss/colors')

function withOpacityValue(variable) {
  return ({ opacityValue }) => {
    if (opacityValue === undefined) {
      return `rgb(var(${variable}))`
    }
    return `rgb(var(${variable}) / ${opacityValue})`
  }
}

module.exports = {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        "main-bg": "#f2f4f8",
        "nav-top-bg": "#1b55e2",
        "nav-left-bg": "#1946b8",
        "nav-title-bg": "#1841a6",
        card: "#ffffff",
        primary: "#FD6E76",
        secondary: "#FD7D85",
        success: "#1d9110",
        info: "#3bb4d8",
        warning: "#f9da7b",
        error: "#b71540",
        "side-panel": "var(--color-side-panel)",
        "side-panel-text": "var(--color-side-panel-text)",
        "input-bg": "var(--color-input-bg)",
        "input-text": "var(--color-input-text)",
        "input-border": "#313c49",
        "input-bg-active": "var(--color-input-bg-active)",
        "input-text-active": "var(--color-input-text-active)",
        "input-border-active": colors.white
      },
      gridTemplateColumns: {
        '13': 'repeat(13, minmax(0, 1fr))',
        '14': 'repeat(14, minmax(0, 1fr))',
        '15': 'repeat(15, minmax(0, 1fr))',
        '16': 'repeat(16, minmax(0, 1fr))',
        '17': 'repeat(17, minmax(0, 1fr))',
      }
    },
  },
  plugins: [],
}