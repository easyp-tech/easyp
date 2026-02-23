export type SearchLanguage = 'en' | 'ru'

export interface SearchDocument {
    id: string
    title: string
    section?: string
    path: string
    headings: string[]
    excerpt: string
    content: string
}

export interface SearchHit extends SearchDocument {
    score: number
    preview: string
    pagePath: string
}

export interface SearchResult extends SearchHit {
    moreMatches?: SearchHit[]
}

export interface SearchParams {
    query: string
    language: SearchLanguage
    limit?: number
}
