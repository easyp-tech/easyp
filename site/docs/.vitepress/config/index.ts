import {defineConfig} from 'vitepress'
import {shared} from './shared.mjs'
import {en} from './en.mjs'
import {ru} from './ru.mjs'

export default defineConfig({
    ...shared,
    locales: {
        root: {label: 'English', ...en},
        ru: {label: 'Русский', ...ru},
    }
})
