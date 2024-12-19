import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import viteCompression from 'vite-plugin-compression'
import basicSSL from '@vitejs/plugin-basic-ssl'

// https://vitejs.dev/config/
export default defineConfig({
  server: {
    https: true,
    proxy: {
      '/api': {
        target: 'https://repman.marie-dev.svc.cloud18:10005/',
        secure: false
      }
    }
  },
  plugins: [react(), viteCompression({ algorithm: 'gzip' }), basicSSL()],
  css: {
    preprocessorOptions: {
      scss: {
        additionalData: `@import './src/styles/_mixins.scss';
         @import './src/styles/_variables.scss';
         @import './src/styles/_lighttheme.scss'; 
         @import './src/styles/_darktheme.scss';
         @import './src/styles/_global.scss';`
      }
    }
  }
})
