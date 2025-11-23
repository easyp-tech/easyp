import { useEffect, useState } from 'react'
import { useLocation, Link } from 'react-router-dom'
import SidebarItem from './SidebarItem'
import type { SidebarConfig } from '../../types/sidebar'
import {
    loadSidebarConfig,
    getParentPaths,
    getItemId,
    saveExpandedState,
    loadExpandedState,
    isPathActive
} from '../../utils/sidebarUtils'

interface DocsSidebarProps {
    isOpen: boolean
    onClose: () => void
}

export default function DocsSidebar({ isOpen, onClose }: DocsSidebarProps) {
    const location = useLocation()
    const [sidebarData, setSidebarData] = useState<SidebarConfig>([])
    const [expandedSections, setExpandedSections] = useState<Set<string>>(new Set())

    // Загрузка конфигурации sidebar
    useEffect(() => {
        const config = loadSidebarConfig()
        setSidebarData(config)
    }, [])

    // Загрузка сохраненного состояния из localStorage
    useEffect(() => {
        const saved = loadExpandedState()
        setExpandedSections(saved)
    }, [])

    // Автоматическое разворачивание секции с активной страницей
    useEffect(() => {
        if (sidebarData.length > 0) {
            const parentPaths = getParentPaths(sidebarData, location.pathname)
            if (parentPaths.length > 0) {
                setExpandedSections((prev) => {
                    const newSet = new Set(prev)
                    parentPaths.forEach((path) => newSet.add(path))
                    return newSet
                })
            }
        }
    }, [location.pathname, sidebarData])

    // Сохранение состояния в localStorage при изменении
    useEffect(() => {
        saveExpandedState(expandedSections)
    }, [expandedSections])

    const toggleSection = (id: string) => {
        setExpandedSections((prev) => {
            const newSet = new Set(prev)
            if (newSet.has(id)) {
                newSet.delete(id)
            } else {
                newSet.add(id)
            }
            return newSet
        })
    }

    const renderItem = (item: any, index: number, level: number = 0) => {
        const itemId = getItemId(item, index)
        const isExpanded = expandedSections.has(itemId) || expandedSections.has(item.path || '')
        const isActive = isPathActive(item.path, location.pathname)
        const hasChildren = item.children && item.children.length > 0

        return (
            <div key={itemId}>
                <SidebarItem
                    item={item}
                    level={level}
                    isExpanded={isExpanded}
                    isActive={isActive}
                    onToggle={() => {
                        const toggleId = item.path || itemId
                        toggleSection(toggleId)
                    }}
                />

                {/* Рекурсивный рендеринг children */}
                {hasChildren && (
                    <div
                        className={`sidebar-children ${isExpanded ? 'sidebar-children--expanded' : ''
                            }`}
                    >
                        {isExpanded &&
                            item.children.map((child: any, childIndex: number) =>
                                renderItem(child, childIndex, level + 1)
                            )}
                    </div>
                )}
            </div>
        )
    }

    const sidebarClasses = `
        docs-sidebar
        ${isOpen ? 'docs-sidebar--open' : ''}
    `.trim()

    return (
        <>
            {/* Backdrop для мобильных */}
            {isOpen && (
                <div
                    className="docs-sidebar-backdrop"
                    onClick={onClose}
                    aria-hidden="true"
                />
            )}

            {/* Sidebar */}
            <aside className={sidebarClasses}>
                <div className="docs-sidebar-header">
                    <Link to="/" className="flex items-center gap-2 hover:opacity-80 transition-opacity">
                        <div className="w-8 h-8 bg-blue-500/20 rounded-lg flex items-center justify-center border border-blue-500/30">
                            <span className="font-bold text-blue-500 text-sm">EP</span>
                        </div>
                        <h2 className="text-xl font-bold text-gray-900 dark:text-white">
                            EasyP
                        </h2>
                    </Link>
                </div>

                <nav className="docs-sidebar-nav">
                    {sidebarData.map((item, index) => renderItem(item, index, 0))}
                </nav>
            </aside>
        </>
    )
}
