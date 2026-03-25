import { useTranslation } from 'react-i18next'
import { useState } from 'react'

interface Partner {
    name: string
    url: string
    logo?: string
}

const partners: Partner[] = [
    { name: 'comazo', url: 'https://comazo.ru/site_new/index.php', logo: '/assets/partners/comazo.svg' },
    { name: 'YADRO', url: 'https://yadro.com/', logo: '/assets/partners/yadro.png' },
    { name: 'h3', url: 'https://h3llo.cloud/', logo: '/assets/partners/h3.png' },
    { name: 'OpenIDE', url: 'https://openide.ru/', logo: '/assets/partners/openide.png' },
    { name: 'Positive Tech', url: 'https://ptsecurity.com/', logo: '/assets/partners/pt.png' },
]

function PartnerItem({ partner }: { partner: Partner }) {
    const [imgError, setImgError] = useState(false)

    return (
        <a
            href={partner.url}
            target="_blank"
            rel="noopener noreferrer"
            className="opacity-40 grayscale whitespace-nowrap flex-shrink-0 hover:opacity-70 transition-opacity"
        >
            {partner.logo && !imgError ? (
                <img
                    src={partner.logo}
                    alt={partner.name}
                    className="h-8 w-auto"
                    onError={() => setImgError(true)}
                />
            ) : (
                <span className="text-xl font-bold text-white">
                    {partner.name}
                </span>
            )}
        </a>
    )
}

export default function TrustedBy() {
    const { t } = useTranslation()

    const duplicatedPartners = [...partners, ...partners, ...partners, ...partners]

    return (
        <section className="py-12 border-y border-white/5 bg-slate-900/30 relative overflow-hidden">
            <div className="max-w-7xl mx-auto px-6 text-center">
                <p className="text-sm font-medium text-slate-500 mb-8 uppercase tracking-wider">{t('trustedBy.title')}</p>

                <div className="relative">
                    <div className="absolute left-0 top-0 bottom-0 w-[100px] bg-gradient-to-r from-background to-transparent z-10 pointer-events-none"></div>
                    <div className="absolute right-0 top-0 bottom-0 w-[100px] bg-gradient-to-l from-background to-transparent z-10 pointer-events-none"></div>

                    <div className="flex items-center gap-12 animate-marquee hover:[animation-play-state:paused]">
                        {duplicatedPartners.map((partner, index) => (
                            <PartnerItem key={`${partner.name}-${index}`} partner={partner} />
                        ))}
                    </div>
                </div>
            </div>
        </section>
    )
}
