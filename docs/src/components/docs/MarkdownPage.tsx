import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { MarkdownContent } from '../../lib/markdown'
import { Loader2, AlertCircle } from 'lucide-react'
import { useTranslation } from 'react-i18next'

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
    const { i18n } = useTranslation()
    const [content, setContent] = useState<string>('')
    const [isLoading, setIsLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

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
                // filePath at this point is like "guide/introduction/what-is" or "introduction/what-is"
                const normalizeBase = (fp: string): string => {
                    if (fp.startsWith('guide/')) return fp.slice('guide/'.length)
                    if (fp.startsWith('ru-guide/')) return fp.slice('ru-guide/'.length)
                    return fp
                }

                const baseRelative = normalizeBase(filePath)
                const langRoot = i18n.language === 'ru' ? 'ru-guide' : 'guide'

                // Build candidate paths in priority order:
                // 1. Language-specific path (ru-guide/...)
                // 2. English fallback (guide/...)
                const candidates = [
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
            } catch (err) {
                console.error('Error loading markdown:', err)
                setError(err instanceof Error ? err.message : 'An unexpected error occurred')
            } finally {
                setIsLoading(false)
            }
        }

        loadMarkdown()
    }, [path, params.category, params.subcategory, params.subsubcategory, params.section, params.page])

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
        <div className="markdown-page">
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
        </div>
    )
}
