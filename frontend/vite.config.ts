import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  build: {
    target: 'es2020',
    manifest: true,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes('node_modules')) return
          if (id.includes('chart.js') || id.includes('vue-chartjs')) return 'charts'
          if (id.includes('@auth0')) return 'auth'
          if (/(@codemirror\/lang-|@lezer\/(markdown|html|css|javascript|json))/.test(id)) return 'editor-languages'
          if (id.includes('codemirror') || id.includes('@lezer') || id.includes('style-mod') || id.includes('w3c-keyname')) return 'editor-core'
        },
      },
    },
  }
})
