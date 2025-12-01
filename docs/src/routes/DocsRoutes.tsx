import { Routes, Route } from 'react-router-dom'
import DocsLayout from '../components/docs/DocsLayout'
import MarkdownPage from '../components/docs/MarkdownPage'
import NotFound from '../components/docs/NotFound'
import { TocProvider } from '../contexts/TocContext'

export default function DocsRoutes() {
    return (
        <TocProvider>
            <Routes>
                <Route path="/" element={<DocsLayout />}>
                    {/* Index route - shows introduction/what-is.md */}
                    <Route index element={<MarkdownPage path="guide/introduction/what-is" />} />

                    {/* Docs routes - paths already include guide/ */}
                    <Route path=":category/:page" element={<MarkdownPage />} />
                    <Route path=":category/:subcategory/:page" element={<MarkdownPage />} />
                    <Route path=":category/:subcategory/:subsubcategory/:page" element={<MarkdownPage />} />
                    <Route path=":category/:subcategory/:subsubcategory/:section/:page" element={<MarkdownPage />} />

                    {/* 404 for invalid docs paths */}
                    <Route path="*" element={<NotFound />} />
                </Route>
            </Routes>
        </TocProvider>
    )
}
