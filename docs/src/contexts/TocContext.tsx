import { createContext, useContext, useState, ReactNode } from 'react'
import type { TocItem } from '../lib/markdown/types'

interface TocContextValue {
    toc: TocItem[]
    setToc: (toc: TocItem[]) => void
    activeId: string
    setActiveId: (id: string) => void
}

const TocContext = createContext<TocContextValue | null>(null)

export function useToc() {
    const context = useContext(TocContext)
    return context
}

interface TocProviderProps {
    children: ReactNode
}

export function TocProvider({ children }: TocProviderProps) {
    const [toc, setToc] = useState<TocItem[]>([])
    const [activeId, setActiveId] = useState<string>('')

    const value: TocContextValue = {
        toc,
        setToc,
        activeId,
        setActiveId
    }

    return <TocContext.Provider value={value}>{children}</TocContext.Provider>
}
