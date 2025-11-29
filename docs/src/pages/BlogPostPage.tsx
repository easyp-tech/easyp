import { useParams, Link } from 'react-router-dom'
import { useState, useEffect } from 'react'
import { MarkdownContent } from '../lib/markdown'
import { Loader2, AlertCircle, ArrowLeft } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import Navbar from '../components/Navbar'

export default function BlogPostPage() {
    const { slug } = useParams<{ slug: string }>()
    const { i18n } = useTranslation()
    const [content, setContent] = useState<string>('')
    const [isLoading, setIsLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        const loadMarkdown = async () => {
            setIsLoading(true)
            setError(null)

            try {
                // Helper to fetch and validate markdown
                const fetchMarkdown = async (url: string): Promise<string | null> => {
                    const response = await fetch(url)

                    if (!response.ok) return null

                    const contentType = response.headers.get('content-type')
                    if (contentType && contentType.includes('text/html')) {
                        return null // It's the SPA fallback, not the file
                    }

                    return response.text()
                }

                // Build candidate paths based on language
                const langPrefix = i18n.language === 'ru' ? 'ru-blog' : 'blog'
                const candidates = [
                    `/docs/${langPrefix}/${slug}.md`,
                    `/docs/blog/${slug}.md` // Fallback to English
                ]

                let text: string | null = null
                for (const url of candidates) {
                    text = await fetchMarkdown(url)
                    if (text) {
                        break
                    }
                }

                if (!text) {
                    throw new Error('Blog post not found')
                }

                setContent(text)
            } catch (err) {
                console.error('Error loading blog post:', err)
                setError(err instanceof Error ? err.message : 'An unexpected error occurred')
            } finally {
                setIsLoading(false)
            }
        }

        if (slug) {
            loadMarkdown()
        }
    }, [slug, i18n.language])

    if (isLoading) {
        return (
            <div className="min-h-screen bg-gradient-to-b from-slate-950 via-slate-900 to-slate-950">
                <Navbar />
                <div className="flex items-center justify-center min-h-[60vh] pt-20">
                    <div className="text-center">
                        <Loader2 className="w-12 h-12 text-blue-500 animate-spin mx-auto mb-4" />
                        <p className="text-slate-400">Loading blog post...</p>
                    </div>
                </div>
            </div>
        )
    }

    if (error) {
        return (
            <div className="min-h-screen bg-gradient-to-b from-slate-950 via-slate-900 to-slate-950">
                <Navbar />
                <div className="flex items-center justify-center min-h-[60vh] pt-20">
                    <div className="text-center max-w-md">
                        <AlertCircle className="w-12 h-12 text-red-500 mx-auto mb-4" />
                        <h2 className="text-2xl font-bold text-white mb-2">
                            Failed to Load Post
                        </h2>
                        <p className="text-slate-400 mb-4">{error}</p>
                        <Link
                            to="/blog"
                            className="inline-flex items-center text-blue-400 hover:text-blue-300 transition-colors"
                        >
                            <ArrowLeft className="w-4 h-4 mr-2" />
                            Back to Blog
                        </Link>
                    </div>
                </div>
            </div>
        )
    }

    return (
        <div className="min-h-screen bg-gradient-to-b from-slate-950 via-slate-900 to-slate-950">
            <Navbar />

            <div className="max-w-4xl mx-auto px-6 pt-32 pb-20">
                {/* Back to blog */}
                <Link
                    to="/blog"
                    className="inline-flex items-center text-slate-400 hover:text-white transition-colors mb-8"
                >
                    <ArrowLeft className="w-4 h-4 mr-2" />
                    {i18n.language === 'ru' ? 'Назад к блогу' : 'Back to Blog'}
                </Link>

                {/* Article content */}
                <article className="prose prose-invert prose-lg max-w-none">
                    <MarkdownContent
                        content={content}
                        options={{
                            enableToc: false,
                            enableCodeFocus: true,
                            enableHtmlBlocks: true,
                            enableLinkProcessing: true
                        }}
                    />
                </article>
            </div>
        </div>
    )
}
