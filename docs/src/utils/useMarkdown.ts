import { useState, useEffect } from 'react'

interface UseMarkdownResult {
    content: string
    isLoading: boolean
    error: string | null
}

export function useMarkdown(path: string): UseMarkdownResult {
    const [content, setContent] = useState('')
    const [isLoading, setIsLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        let isMounted = true

        async function loadMarkdown() {
            try {
                setIsLoading(true)
                setError(null)

                // Construct path to MD file in public folder
                const mdPath = `/docs/${path}.md`
                const response = await fetch(mdPath)

                if (!response.ok) {
                    throw new Error(`Failed to load documentation: ${response.statusText}`)
                }

                const text = await response.text()

                if (isMounted) {
                    setContent(text)
                    setIsLoading(false)
                }
            } catch (err) {
                if (isMounted) {
                    setError(err instanceof Error ? err.message : 'Failed to load documentation')
                    setIsLoading(false)
                }
            }
        }

        if (path) {
            loadMarkdown()
        }

        return () => {
            isMounted = false
        }
    }, [path])

    return { content, isLoading, error }
}
