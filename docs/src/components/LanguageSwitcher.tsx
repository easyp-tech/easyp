import { useState, useRef, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { Globe, Check } from 'lucide-react'

export default function LanguageSwitcher() {
    const { i18n } = useTranslation()
    const [isOpen, setIsOpen] = useState(false)
    const dropdownRef = useRef<HTMLDivElement>(null)

    const languages = [
        { code: 'en', label: 'English', disabled: false },
        { code: 'ru', label: 'Русский', disabled: false }
    ]

    const currentLanguage = languages.find(lang => lang.code === i18n.language) || languages[0]

    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
                setIsOpen(false)
            }
        }
        document.addEventListener('mousedown', handleClickOutside)
        return () => document.removeEventListener('mousedown', handleClickOutside)
    }, [])

    const changeLanguage = (langCode: string, isDisabled: boolean) => {
        if (isDisabled) return // Don't allow changing to disabled language
        i18n.changeLanguage(langCode)
        setIsOpen(false)
    }

    return (
        <div className="relative" ref={dropdownRef}>
            <button
                onClick={() => setIsOpen(!isOpen)}
                className="flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-slate-800 transition-colors text-slate-400 hover:text-white"
                title="Change language"
            >
                <Globe size={18} />
                <span className="text-sm font-medium">{currentLanguage.code.toUpperCase()}</span>
            </button>

            {isOpen && (
                <div className="absolute right-0 top-full mt-2 w-40 bg-slate-900 border border-slate-800 rounded-xl shadow-2xl overflow-hidden z-50">
                    <div className="py-1">
                        {languages.map(lang => (
                            <button
                                key={lang.code}
                                onClick={() => changeLanguage(lang.code, lang.disabled)}
                                disabled={lang.disabled}
                                className={`w-full text-left px-4 py-3 text-sm transition-colors flex items-center justify-between ${i18n.language === lang.code
                                        ? 'bg-primary/10 text-primary font-medium'
                                        : lang.disabled
                                            ? 'text-slate-600 cursor-not-allowed opacity-50'
                                            : 'text-slate-300 hover:bg-slate-800 hover:text-white'
                                    }`}
                            >
                                <span className={lang.disabled ? 'line-through' : ''}>{lang.label}</span>
                                {i18n.language === lang.code && !lang.disabled && <Check size={16} className="text-primary" />}
                            </button>
                        ))}
                    </div>
                </div>
            )}
        </div>
    )
}
