import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueJsx from '@vitejs/plugin-vue-jsx'
import vueDevTools from 'vite-plugin-vue-devtools'

const isDev = process.env.NODE_ENV === 'development';

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueJsx(),
    vueDevTools(),
  ],
  build: {
    sourcemap: isDev // Enable sourcemaps only in development
  },
  css: {
    devSourcemap: isDev // Enable CSS sourcemaps only in development
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      'vue-weather-widget': 'vue-weather-widget'
    },
  },
})
