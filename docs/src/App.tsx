import { BrowserRouter, Routes, Route } from 'react-router-dom'
import HomePage from './pages/HomePage'
import BlogPage from './pages/BlogPage'
import BlogPostPage from './pages/BlogPostPage'
import DocsRoutes from './routes/DocsRoutes'
import ErrorBoundary from './components/ErrorBoundary'

function App() {
    return (
        <BrowserRouter>
            <Routes>
                <Route path="/" element={<HomePage />} />
                <Route path="/blog" element={<BlogPage />} />
                <Route path="/blog/:slug" element={<BlogPostPage />} />
                <Route
                    path="/docs/*"
                    element={
                        <ErrorBoundary>
                            <DocsRoutes />
                        </ErrorBoundary>
                    }
                />
                <Route path="*" element={<div>404 Not Found</div>} />
            </Routes>
        </BrowserRouter>
    )
}

export default App
