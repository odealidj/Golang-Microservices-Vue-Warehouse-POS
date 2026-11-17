 /** @type {import('tailwindcss').Config} */
 export default {
    content: [
      "./index.html",
      "./src/**/*.{vue,js,ts,jsx,tsx}",
    ],
    theme: {
        extend: {
          colors: {
            "monday-blue": "#1053D5",
            "monday-black": "#292D32",
            "monday-gray": "#6A7686",
            "monday-border": "#E8E8E8",
            "monday-red": "#FF5070",
            "monday-background": "#F3F5F9",
            "monday-gray-background": "#F3F3F3",
            "monday-lime-green": "#CBEE5A",
            "monday-lime-green-char": "#DEFF6E",
            "lte-sidebar": "#343a40",
            "lte-sidebar-dark": "#2c3136",
            "lte-sidebar-hover": "#494e53",
            "lte-sidebar-border": "#4b545c",
            "lte-sidebar-text": "#c2c7d0",
            "lte-primary": "#007bff",
            "lte-primary-hover": "#0069d9",
            "lte-info": "#17a2b8",
            "lte-success": "#28a745",
            "lte-warning": "#ffc107",
            "lte-danger": "#dc3545",
            "lte-bg": "#f4f6f9",
            "lte-border": "#dee2e6",
            "lte-text": "#212529",
            "lte-muted": "#6c757d",
          },
        },
    },
    plugins: [],
}