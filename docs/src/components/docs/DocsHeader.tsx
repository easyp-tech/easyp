import { Link, useLocation } from 'react-router-dom'
import { ChevronRight, Home } from 'lucide-react'

export default function DocsHeader() {
    const location = useLocation()
    const pathSegments = location.pathname
        .split('/')
        .filter(segment => segment && segment !== 'docs')

    return (
        <header className="sticky top-0 z-30 w-full border-b border-gray-200 dark:border-gray-800 bg-white/80 dark:bg-gray-950/80 backdrop-blur supports-[backdrop-filter]:bg-white/60">
            <div className="flex h-14 items-center gap-4 px-4 md:px-8">
                {/* Breadcrumbs */}
                <nav className="hidden md:flex items-center text-sm text-gray-500 dark:text-gray-400 overflow-hidden whitespace-nowrap">
                    <Link to="/docs" className="hover:text-gray-900 dark:hover:text-gray-100 transition-colors flex-shrink-0">
                        <Home className="w-4 h-4" />
                    </Link>
                    {pathSegments.map((segment, index) => {
                        const path = `/docs/${pathSegments.slice(0, index + 1).join('/')}`
                        const isLast = index === pathSegments.length - 1
                        const title = segment.charAt(0).toUpperCase() + segment.slice(1).replace(/-/g, ' ')

                        return (
                            <div key={path} className="flex items-center flex-shrink-0">
                                <ChevronRight className="w-4 h-4 mx-1 flex-shrink-0" />
                                {isLast ? (
                                    <span className="font-medium text-gray-900 dark:text-gray-100 truncate">
                                        {title}
                                    </span>
                                ) : (
                                    <Link
                                        to={path}
                                        className="hover:text-gray-900 dark:hover:text-gray-100 transition-colors truncate"
                                    >
                                        {title}
                                    </Link>
                                )}
                            </div>
                        )
                    })}
                </nav>

                <div className="flex-1" />
            </div>
        </header>
    )
}
