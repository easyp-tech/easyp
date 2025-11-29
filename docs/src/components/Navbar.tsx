import { Code } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useLocation, useNavigate } from 'react-router-dom'
import LanguageSwitcher from './LanguageSwitcher'

export default function Navbar() {
    const { t } = useTranslation()
    const location = useLocation()
    const navigate = useNavigate()
    const isHomePage = location.pathname === '/'

    const scrollToTop = () => {
        if (isHomePage) {
            window.scrollTo({ top: 0, behavior: 'smooth' })
        } else {
            navigate('/')
            window.scrollTo(0, 0)
        }
    }


    return (
        <nav className="fixed top-0 w-full z-50 glass-panel border-b-0 border-white/5">
            <div className="max-w-7xl mx-auto px-6">
                <div className="flex items-center justify-between h-16">
                    {/* Logo + Docs */}
                    <div className="flex items-center gap-8">
                        <div
                            onClick={scrollToTop}
                            className="flex items-center gap-2 cursor-pointer hover:opacity-80 transition-opacity"
                        >
                            <div className="w-8 h-8 bg-primary/20 rounded-lg flex items-center justify-center border border-primary/30">
                                <span className="font-bold text-primary text-sm">EP</span>
                            </div>
                            <span className="font-semibold text-lg tracking-tight text-white">EasyP</span>
                        </div>

                        <a
                            href="/docs"
                            className="text-sm font-medium text-slate-400 hover:text-white transition-colors"
                        >
                            {t('nav.documentation')}
                        </a>

                        <a
                            href="/blog"
                            className="text-sm font-medium text-slate-400 hover:text-white transition-colors"
                        >
                            {t('nav.blog')}
                        </a>
                    </div>

                    {/* Spacer */}
                    <div className="flex-1"></div>

                    {/* CTA */}
                    <div className="flex items-center gap-4">
                        <a href="https://github.com/easyp-tech/easyp" className="hidden sm:flex items-center gap-2 text-sm font-medium text-slate-400 hover:text-white transition-colors">
                            <Code size={16} /> {t('nav.github')}
                        </a>
                        <LanguageSwitcher />
                    </div>
                </div>
            </div>
        </nav>
    )
}
