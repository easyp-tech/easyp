import { Outlet, useNavigate, Link } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { Menu, X } from 'lucide-react'
import DocsSidebar from './DocsSidebar'
import TableOfContents from '../../lib/markdown/components/TableOfContents'
import { useToc } from '../../contexts/TocContext'

export default function DocsLayout() {
    const navigate = useNavigate()
    const [sidebarOpen, setSidebarOpen] = useState(false)
    const tocContext = useToc()

    // Set up window.routerNavigate for markdown link integration
    useEffect(() => {
        (window as any).routerNavigate = (path: string) => {
            navigate(path)
            // Закрываем sidebar на мобильных после навигации
            setSidebarOpen(false)
        }

        return () => {
            delete (window as any).routerNavigate
        }
    }, [navigate])

    return (
        <div className="min-h-screen bg-white dark:bg-gray-950">
            {/* Documentation Layout */}
            <div className="docs-layout">
                {/* Sidebar */}
                <DocsSidebar
                    isOpen={sidebarOpen}
                    onClose={() => setSidebarOpen(false)}
                />

                {/* Main Content Area */}
                <main className="docs-main">
                    {/* Mobile Header with Hamburger */}
                    <div className="docs-mobile-header md:hidden">
                        <button
                            className="docs-sidebar-toggle"
                            onClick={() => setSidebarOpen(!sidebarOpen)}
                            aria-label="Toggle sidebar"
                        >
                            {sidebarOpen ? (
                                <X className="w-6 h-6" />
                            ) : (
                                <Menu className="w-6 h-6" />
                            )}
                        </button>
                        <Link to="/" className="flex items-center gap-2 hover:opacity-80 transition-opacity">
                            <div className="w-6 h-6 bg-blue-500/20 rounded flex items-center justify-center border border-blue-500/30">
                                <span className="font-bold text-blue-500 text-xs">EP</span>
                            </div>
                            <h1 className="text-lg font-bold text-gray-900 dark:text-white">
                                EasyP
                            </h1>
                        </Link>
                    </div>

                    <div className="docs-content">
                        <Outlet />
                    </div>
                </main>

                {/* Table of Contents - Right Column */}
                <aside className="docs-toc-container">
                    {tocContext?.toc && tocContext.toc.length > 0 && (
                        <TableOfContents
                            toc={tocContext.toc}
                            activeId={tocContext.activeId}
                        />
                    )}
                </aside>
            </div>
        </div>
    )
}
