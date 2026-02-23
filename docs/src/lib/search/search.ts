import Fuse, { type FuseResult, type IFuseOptions } from 'fuse.js'
import type { SearchDocument, SearchHit, SearchLanguage, SearchParams, SearchResult } from './types'

const INDEX_BY_LANGUAGE: Record<SearchLanguage, string> = {
    en: '/search/index.en.json',
    ru: '/search/index.ru.json',
}

const DEFAULT_LIMIT = 10
const MIN_QUERY_LENGTH = 2
const PREVIEW_CONTEXT_BEFORE = 55
const PREVIEW_CONTEXT_AFTER = 145
const MAX_RESULTS_PER_PAGE = 3
const RAW_RESULT_MULTIPLIER = 12
const RAW_RESULT_FLOOR = 120

const documentCache = new Map<SearchLanguage, Promise<SearchDocument[]>>()
const fuseCache = new Map<SearchLanguage, Promise<Fuse<SearchDocument>>>()

const fuseOptions: IFuseOptions<SearchDocument> = {
    includeScore: true,
    includeMatches: true,
    ignoreLocation: true,
    threshold: 0.35,
    minMatchCharLength: MIN_QUERY_LENGTH,
    keys: [
        { name: 'title', weight: 0.45 },
        { name: 'headings', weight: 0.3 },
        { name: 'content', weight: 0.25 },
    ],
}

export function resolveSearchLanguage(language: string): SearchLanguage {
    const normalized = language.toLowerCase()
    if (normalized === 'ru' || normalized.startsWith('ru-')) {
        return 'ru'
    }
    return 'en'
}

async function fetchSearchIndex(language: SearchLanguage): Promise<SearchDocument[]> {
    const response = await fetch(INDEX_BY_LANGUAGE[language], {
        cache: 'no-cache',
    })

    if (!response.ok) {
        throw new Error(`Failed to load search index for "${language}"`)
    }

    const payload: unknown = await response.json()

    if (Array.isArray(payload)) {
        return payload as SearchDocument[]
    }

    if (
        payload &&
        typeof payload === 'object' &&
        'documents' in payload &&
        Array.isArray((payload as { documents: unknown }).documents)
    ) {
        return (payload as { documents: SearchDocument[] }).documents
    }

    throw new Error(`Invalid search index format for "${language}"`)
}

async function getDocuments(language: SearchLanguage): Promise<SearchDocument[]> {
    let loader = documentCache.get(language)
    if (!loader) {
        loader = fetchSearchIndex(language)
        documentCache.set(language, loader)
    }
    return loader
}

async function getFuseInstance(language: SearchLanguage): Promise<Fuse<SearchDocument>> {
    let fusePromise = fuseCache.get(language)
    if (!fusePromise) {
        fusePromise = getDocuments(language).then((documents) => new Fuse(documents, fuseOptions))
        fuseCache.set(language, fusePromise)
    }
    return fusePromise
}

function getQueryTokens(query: string): string[] {
    return query
        .toLowerCase()
        .split(/\s+/)
        .map((token) => token.trim())
        .filter((token) => token.length >= MIN_QUERY_LENGTH)
}

function documentContainsAllTokens(document: SearchDocument, tokens: string[]): boolean {
    const haystack = `${document.title}\n${document.section || ''}\n${document.headings.join('\n')}\n${document.content}`.toLowerCase()
    return tokens.every((token) => haystack.includes(token))
}

function buildSnippet(value: string, start: number, tokenLength: number): string {
    const from = Math.max(0, start - PREVIEW_CONTEXT_BEFORE)
    const to = Math.min(value.length, start + tokenLength + PREVIEW_CONTEXT_AFTER)

    const prefix = from > 0 ? '...' : ''
    const suffix = to < value.length ? '...' : ''
    const snippet = value.slice(from, to).trim()

    return snippet ? `${prefix}${snippet}${suffix}` : value.slice(0, PREVIEW_CONTEXT_AFTER).trim()
}

function buildPreviewFromMatch(result: FuseResult<SearchDocument>, queryTokens: string[]): string {
    const fallback = result.item.excerpt
    const fields = [
        result.item.content,
        result.item.section || '',
        result.item.title,
        ...result.item.headings,
    ]

    for (const field of fields) {
        const lowerField = field.toLowerCase()
        for (const token of queryTokens) {
            const exactIndex = lowerField.indexOf(token)
            if (exactIndex >= 0) {
                return buildSnippet(field, exactIndex, token.length)
            }
        }
    }

    const match = result.matches?.find(
        (entry) => typeof entry.value === 'string' && entry.indices && entry.indices.length > 0,
    )
    if (!match || typeof match.value !== 'string' || !match.indices || match.indices.length === 0) {
        return fallback
    }

    const [start, end] = match.indices[0]
    return buildSnippet(match.value, start, end - start + 1) || fallback
}

export async function searchDocumentation({
    query,
    language,
    limit = DEFAULT_LIMIT,
}: SearchParams): Promise<SearchResult[]> {
    const normalizedQuery = query.trim()
    if (normalizedQuery.length < MIN_QUERY_LENGTH) {
        return []
    }
    const queryTokens = getQueryTokens(normalizedQuery)

    const fuse = await getFuseInstance(language)
    const rawLimit = Math.max(limit * RAW_RESULT_MULTIPLIER, RAW_RESULT_FLOOR)
    const results = fuse.search(normalizedQuery, { limit: rawLimit })
    const dedupe = new Set<string>()

    const hits: SearchHit[] = results
        .filter((result) => documentContainsAllTokens(result.item, queryTokens))
        .map((result) => ({
            ...result.item,
            score: result.score ?? 1,
            preview: buildPreviewFromMatch(result, queryTokens),
            pagePath: result.item.path.split('#')[0],
        }))
        .filter((result) => {
            const key = `${result.path}::${result.preview}`
            if (dedupe.has(key)) {
                return false
            }
            dedupe.add(key)
            return true
        })

    const perPageBuckets = new Map<string, { visible: SearchHit[]; hidden: SearchHit[] }>()
    for (const hit of hits) {
        const bucket = perPageBuckets.get(hit.pagePath) || { visible: [], hidden: [] }
        if (bucket.visible.length < MAX_RESULTS_PER_PAGE) {
            bucket.visible.push(hit)
        } else {
            bucket.hidden.push(hit)
        }
        perPageBuckets.set(hit.pagePath, bucket)
    }

    const visibleHitIds = new Set<string>()
    for (const bucket of perPageBuckets.values()) {
        for (const hit of bucket.visible) {
            visibleHitIds.add(hit.id)
        }
    }

    const hiddenByPage = new Map<string, SearchHit[]>()
    for (const [pagePath, bucket] of perPageBuckets.entries()) {
        hiddenByPage.set(pagePath, bucket.hidden)
    }

    const attachedHiddenToPage = new Set<string>()
    const mapped: SearchResult[] = hits
        .filter((hit) => visibleHitIds.has(hit.id))
        .map((hit) => {
            if (!attachedHiddenToPage.has(hit.pagePath)) {
                attachedHiddenToPage.add(hit.pagePath)
                const hidden = hiddenByPage.get(hit.pagePath) || []
                if (hidden.length > 0) {
                    return {
                        ...hit,
                        moreMatches: hidden,
                    }
                }
            }
            return hit
        })

    return mapped.slice(0, limit)
}
