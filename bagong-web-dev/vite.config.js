import { fileURLToPath, URL } from 'url'

import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import  viteAutoRoute from 'vite-plugin-vue-auto-route'
//import { resolve } from 'path'

// https://vitejs.dev/config/
export default ({mode}) => { 
  process.env = {...process.env, ...loadEnv(mode, process.cwd())}

  return defineConfig({
    model: mode,
    plugins: [vue(),viteAutoRoute({ 
      pagesDir: './src/views', 
      excludeDirs:[(dir)=>{
        return dir?.match(/.*\/(widget)\/.*/g) != null
      }],
      home:"/HomeView"
    })],
    server: {
      fs: {
        allow: ['../../../']
      },
      port: process.env.VITE_PORT || 38001,
      host: process.env.VITE_HOST ||'0.0.0.0',
      hmr: {
        host: process.env.VITE_HMR_HOST || '0.0.0.0',
        clientPort: process.env.VITE_HMR_PORT || 36900,
        protocol: process.env.VITE_HMR_PROTOCOL || 'wss',
      }
    },
    base: '/bagong/',
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
      dedupe: [
        'vue'
      ]
    }
  })
}
