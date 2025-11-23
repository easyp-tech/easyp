import { useState, useEffect, useRef } from 'react'
import { Shield, ChevronRight, ChevronDown, Copy, CheckCircle } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Link } from 'react-router-dom'
import { GetLatestRelease } from '../utils/github'

interface InstallOption {
    id: string
    label: string
    fullLabel: string
    cmd: string
}

export default function Hero() {
    const { t } = useTranslation()
    const [installMethod, setInstallMethod] = useState('brew')
    const [isDropdownOpen, setIsDropdownOpen] = useState(false)
    const [version, setVersion] = useState('v0.8.1')
    const dropdownRef = useRef<HTMLDivElement>(null)

    const installOptions: InstallOption[] = [
        { id: 'brew', label: 'brew', fullLabel: 'macOS (homebrew)', cmd: 'brew install easyp-tech/tap/easyp' },
        { id: 'go', label: 'go', fullLabel: 'go Install (any OS)', cmd: 'go install github.com/easyp-tech/easyp/cmd/easyp@latest' },
        { id: 'docker', label: 'docker', fullLabel: 'docker', cmd: 'docker pull easyp/easyp:latest' }
    ]

    const activeOption = installOptions.find(o => o.id === installMethod) || installOptions[0]

    // Auto-detect OS on mount
    useEffect(() => {
        const platform = navigator.platform.toLowerCase()
        const userAgent = navigator.userAgent.toLowerCase()

        if (platform.includes('mac') || userAgent.includes('mac')) {
            setInstallMethod('brew')
        } else if (platform.includes('linux') || userAgent.includes('linux') || platform.includes('win') || userAgent.includes('win')) {
            setInstallMethod('go')
        } else {
            setInstallMethod('go')
        }

        const handleClickOutside = (event: MouseEvent) => {
            if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
                setIsDropdownOpen(false)
            }
        }
        document.addEventListener('mousedown', handleClickOutside)
        return () => document.removeEventListener('mousedown', handleClickOutside)
    }, [])

    useEffect(() => {
        GetLatestRelease().then(v => {
            if (v !== 'unknown version') {
                setVersion(v)
            }
        })
    }, [])

    const handleCopy = () => {
        navigator.clipboard.writeText(activeOption.cmd)
    }

    return (
        <section className="relative pt-32 pb-20 lg:pt-48 lg:pb-32 overflow-hidden">
            {/* Ambient Background */}
            <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[1000px] h-[600px] bg-primary/20 rounded-full blur-[120px] opacity-30 pointer-events-none"></div>

            <div className="max-w-7xl mx-auto px-6 relative z-10">
                <div className="flex flex-col items-center text-center mb-16">

                    <div className="flex gap-3 mb-8 animate-fade-in-up">
                        <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-slate-800/50 border border-slate-700/50 text-primary text-xs font-medium">
                            <span className="relative flex h-2 w-2">
                                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-primary opacity-75"></span>
                                <span className="relative inline-flex rounded-full h-2 w-2 bg-primary"></span>
                            </span>
                            {version}
                        </div>
                        <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-slate-800/50 border border-slate-700/50 text-slate-300 text-xs font-medium">
                            <Shield size={12} className="text-emerald-400" />
                            {t('hero.badge.license')}
                        </div>
                    </div>

                    <h1 className="text-5xl lg:text-7xl font-bold text-white tracking-tight leading-[1.1] mb-6 max-w-4xl animate-fade-in-up [animation-delay:100ms]">
                        {t('hero.title')} <br />
                        <span className="text-gradient">{t('hero.titleHighlight')}</span>
                    </h1>

                    <p className="text-lg text-slate-400 max-w-2xl mb-10 leading-relaxed animate-fade-in-up [animation-delay:200ms]">
                        {t('hero.subtitle')}
                    </p>

                    <div className="flex flex-col sm:flex-row items-center gap-4 animate-fade-in-up [animation-delay:300ms] w-full max-w-2xl relative z-20">

                        {/* Install Box with Integrated Dropdown */}
                        <div className="relative w-full sm:w-auto flex-grow group" ref={dropdownRef}>
                            <div className="flex items-center bg-slate-900 border border-slate-800 rounded-xl px-1 py-1 pr-2 shadow-xl w-full group-hover:border-slate-600 transition-colors">

                                {/* Trigger Button */}
                                <button
                                    onClick={() => setIsDropdownOpen(!isDropdownOpen)}
                                    className="flex items-center gap-2 text-slate-300 hover:text-white hover:bg-slate-800 transition-all text-sm font-medium px-3 py-2 rounded-lg shrink-0 border-r border-slate-800/50 mr-2"
                                    title="Select installation method"
                                >
                                    <span>{activeOption.label}</span>
                                    <ChevronDown size={14} className={`text-slate-500 transition-transform duration-200 ${isDropdownOpen ? 'rotate-180' : ''}`} />
                                </button>

                                {/* Command Code */}
                                <div className="flex-grow overflow-hidden flex items-center">
                                    <ChevronRight size={14} className="text-slate-600 mr-2 shrink-0" />
                                    <code className="font-mono text-sm text-slate-300 whitespace-nowrap overflow-x-auto no-scrollbar block selection:bg-primary/30 selection:text-white">
                                        {activeOption.cmd}
                                    </code>
                                </div>

                                {/* Copy Button */}
                                <button
                                    className="p-2 hover:bg-slate-800 rounded-lg transition-colors text-slate-500 hover:text-white shrink-0 ml-2"
                                    title="Copy to clipboard"
                                    onClick={handleCopy}
                                >
                                    <Copy size={16} />
                                </button>
                            </div>

                            {/* Dropdown Menu */}
                            {isDropdownOpen && (
                                <div className="absolute top-full left-0 mt-2 w-64 bg-slate-900 border border-slate-800 rounded-xl shadow-2xl overflow-hidden z-30">
                                    <div className="py-1">
                                        {installOptions.map(opt => (
                                            <button
                                                key={opt.id}
                                                onClick={() => {
                                                    setInstallMethod(opt.id)
                                                    setIsDropdownOpen(false)
                                                }}
                                                className={`w-full text-left px-4 py-3 text-sm transition-colors flex items-center justify-between group ${installMethod === opt.id ? 'bg-primary/10' : 'hover:bg-slate-800'}`}
                                            >
                                                <span className={`${installMethod === opt.id ? 'text-primary font-medium' : 'text-slate-300 group-hover:text-white'}`}>
                                                    {opt.fullLabel}
                                                </span>
                                                {installMethod === opt.id && <CheckCircle size={16} className="text-primary" />}
                                            </button>
                                        ))}
                                    </div>
                                </div>
                            )}
                        </div>

                        {/* Explore Docs Button */}
                        <Link to="/docs" className="w-full sm:w-auto px-6 py-3 rounded-full bg-primary/10 text-primary font-medium hover:bg-primary/20 transition-colors border border-primary/20 whitespace-nowrap text-center inline-block">
                            {t('hero.exploreDocs')}
                        </Link>
                    </div>
                </div>

                {/* Hero Image / Visual */}
                <div className="relative max-w-5xl mx-auto rounded-xl border border-slate-800 bg-slate-950/80 shadow-2xl shadow-primary/10 overflow-hidden animate-fade-in-up [animation-delay:500ms]">
                    <div className="flex items-center px-4 py-3 border-b border-slate-800 bg-slate-900/50 gap-2">
                        <div className="w-3 h-3 rounded-full bg-slate-700"></div>
                        <div className="w-3 h-3 rounded-full bg-slate-700"></div>
                        <div className="w-3 h-3 rounded-full bg-slate-700"></div>
                        <div className="ml-4 text-xs text-slate-500 font-mono">easyp — zsh</div>
                    </div>
                    <div className="p-6 sm:p-10 font-mono text-sm sm:text-base leading-loose">
                        <div className="flex items-start gap-2">
                            <span className="text-primary font-bold">➜</span>
                            <span className="text-slate-300">easyp generate -v</span>
                        </div>
                        <div className="text-slate-400 mt-2 pl-4 border-l-2 border-slate-800">
                            <div className="flex items-center gap-2">
                                <span className="text-emerald-400">✔</span>
                                <span>Resolving dependencies from <span className="text-slate-300 underline decoration-slate-700">easyp.lock</span>...</span>
                            </div>
                            <div className="flex items-center gap-2 mt-1">
                                <span className="text-emerald-400">✔</span>
                                <span>Connecting to EasyP Service... <span className="text-xs bg-emerald-500/10 text-emerald-400 px-1.5 py-0.5 rounded">Connected</span></span>
                            </div>
                            <div className="mt-2">
                                <span className="text-blue-400">[INFO]</span> Dispatching plugins:
                                <div className="pl-4 mt-1 text-slate-500">
                                    ├─ protoc-gen-go (v1.31) <span className="text-xs border border-slate-700 px-1 rounded text-slate-400">WASM</span><br />
                                    └─ custom-java-gen (v2.0) <span className="text-xs border border-slate-700 px-1 rounded text-slate-400">DOCKER</span>
                                </div>
                            </div>
                            <div className="mt-2 text-slate-300">
                                ✨ Generation complete in <span className="text-primary">420ms</span>.
                            </div>
                        </div>
                        <div className="flex items-start gap-2 mt-4 animate-pulse">
                            <span className="text-primary font-bold">➜</span>
                            <span className="w-2 h-5 bg-slate-500 inline-block align-middle"></span>
                        </div>
                    </div>
                </div>
            </div>
        </section>
    )
}
