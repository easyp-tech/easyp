import { useState, useEffect } from 'react'
import { useParams, useLocation, Link } from 'react-router-dom'
import { MarkdownContent } from '../../lib/markdown'
import { Loader2, AlertCircle, ChevronLeft, ChevronRight, Github } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import DocsHeader from './DocsHeader'
import { getPrevNext, loadSidebarConfig } from '../../utils/sidebarUtils'
import type { SidebarItem } from '../../types/sidebar'

interface MarkdownPageProps {
    path?: string
}

export default function MarkdownPage({ path }: MarkdownPageProps) {
    const params = useParams<{
        category?: string;
        subcategory?: string;
        subsubcategory?: string;
        section?: string;
        page?: string
    }>()
    const location = useLocation()
    const { i18n } = useTranslation()
    const [content, setContent] = useState<string>('')
    const [isLoading, setIsLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [activeFilePath, setActiveFilePath] = useState<string>('')
    const [prevNext, setPrevNext] = useState<{ prev: SidebarItem | null; next: SidebarItem | null }>({
        prev: null,
        next: null
    })

    useEffect(() => {
        const loadMarkdown = async () => {
            setIsLoading(true)
            setError(null)

            try {
                // Construct the file path from either prop or route params
                let filePath: string

                if (path) {
                    filePath = path
                } else if (params.section) {
                    // Five-level route: /docs/:category/:subcategory/:subsubcategory/:section/:page
                    filePath = `${params.category}/${params.subcategory}/${params.subsubcategory}/${params.section}/${params.page}`
                } else if (params.subsubcategory) {
                    // Four-level route: /docs/:category/:subcategory/:subsubcategory/:page
                    filePath = `${params.category}/${params.subcategory}/${params.subsubcategory}/${params.page}`
                } else if (params.subcategory) {
                    // Three-level route: /docs/:category/:subcategory/:page
                    filePath = `${params.category}/${params.subcategory}/${params.page}`
                } else if (params.category && params.page) {
                    // Two-level route: /docs/:category/:page
                    filePath = `${params.category}/${params.page}`
                } else {
                    throw new Error('Invalid route parameters')
                }

                // Helper to fetch and validate markdown
                const fetchMarkdown = async (url: string): Promise<string | null> => {
                    const response = await fetch(url);
                    if (!response.ok) return null;

                    const contentType = response.headers.get('content-type');
                    if (contentType && contentType.includes('text/html')) {
                        return null; // It's the SPA fallback, not the file
                    }

                    return response.text();
                };

                // Determine language-specific root directory and normalize relative path
                // filePath at this point is like "guide/introduction/what-is" or "introduction/what-is" or "blog/finally-give-up-gin-echo"
                const normalizeBase = (fp: string): string => {
                    if (fp.startsWith('guide/')) return fp.slice('guide/'.length)
                    if (fp.startsWith('ru-guide/')) return fp.slice('ru-guide/'.length)
                    if (fp.startsWith('blog/')) return fp.slice('blog/'.length)
                    return fp
                }

                const baseRelative = normalizeBase(filePath)

                // Check if this is a blog path
                const isBlogPath = filePath.startsWith('blog/')

                // Build candidate paths in priority order:
                let candidates: string[] = []

                if (isBlogPath) {
                    // For blog posts, only check the blog directory
                    candidates = [`/docs/blog/${baseRelative}.md`]
                } else {
                    // For regular docs:
                    // 1. Language-specific path (ru-guide/...)
                    // 2. English fallback (guide/...)
                    const langRoot = i18n.language === 'ru' ? 'ru-guide' : 'guide'
                    candidates = [
                        `/docs/${langRoot}/${baseRelative}.md`,
                        `/docs/guide/${baseRelative}.md`
                    ]

                    // Support nested structure variant: /path/to/page/page.md
                    if (params.page) {
                        candidates.push(
                            `/docs/${langRoot}/${baseRelative}/${params.page}.md`,
                            `/docs/guide/${baseRelative}/${params.page}.md`
                        )
                    }
                }

                let text: string | null = null
                for (const url of candidates) {
                    text = await fetchMarkdown(url)
                    if (text) {
                        break
                    }
                }

                if (!text) {
                    throw new Error('Documentation page not found (checked language + fallback).')
                }

                setContent(text)
                // Store the path that was actually loaded (relative to docs root)
                // For the edit link, we need to know the correct path
                const effectivePath = isBlogPath ? `blog/${baseRelative}.md` : `${i18n.language === 'ru' ? 'ru-guide' : 'guide'}/${baseRelative}.md`
                setActiveFilePath(effectivePath)
            } catch (err) {
                console.error('Error loading markdown:', err)
                setError(err instanceof Error ? err.message : 'An unexpected error occurred')
            } finally {
                setIsLoading(false)
            }
        }

        loadMarkdown()
    }, [path, params.category, params.subcategory, params.subsubcategory, params.section, params.page, i18n.language])

    useEffect(() => {
        const config = loadSidebarConfig()
        const { prev, next } = getPrevNext(config, location.pathname)
        setPrevNext({ prev, next })
    }, [location.pathname])

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-[60vh]">
                <div className="text-center">
                    <Loader2 className="w-12 h-12 text-blue-600 animate-spin mx-auto mb-4" />
                    <p className="text-gray-600 dark:text-gray-400">Loading documentation...</p>
                </div>
            </div>
        )
    }

    if (error) {
        return (
            <div className="flex items-center justify-center min-h-[60vh]">
                <div className="text-center max-w-md">
                    <AlertCircle className="w-12 h-12 text-red-500 mx-auto mb-4" />
                    <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
                        Failed to Load Page
                    </h2>
                    <p className="text-gray-600 dark:text-gray-400 mb-4">
                        {error}
                    </p>
                    <p className="text-sm text-gray-500">
                        Please check the URL or try again later.
                    </p>
                </div>
            </div>
        )
    }

    return (
        <div className="flex flex-col min-h-screen bg-white dark:bg-gray-950">
            <DocsHeader />

            <div className="flex-1 w-full max-w-4xl mx-auto px-4 md:px-8 py-8">
                <MarkdownContent
                    content={content}
                    options={{
                        enableToc: true,
                        enableCodeFocus: true,
                        enableHtmlBlocks: true,
                        enableLinkProcessing: true,
                        tocDepth: { min: 2, max: 4 }
                    }}
                />

                <div className="mt-16 pt-8 border-t border-gray-200 dark:border-gray-800">
                    <div className="flex justify-end mb-8">
                        <a
                            href={`https://github.com/easyp-tech/easyp/edit/main/docs/public/docs/${activeFilePath}`}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="flex items-center text-sm text-gray-500 hover:text-blue-600 dark:text-gray-400 dark:hover:text-blue-400 transition-colors"
                        >
                            <Github className="w-4 h-4 mr-2" />
                            Edit this page on GitHub
                        </a>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        {prevNext.prev ? (
                            <Link
                                to={prevNext.prev.path!}
                                className="flex flex-col p-4 rounded-lg border border-gray-200 dark:border-gray-800 hover:border-blue-500 dark:hover:border-blue-500 transition-colors group"
                            >
                                <span className="text-xs text-gray-500 dark:text-gray-400 mb-1 flex items-center">
                                    <ChevronLeft className="w-3 h-3 mr-1" />
                                    Previous
                                </span>
                                <span className="font-medium text-blue-600 dark:text-blue-400 group-hover:underline truncate">
                                    {prevNext.prev.title}
                                </span>
                            </Link>
                        ) : <div />}

                        {prevNext.next && (
                            <Link
                                to={prevNext.next.path!}
                                className="flex flex-col items-end p-4 rounded-lg border border-gray-200 dark:border-gray-800 hover:border-blue-500 dark:hover:border-blue-500 transition-colors group"
                            >
                                <span className="text-xs text-gray-500 dark:text-gray-400 mb-1 flex items-center">
                                    Next
                                    <ChevronRight className="w-3 h-3 ml-1" />
                                </span>
                                <span className="font-medium text-blue-600 dark:text-blue-400 group-hover:underline truncate">
                                    {prevNext.next.title}
                                </span>
                            </Link>
                        )}
                    </div>
                </div>
            </div>
        </div>
    )
}
