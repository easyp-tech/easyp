import { ArrowRight } from 'lucide-react'

export default function RegistryPreview() {
    const plugins = [
        { name: 'Go', version: 'v1.36.0', type: 'Official' },
        { name: 'Python', version: 'v4.24.0', type: 'Docker' },
        { name: 'Java', version: 'v3.21.0', type: 'Community' },
        { name: 'TypeScript', version: 'v1.12.0', type: 'Wasm' },
    ]

    return (
        <section id="registry" className="py-24 relative overflow-hidden">
            <div className="max-w-7xl mx-auto px-6">
                <div className="flex items-center justify-between mb-12">
                    <h2 className="text-2xl font-bold text-white">Plugin Registry</h2>
                    <a href="#" className="text-sm text-primary hover:text-primaryGlow flex items-center gap-1">
                        View all <ArrowRight size={14} />
                    </a>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
                    {plugins.map((plugin, i) => (
                        <div key={i} className="bg-slate-900 border border-slate-800 p-6 rounded-xl hover:border-slate-600 transition-all cursor-pointer group">
                            <div className="flex justify-between items-start mb-4">
                                <div className="w-10 h-10 bg-slate-800 rounded flex items-center justify-center text-lg font-bold text-white group-hover:bg-white group-hover:text-black transition-colors">
                                    {plugin.name[0]}
                                </div>
                                <span className="text-xs font-mono text-slate-500 bg-slate-950 px-2 py-1 rounded border border-slate-800">
                                    {plugin.type}
                                </span>
                            </div>
                            <h4 className="text-white font-medium mb-1">{plugin.name}</h4>
                            <p className="text-xs text-slate-500 font-mono">{plugin.version}</p>
                        </div>
                    ))}
                </div>
            </div>
        </section>
    )
}
