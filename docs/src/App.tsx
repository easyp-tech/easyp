import { BrowserRouter, Routes, Route } from 'react-router-dom'
import HomePage from './pages/HomePage'
import DocsRoutes from './routes/DocsRoutes'
import ErrorBoundary from './components/ErrorBoundary'

function App() {
    return (
        <BrowserRouter>
            <Routes>
                <Route path="/" element={<HomePage />} />
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

