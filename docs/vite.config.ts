import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
    plugins: [react()],
    define: {
        // Provide Buffer polyfill for gray-matter library
        global: 'globalThis',
    },
    resolve: {
        alias: {
            // Use buffer polyfill for browser
            buffer: 'buffer',
        },
    },
    optimizeDeps: {
        esbuildOptions: {
            // Node.js global to browser globalThis
            define: {
                global: 'globalThis',
            },
        },
    },
})
