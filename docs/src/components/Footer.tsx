import { Shield } from 'lucide-react'
import { useTranslation } from 'react-i18next'

export default function Footer() {
    const { t } = useTranslation()
    return (
        <footer className="bg-slate-950 border-t border-slate-800 py-16">
            <div className="max-w-7xl mx-auto px-6">
                <div className="grid md:grid-cols-4 gap-12">
                    <div className="col-span-2">
                        <div className="flex items-center gap-2 mb-6">
                            <div className="w-6 h-6 bg-primary/20 rounded flex items-center justify-center text-primary text-xs font-bold">E</div>
                            <span className="font-bold text-white">EasyP</span>
                        </div>
                        <p className="text-slate-400 text-sm max-w-xs leading-relaxed mb-4">
                            {t('footer.description')}
                        </p>
                        <div className="flex items-center gap-2 text-xs text-slate-500">
                            <Shield size={12} /> {t('footer.license')}
                        </div>
                    </div>
                    <div>
                        <h4 className="text-white font-medium mb-4">{t('footer.product.title')}</h4>
                        <ul className="space-y-3 text-sm text-slate-400">
                            <li><a href="/docs/guide/cli/linter/linter" className="hover:text-primary">{t('footer.product.cliTool')}</a></li>
                            <li><a href="/docs/guide/api-service/overview" className="hover:text-primary">{t('footer.product.apiService')}</a></li>
                            {/* <li><a href="#" className="hover:text-primary">{t('footer.product.registry')}</a></li> */}
                            <li><a href="/docs" className="hover:text-primary">{t('footer.product.documentation')}</a></li>
                            {/* <li><a href="#" className="hover:text-primary">{t('footer.product.enterprise')}</a></li> */}
                        </ul>
                    </div>
                    <div>
                        <h4 className="text-white font-medium mb-4">{t('footer.community.title')}</h4>
                        <ul className="space-y-3 text-sm text-slate-400">
                            <li><a href="https://github.com/easyp-tech/easyp" className="hover:text-primary">{t('footer.community.github')}</a></li>
                            <li><a href="https://t.me/easyptech" className="hover:text-primary">{t('footer.community.telegram')}</a></li>
                            {/* <li><a href="#" className="hover:text-primary">{t('footer.community.youtube')}</a></li> */}
                        </ul>
                    </div>
                </div>
                <div className="border-t border-slate-800 mt-16 pt-8 flex flex-col md:flex-row justify-between items-center text-xs text-slate-600">
                    <p>{t('footer.rights')}</p>
                    {/* <div className="flex gap-6 mt-4 md:mt-0">
                        <a href="#" className="hover:text-slate-400">{t('footer.privacy')}</a>
                        <a href="#" className="hover:text-slate-400">{t('footer.terms')}</a>
                    </div> */}
                </div>
            </div>
        </footer>
    )
}
