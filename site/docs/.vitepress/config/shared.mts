import {defineConfig} from 'vitepress'

export const shared = defineConfig({
    title: 'EasyP',

    lastUpdated: true,
    cleanUrls: true,
    metaChunk: true,

    markdown: {
        math: true,
        lineNumbers: true,
        codeTransformers: [
            // We use `[!!code` in demo to prevent transformation, here we revert it back.
            {
                postprocess(code) {
                    return code.replace(/\[\!\!code/g, '[!code')
                }
            }
        ]
    },

    sitemap: {
        hostname: 'https://easyp.tech',
        transformItems(items) {
            return items.filter((item) => !item.url.includes('migration'))
        }
    },

    /* prettier-ignore */
    head: [
        ['link', {rel: 'icon', type: 'image/svg+xml', href: '/logo-mini.svg'}],
        ['link', {rel: 'icon', type: 'image/png', href: '/logo-mini.png'}],
        ['meta', {name: 'theme-color', content: '#5f67ee'}],
        ['meta', {property: 'og:type', content: 'website'}],
        ['meta', {property: 'og:locale', content: 'en'}],
        ['meta', {property: 'og:title', content: 'EasyP | Easy modern protobuf tooling for all your needs.'}],
        ['meta', {property: 'og:site_name', content: 'EasyP'}],
        ['meta', {property: 'og:url', content: 'https://easyp.tech/'}],
        ['meta', {
            property: 'og:image',
            content: 'https://files.easyp.tech/Screenshot%202024-05-12%20at%2012.49.50.png'
        }],
        ['script', {src: 'https://cdn.usefathom.com/script.js', 'data-site': 'AZBRSFGG', 'data-spa': 'auto', defer: ''}]
    ],

    themeConfig: {
        logo: {src: '/logo-mini.svg', width: 24, height: 24},

        socialLinks: [
            {icon: 'github', link: 'https://github.com/easyp-tech/easyp'}
        ],

        carbonAds: {code: 'CEBDT27Y', placement: 'easyp-tech'}
    },

    vite: {
        plugins: [
            {
                name: 'patch-vitepress-symbol',
                transform(code, id) {
                    if (id.includes('vitepress/dist/client/app/data.js')) {
                        return code.replace(
                            'const dataSymbol = Symbol();',
                            'const dataSymbol = "__vitepress_data__";'
                        )
                    }
                }
            }
        ]
    },
})
