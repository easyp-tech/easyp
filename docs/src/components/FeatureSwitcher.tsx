import { useState } from 'react'
import { Package, CheckCircle, Shield, Box } from 'lucide-react'
import { useTranslation } from 'react-i18next'

export default function FeatureSwitcher() {
    const { t } = useTranslation()
    const [activeTab, setActiveTab] = useState('cli')

    return (
        <section id="features" className="py-32 bg-background relative">
            <div className="max-w-7xl mx-auto px-6">
                <div className="text-center mb-16">
                    <h2 className="text-3xl md:text-4xl font-bold text-white mb-4">
                        {t('features.title')} <span className="text-primary">{t('features.titleHighlight')}</span>
                    </h2>
                    <p className="text-slate-400 text-lg max-w-2xl mx-auto">
                        {t('features.subtitle')}
                    </p>
                </div>

                {/* Toggle */}
                <div className="flex justify-center mb-16">
                    <div className="bg-slate-900 p-1.5 rounded-full border border-slate-800 inline-flex shadow-inner">
                        <button
                            onClick={() => setActiveTab('cli')}
                            className={`px-8 py-2.5 rounded-full text-sm font-medium transition-all duration-300 ${activeTab === 'cli' ? 'bg-white text-black shadow-lg' : 'text-slate-400 hover:text-white'}`}
                        >
                            {t('features.localDev')}
                        </button>
                        <button
                            onClick={() => setActiveTab('server')}
                            className={`px-8 py-2.5 rounded-full text-sm font-medium transition-all duration-300 ${activeTab === 'server' ? 'bg-white text-black shadow-lg' : 'text-slate-400 hover:text-white'}`}
                        >
                            {t('features.remoteInfra')}
                        </button>
                    </div>
                </div>

                {/* Content Area */}
                <div className="grid lg:grid-cols-2 gap-12 items-center">

                    {/* Left Text */}
                    <div className="space-y-8">
                        {activeTab === 'cli' ? (
                            <div className="animate-fade-in-up">
                                <h3 className="text-2xl font-semibold text-white mb-4">{t('features.cli.title')}</h3>
                                <p className="text-slate-400 text-lg leading-relaxed mb-8">
                                    {t('features.cli.description')}
                                </p>
                                <ul className="space-y-4">
                                    <li className="flex gap-4">
                                        <div className="w-6 h-6 rounded-full bg-primary/10 flex items-center justify-center shrink-0 text-primary mt-1">
                                            <Package size={14} />
                                        </div>
                                        <div>
                                            <h4 className="text-white font-medium">{t('features.cli.packageManager.title')}</h4>
                                            <p className="text-sm text-slate-500">{t('features.cli.packageManager.desc')}</p>
                                        </div>
                                    </li>
                                    <li className="flex gap-4">
                                        <div className="w-6 h-6 rounded-full bg-primary/10 flex items-center justify-center shrink-0 text-primary mt-1">
                                            <CheckCircle size={14} />
                                        </div>
                                        <div>
                                            <h4 className="text-white font-medium">{t('features.cli.linting.title')}</h4>
                                            <p className="text-sm text-slate-500">{t('features.cli.linting.desc')}</p>
                                        </div>
                                    </li>
                                    <li className="flex gap-4">
                                        <div className="w-6 h-6 rounded-full bg-primary/10 flex items-center justify-center shrink-0 text-primary mt-1">
                                            <CheckCircle size={14} />
                                        </div>
                                        <div>
                                            <h4 className="text-white font-medium">{t('features.cli.apiSafety.title')}</h4>
                                            <p className="text-sm text-slate-500">{t('features.cli.apiSafety.desc')}</p>
                                        </div>
                                    </li>
                                </ul>
                            </div>
                        ) : (
                            <div className="animate-fade-in-up">
                                <h3 className="text-2xl font-semibold text-white mb-4">{t('features.server.title')}</h3>
                                <p className="text-slate-400 text-lg leading-relaxed mb-8">
                                    {t('features.server.description')}
                                </p>
                                <ul className="space-y-4">
                                    <li className="flex gap-4">
                                        <div className="w-6 h-6 rounded-full bg-secondary/10 flex items-center justify-center shrink-0 text-secondary mt-1">
                                            <Shield size={14} />
                                        </div>
                                        <div>
                                            <h4 className="text-white font-medium">{t('features.server.dockerNative.title')}</h4>
                                            <p className="text-sm text-slate-500">{t('features.server.dockerNative.desc')}</p>
                                        </div>
                                    </li>
                                    <li className="flex gap-4">
                                        <div className="w-6 h-6 rounded-full bg-secondary/10 flex items-center justify-center shrink-0 text-secondary mt-1">
                                            <Shield size={14} />
                                        </div>
                                        <div>
                                            <h4 className="text-white font-medium">{t('features.server.airGap.title')}</h4>
                                            <p className="text-sm text-slate-500">{t('features.server.airGap.desc')}</p>
                                        </div>
                                    </li>
                                    <li className="flex gap-4">
                                        <div className="w-6 h-6 rounded-full bg-secondary/10 flex items-center justify-center shrink-0 text-secondary mt-1">
                                            <Shield size={14} />
                                        </div>
                                        <div>
                                            <h4 className="text-white font-medium">{t('features.server.costEffective.title')}</h4>
                                            <p className="text-sm text-slate-500">{t('features.server.costEffective.desc')}</p>
                                        </div>
                                    </li>
                                </ul>
                            </div>
                        )}
                    </div>

                    {/* Right Visual */}
                    <div className="relative group">
                        <div className="absolute -inset-1 bg-gradient-to-r from-primary to-secondary rounded-2xl blur opacity-20 group-hover:opacity-40 transition duration-1000 group-hover:duration-200"></div>
                        <div className="relative rounded-xl bg-slate-900 border border-slate-800 overflow-hidden shadow-2xl h-[400px] flex flex-col">
                            <div className="flex items-center px-4 py-3 border-b border-slate-800 bg-slate-900/50 gap-2">
                                <div className="w-3 h-3 rounded-full bg-slate-700"></div>
                                <div className="w-3 h-3 rounded-full bg-slate-700"></div>
                            </div>
                            <div className="p-6 font-mono text-sm flex-grow overflow-hidden">
                                {activeTab === 'cli' ? (
                                    <div className="space-y-2">
                                        <div className="text-slate-400"># easyp.yaml configuration</div>
                                        <div className="text-purple-400">version: <span className="text-green-400">v1</span></div>
                                        <div className="text-purple-400">deps:</div>
                                        <div className="pl-4 text-slate-300">- github.com/googleapis/googleapis@v1.0</div>
                                        <div className="text-purple-400 mt-2">lint:</div>
                                        <div className="pl-4 text-blue-400">use:</div>
                                        <div className="pl-8 text-slate-300">- STANDARD</div>
                                        <div className="pl-8 text-slate-300">- COMMENTS</div>
                                        <div className="text-purple-400 mt-4">breaking:</div>
                                        <div className="pl-4 text-blue-400">use:</div>
                                        <div className="pl-8 text-slate-300">- WIRE_JSON</div>
                                    </div>
                                ) : (
                                    <div className="space-y-2">
                                        <div className="text-slate-400"># Server Infrastructure</div>
                                        <div className="flex items-center gap-2 py-2 border-b border-slate-800/50">
                                            <Box size={16} className="text-blue-400" />
                                            <span className="text-slate-200">EasyP Service</span>
                                            <span className="ml-auto text-xs bg-green-500/10 text-green-500 px-2 rounded-full">Healthy</span>
                                        </div>
                                        <div className="pl-4 border-l border-slate-800 space-y-3 mt-3">
                                            <div className="flex items-center gap-2">
                                                <div className="w-2 h-2 bg-slate-600 rounded-full"></div>
                                                <span className="text-slate-400">Registry:</span>
                                                <span className="text-slate-200">Local Docker</span>
                                            </div>
                                            <div className="flex items-center gap-2">
                                                <div className="w-2 h-2 bg-slate-600 rounded-full"></div>
                                                <span className="text-slate-400">Plugins:</span>
                                                <span className="text-slate-200">14 Loaded</span>
                                            </div>
                                            <div className="flex items-center gap-2">
                                                <div className="w-2 h-2 bg-slate-600 rounded-full"></div>
                                                <span className="text-slate-400">Uptime:</span>
                                                <span className="text-slate-200">42d 12h</span>
                                            </div>
                                        </div>
                                    </div>
                                )}
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </section>
    )
}
