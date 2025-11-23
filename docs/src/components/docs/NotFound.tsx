import { Link } from 'react-router-dom'
import { FileQuestion, Home } from 'lucide-react'

export default function NotFound() {
    return (
        <div className="flex flex-col items-center justify-center min-h-[60vh] px-4 text-center">
            <div className="mb-8">
                <FileQuestion className="w-24 h-24 text-gray-400 mx-auto mb-4" />
            </div>

            <h1 className="text-4xl font-bold text-gray-900 dark:text-white mb-4">
                Page Not Found
            </h1>

            <p className="text-lg text-gray-600 dark:text-gray-400 mb-8 max-w-md">
                The documentation page you're looking for doesn't exist.
                It may have been moved or deleted.
            </p>

            <div className="flex flex-col sm:flex-row gap-4">
                <Link
                    to="/docs"
                    className="inline-flex items-center gap-2 px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                >
                    <Home className="w-5 h-5" />
                    Go to Documentation Home
                </Link>

                <Link
                    to="/"
                    className="inline-flex items-center gap-2 px-6 py-3 border border-gray-300 dark:border-gray-700 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors"
                >
                    Back to Main Site
                </Link>
            </div>

            <div className="mt-12 text-sm text-gray-500 dark:text-gray-500">
                <p className="mb-2">Looking for something specific?</p>
                <ul className="space-y-1">
                    <li><Link to="/docs/introduction/what-is" className="text-blue-600 hover:underline">What is EasyP?</Link></li>
                    <li><Link to="/docs/introduction/quickstart" className="text-blue-600 hover:underline">Quickstart Guide</Link></li>
                    <li><Link to="/docs/introduction/install" className="text-blue-600 hover:underline">Installation</Link></li>
                </ul>
            </div>
        </div>
    )
}
