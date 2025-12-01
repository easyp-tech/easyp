import { Globe, Layers, Lock } from 'lucide-react'
import { useTranslation } from 'react-i18next'

export default function ArchitectureComparison() {
    const { t } = useTranslation()

    const features = [
        {
            icon: Globe,
            titleKey: 'architecture.dataSovereignty.title',
            descriptionKey: 'architecture.dataSovereignty.description',
            iconBg: 'bg-blue-500/10',
            iconColor: 'text-blue-400',
            hoverBorder: 'hover:border-blue-500/30'
        },
        {
            icon: Layers,
            titleKey: 'architecture.hybridRuntime.title',
            descriptionKey: 'architecture.hybridRuntime.description',
            iconBg: 'bg-purple-500/10',
            iconColor: 'text-purple-400',
            hoverBorder: 'hover:border-purple-500/30'
        },
        {
            icon: Lock,
            titleKey: 'architecture.noVendorLock.title',
            descriptionKey: 'architecture.noVendorLock.description',
            iconBg: 'bg-emerald-500/10',
            iconColor: 'text-emerald-400',
            hoverBorder: 'hover:border-emerald-500/30'
        }
    ]

    return (
        <section className="py-32 bg-slate-950/50 relative overflow-hidden">
            {/* Background Decoration */}
            <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[800px] bg-primary/10 rounded-full blur-[150px] opacity-20 pointer-events-none"></div>

            <div className="max-w-7xl mx-auto px-6 relative z-10">
                <div className="text-center mb-20">
                    <h2 className="text-3xl md:text-5xl font-bold text-white mb-5">
                        {t('architecture.title')}
                    </h2>
                    <p className="text-slate-400 text-lg max-w-2xl mx-auto">
                        {t('architecture.subtitle')}
                    </p>
                </div>

                <div className="grid md:grid-cols-3 gap-8">
                    {features.map((feature, i) => {
                        const Icon = feature.icon

                        return (
                            <div key={i} className={`glass-panel p-8 rounded-2xl border border-white/5 ${feature.hoverBorder} transition-colors group`}>
                                <div className={`w-14 h-14 ${feature.iconBg} rounded-xl flex items-center justify-center ${feature.iconColor} mb-6 group-hover:scale-110 transition-transform`}>
                                    <Icon size={28} />
                                </div>
                                <h3 className="text-xl font-semibold text-white mb-3">{t(feature.titleKey)}</h3>
                                <p className="text-slate-400 leading-relaxed" dangerouslySetInnerHTML={{ __html: t(feature.descriptionKey).replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>') }} />
                            </div>
                        )
                    })}
                </div>
            </div>
        </section>
    )
}
