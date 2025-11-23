// Polyfill Buffer for gray-matter library
import { Buffer } from 'buffer'
    ; (globalThis as any).Buffer = Buffer


import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import './lib/markdown/styles.css'
import App from './App.tsx'
import './i18n/config'


createRoot(document.getElementById('root')!).render(
    <StrictMode>
        <App />
    </StrictMode>,
)
