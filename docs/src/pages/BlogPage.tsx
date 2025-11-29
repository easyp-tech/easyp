import { Link } from 'react-router-dom'
import Navbar from '../components/Navbar'
import { Calendar, ArrowRight } from 'lucide-react'
import { useTranslation } from 'react-i18next'

interface BlogPost {
    title: string
    titleRu: string
    slug: string
    date: string
    excerpt: string
    excerptRu: string
}

const blogPosts: BlogPost[] = [
    {
        title: "Let's Work with Proto Errors Correctly :)",
        titleRu: "Давайте работать с proto ошибками правильно :)",
        slug: "working-with-proto-errors-correctly",
        date: "2025-11-30",
        excerpt: "Best practices for handling errors in gRPC and Protocol Buffers",
        excerptRu: "Лучшие практики обработки ошибок в gRPC и Protocol Buffers"
    },
    {
        title: "Finally, give up gin, echo and <your other framework>",
        titleRu: "Откажитесь уже наконец от gin, echo и <иной ваш фреймворк>",
        slug: "finally-give-up-gin-echo",
        date: "2025-11-29",
        excerpt: "Why it's time to reconsider your web framework choices",
        excerptRu: "Почему пора пересмотреть выбор вашего веб-фреймворка"
    }
]

export default function BlogPage() {
    const { i18n } = useTranslation()
    const isRu = i18n.language === 'ru'

    return (
        <div className="min-h-screen bg-gradient-to-b from-slate-950 via-slate-900 to-slate-950">
            <Navbar />

            <div className="max-w-4xl mx-auto px-6 pt-32 pb-20">
                {/* Header */}
                <div className="mb-16">
                    <h1 className="text-5xl font-bold text-white mb-4">
                        {isRu ? 'Блог' : 'Blog'}
                    </h1>
                    <p className="text-xl text-slate-400">
                        {isRu
                            ? 'Статьи о современных подходах к разработке'
                            : 'Articles about modern development approaches'}
                    </p>
                </div>

                {/* Blog Posts */}
                <div className="space-y-8">
                    {blogPosts.map((post) => (
                        <Link
                            key={post.slug}
                            to={`/blog/${post.slug}`}
                            className="group block glass-panel p-8 rounded-2xl border border-white/5 hover:border-primary/30 transition-all duration-300 hover:scale-[1.02]"
                        >
                            <div className="flex items-start justify-between gap-4 mb-4">
                                <h2 className="text-2xl font-bold text-white group-hover:text-primary transition-colors">
                                    {isRu ? post.titleRu : post.title}
                                </h2>
                                <ArrowRight className="w-6 h-6 text-slate-400 group-hover:text-primary group-hover:translate-x-1 transition-all flex-shrink-0" />
                            </div>

                            <p className="text-slate-400 mb-4">
                                {isRu ? post.excerptRu : post.excerpt}
                            </p>

                            <div className="flex items-center text-sm text-slate-500">
                                <Calendar className="w-4 h-4 mr-2" />
                                {new Date(post.date).toLocaleDateString(isRu ? 'ru-RU' : 'en-US', {
                                    year: 'numeric',
                                    month: 'long',
                                    day: 'numeric'
                                })}
                            </div>
                        </Link>
                    ))}
                </div>

                {/* Empty state if no posts */}
                {blogPosts.length === 0 && (
                    <div className="text-center py-20">
                        <p className="text-slate-400 text-lg">
                            {isRu ? 'Статьи скоро появятся' : 'Articles coming soon'}
                        </p>
                    </div>
                )}
            </div>
        </div>
    )
}
