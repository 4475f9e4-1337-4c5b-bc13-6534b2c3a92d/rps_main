/** @type {import('tailwindcss').Config} */
export const content = [
  "./internal/templates/**/*.templ",
  "./internal/templates/*.templ",
];

export const plugins = [
  require("@tailwindcss/forms"),
  require("@tailwindcss/typography"),
];
