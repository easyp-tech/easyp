import type { PrismTheme } from 'prism-react-renderer';

export const cyberTheme: PrismTheme = {
    plain: {
        color: '#e2e8f0', // Slate 200
        backgroundColor: 'transparent',
    },
    styles: [
        {
            types: ['comment', 'prolog', 'doctype', 'cdata'],
            style: {
                color: '#64748b', // Slate 500
                fontStyle: 'italic',
            },
        },
        {
            types: ['punctuation'],
            style: {
                color: '#94a3b8', // Slate 400
            },
        },
        {
            types: ['namespace'],
            style: {
                opacity: 0.7,
            },
        },
        {
            types: ['property', 'tag', 'boolean', 'number', 'constant', 'symbol', 'deleted'],
            style: {
                color: '#f472b6', // Pink 400
            },
        },
        {
            types: ['selector', 'attr-name', 'string', 'char', 'builtin', 'inserted'],
            style: {
                color: '#a78bfa', // Violet 400
            },
        },
        {
            types: ['operator', 'entity', 'url'],
            style: {
                color: '#60a5fa', // Blue 400
            },
        },
        {
            types: ['atrule', 'attr-value', 'keyword'],
            style: {
                color: '#818cf8', // Indigo 400
            },
        },
        {
            types: ['function', 'class-name'],
            style: {
                color: '#38bdf8', // Sky 400
            },
        },
        {
            types: ['regex', 'important', 'variable'],
            style: {
                color: '#fb7185', // Rose 400
            },
        },
    ],
};
