import { useTranslation } from 'react-i18next'

export default function TrustedBy() {
    const { t } = useTranslation()
    // Easy to edit: just add or remove company names from this array
    const companies = ['Yadro', 'RVision', 'ArenaData', 'PT']

    // Duplicate the array multiple times for seamless infinite scroll
    const duplicatedCompanies = [...companies, ...companies, ...companies, ...companies]

    return (
        <section className="py-12 border-y border-white/5 bg-slate-900/30 relative overflow-hidden">
            <div className="max-w-7xl mx-auto px-6 text-center">
                <p className="text-sm font-medium text-slate-500 mb-8 uppercase tracking-wider">{t('trustedBy.title')}</p>

                {/* Scrolling container */}
                <div className="relative">
                    {/* Fade overlay - left side */}
                    <div className="absolute left-0 top-0 bottom-0 w-[100px] bg-gradient-to-r from-background to-transparent z-10 pointer-events-none"></div>

                    {/* Fade overlay - right side */}
                    <div className="absolute right-0 top-0 bottom-0 w-[100px] bg-gradient-to-l from-background to-transparent z-10 pointer-events-none"></div>

                    {/* Scrolling content */}
                    <div className="flex items-center gap-12 animate-marquee hover:[animation-play-state:paused]">
                        {duplicatedCompanies.map((name, index) => (
                            <span
                                key={`${name}-${index}`}
                                className="text-xl font-bold text-white opacity-40 grayscale whitespace-nowrap flex-shrink-0"
                            >
                                {name}
                            </span>
                        ))}
                    </div>
                </div>
            </div>
        </section>
    )
}
