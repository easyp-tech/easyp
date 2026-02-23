import { useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Loader2, Search, X } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { resolveSearchLanguage, searchDocumentation } from '../../lib/search/search'
import type { SearchHit, SearchResult } from '../../lib/search/types'

const MIN_QUERY_LENGTH = 2
const MAX_RESULTS = 20
const SEARCH_DEBOUNCE_MS = 120

interface ResultsProps {
    query: string
    isLoading: boolean
    hasError: boolean
    results: SearchResult[]
    expandedPages: Set<string>
    onSelectResult: (result: SearchHit) => void
    onTogglePageMatches: (pagePath: string) => void
    t: (key: string, options?: Record<string, unknown>) => string
}

function SearchResults({
    query,
    isLoading,
    hasError,
    results,
    expandedPages,
    onSelectResult,
    onTogglePageMatches,
    t,
}: ResultsProps) {
    if (query.trim().length < MIN_QUERY_LENGTH) {
        return (
            <p className="docs-search-state docs-search-state--hint">
                {t('docsSearch.typeToSearch')}
            </p>
        )
    }

    if (isLoading) {
        return (
            <div className="docs-search-state docs-search-state--loading">
                <Loader2 className="h-4 w-4 animate-spin" />
                <span>{t('docsSearch.loading')}</span>
            </div>
        )
    }

    if (hasError) {
        return (
            <p className="docs-search-state docs-search-state--error">
                {t('docsSearch.error')}
            </p>
        )
    }

    if (results.length === 0) {
        return (
            <p className="docs-search-state docs-search-state--empty">
                {t('docsSearch.noResults')}
            </p>
        )
    }

    return (
        <ul className="docs-search-results-list">
            {results.map((result) => (
                <li key={result.id}>
                    <button
                        type="button"
                        className="docs-search-result-item"
                        onClick={() => onSelectResult(result)}
                    >
                        <span className="docs-search-result-title">{result.title}</span>
                        {result.section && (
                            <span className="docs-search-result-section">{result.section}</span>
                        )}
                        <span className="docs-search-result-excerpt">{result.preview}</span>
                    </button>
                    {result.moreMatches && result.moreMatches.length > 0 && (
                        <div className="docs-search-more-wrapper">
                            <button
                                type="button"
                                className="docs-search-more-toggle"
                                onClick={() => onTogglePageMatches(result.pagePath)}
                            >
                                {expandedPages.has(result.pagePath)
                                    ? t('docsSearch.hideMoreMatches')
                                    : t('docsSearch.showMoreMatches', { count: result.moreMatches.length })}
                            </button>

                            {expandedPages.has(result.pagePath) && (
                                <ul className="docs-search-more-list">
                                    {result.moreMatches.map((match) => (
                                        <li key={match.id}>
                                            <button
                                                type="button"
                                                className="docs-search-result-item docs-search-result-item--nested"
                                                onClick={() => onSelectResult(match)}
                                            >
                                                {match.section && (
                                                    <span className="docs-search-result-section">{match.section}</span>
                                                )}
                                                <span className="docs-search-result-excerpt">{match.preview}</span>
                                            </button>
                                        </li>
                                    ))}
                                </ul>
                            )}
                        </div>
                    )}
                </li>
            ))}
        </ul>
    )
}

export default function DocsSearch() {
    const navigate = useNavigate()
    const { t, i18n } = useTranslation()
    const desktopContainerRef = useRef<HTMLDivElement>(null)
    const mobileInputRef = useRef<HTMLInputElement>(null)

    const [desktopQuery, setDesktopQuery] = useState('')
    const [mobileQuery, setMobileQuery] = useState('')
    const [mobileOpen, setMobileOpen] = useState(false)
    const [showDesktopDropdown, setShowDesktopDropdown] = useState(false)
    const [isLoading, setIsLoading] = useState(false)
    const [hasError, setHasError] = useState(false)
    const [results, setResults] = useState<SearchResult[]>([])
    const [expandedPages, setExpandedPages] = useState<Set<string>>(new Set())

    const searchLanguage = useMemo(
        () => resolveSearchLanguage(i18n.language),
        [i18n.language],
    )

    const activeQuery = mobileOpen ? mobileQuery : desktopQuery

    useEffect(() => {
        const query = activeQuery.trim()

        if (query.length < MIN_QUERY_LENGTH) {
            setResults([])
            setIsLoading(false)
            setHasError(false)
            setExpandedPages(new Set())
            return
        }

        let cancelled = false
        const timer = window.setTimeout(async () => {
            setIsLoading(true)
            setHasError(false)

            try {
                const searchResults = await searchDocumentation({
                    query,
                    language: searchLanguage,
                    limit: MAX_RESULTS,
                })
                if (!cancelled) {
                    setResults(searchResults)
                    setExpandedPages(new Set())
                }
            } catch (error) {
                console.error('Search failed:', error)
                if (!cancelled) {
                    setHasError(true)
                    setResults([])
                }
            } finally {
                if (!cancelled) {
                    setIsLoading(false)
                }
            }
        }, SEARCH_DEBOUNCE_MS)

        return () => {
            cancelled = true
            window.clearTimeout(timer)
        }
    }, [activeQuery, searchLanguage])

    useEffect(() => {
        const onDocumentClick = (event: MouseEvent) => {
            if (!desktopContainerRef.current) {
                return
            }

            if (!desktopContainerRef.current.contains(event.target as Node)) {
                setShowDesktopDropdown(false)
            }
        }

        document.addEventListener('mousedown', onDocumentClick)
        return () => {
            document.removeEventListener('mousedown', onDocumentClick)
        }
    }, [])

    useEffect(() => {
        const onEsc = (event: KeyboardEvent) => {
            if (event.key !== 'Escape') {
                return
            }
            setShowDesktopDropdown(false)
            setMobileOpen(false)
        }

        document.addEventListener('keydown', onEsc)
        return () => {
            document.removeEventListener('keydown', onEsc)
        }
    }, [])

    useEffect(() => {
        if (!mobileOpen) {
            return
        }

        const frame = window.requestAnimationFrame(() => {
            mobileInputRef.current?.focus()
        })

        return () => {
            window.cancelAnimationFrame(frame)
        }
    }, [mobileOpen])

    const selectResult = (result: SearchHit) => {
        navigate(result.path, {
            state: {
                docSearchQuery: activeQuery.trim(),
            },
        })
        setDesktopQuery('')
        setMobileQuery('')
        setResults([])
        setShowDesktopDropdown(false)
        setMobileOpen(false)
        setExpandedPages(new Set())
    }

    const togglePageMatches = (pagePath: string) => {
        setExpandedPages((prev) => {
            const next = new Set(prev)
            if (next.has(pagePath)) {
                next.delete(pagePath)
            } else {
                next.add(pagePath)
            }
            return next
        })
    }

    const openMobileSearch = () => {
        setMobileQuery(desktopQuery)
        setMobileOpen(true)
    }

    const closeMobileSearch = () => {
        setMobileOpen(false)
    }

    const showDropdown = showDesktopDropdown && desktopQuery.trim().length >= MIN_QUERY_LENGTH

    return (
        <>
            <div ref={desktopContainerRef} className="docs-search-desktop hidden md:block">
                <div className="docs-search-input-wrapper">
                    <Search className="docs-search-input-icon" />
                    <input
                        type="text"
                        value={desktopQuery}
                        onChange={(event) => {
                            const nextValue = event.target.value
                            setDesktopQuery(nextValue)
                            if (nextValue.trim().length >= MIN_QUERY_LENGTH) {
                                setShowDesktopDropdown(true)
                            } else {
                                setShowDesktopDropdown(false)
                            }
                        }}
                        onFocus={() => {
                            if (desktopQuery.trim().length >= MIN_QUERY_LENGTH) {
                                setShowDesktopDropdown(true)
                            }
                        }}
                        placeholder={t('docsSearch.placeholder')}
                        className="docs-search-input"
                    />
                    {desktopQuery && (
                        <button
                            type="button"
                            className="docs-search-clear"
                            onClick={() => {
                                setDesktopQuery('')
                                setShowDesktopDropdown(false)
                                setExpandedPages(new Set())
                            }}
                            aria-label={t('docsSearch.clear')}
                        >
                            <X className="h-4 w-4" />
                        </button>
                    )}
                </div>

                {showDropdown && (
                    <div className="docs-search-dropdown">
                        <SearchResults
                            query={desktopQuery}
                            isLoading={isLoading}
                            hasError={hasError}
                            results={results}
                            expandedPages={expandedPages}
                            onSelectResult={selectResult}
                            onTogglePageMatches={togglePageMatches}
                            t={t}
                        />
                    </div>
                )}
            </div>

            <div className="md:hidden">
                <button
                    type="button"
                    className="docs-search-mobile-trigger"
                    onClick={openMobileSearch}
                    aria-label={t('docsSearch.open')}
                >
                    <Search className="h-5 w-5" />
                </button>
            </div>

            {mobileOpen && (
                <div className="docs-search-mobile-overlay md:hidden">
                    <button
                        type="button"
                        className="docs-search-mobile-backdrop"
                        onClick={closeMobileSearch}
                        aria-label={t('docsSearch.close')}
                    />
                    <div className="docs-search-mobile-sheet">
                        <div className="docs-search-mobile-top">
                            <div className="docs-search-input-wrapper docs-search-input-wrapper--mobile">
                                <Search className="docs-search-input-icon" />
                                <input
                                    ref={mobileInputRef}
                                    type="text"
                                    value={mobileQuery}
                                    onChange={(event) => {
                                        setMobileQuery(event.target.value)
                                    }}
                                    placeholder={t('docsSearch.mobilePlaceholder')}
                                    className="docs-search-input"
                                />
                                {mobileQuery && (
                                    <button
                                        type="button"
                                        className="docs-search-clear"
                                        onClick={() => {
                                            setMobileQuery('')
                                            setExpandedPages(new Set())
                                        }}
                                        aria-label={t('docsSearch.clear')}
                                    >
                                        <X className="h-4 w-4" />
                                    </button>
                                )}
                            </div>

                            <button
                                type="button"
                                className="docs-search-mobile-close"
                                onClick={closeMobileSearch}
                            >
                                {t('docsSearch.cancel')}
                            </button>
                        </div>

                        <div className="docs-search-mobile-results">
                            <SearchResults
                                query={mobileQuery}
                                isLoading={isLoading}
                                hasError={hasError}
                                results={results}
                                expandedPages={expandedPages}
                                onSelectResult={selectResult}
                                onTogglePageMatches={togglePageMatches}
                                t={t}
                            />
                        </div>
                    </div>
                </div>
            )}
        </>
    )
}
