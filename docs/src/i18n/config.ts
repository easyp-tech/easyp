import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import en from './locales/en.json'
import ru from './locales/ru.json'

// Get saved language from localStorage; else infer from browser; else default to English
const getSavedLanguage = (): string => {
    const saved = localStorage.getItem('language')
    if (saved) return saved

    // Infer from browser preferences if available
    const navLangs = typeof navigator !== 'undefined'
        ? (navigator.languages || [navigator.language]).map(l => l.toLowerCase())
        : []

    if (navLangs.some(l => l === 'ru' || l.startsWith('ru-'))) {
        return 'ru'
    }

    return 'en'
}

i18n
    .use(initReactI18next)
    .init({
        resources: {
            en: { translation: en },
            ru: { translation: ru }
        },
        lng: getSavedLanguage(),
        fallbackLng: 'en',
        interpolation: {
            escapeValue: false // React already escapes
        }
    })

// Save language preference when it changes
i18n.on('languageChanged', (lng) => {
    localStorage.setItem('language', lng)
})

export default i18n
