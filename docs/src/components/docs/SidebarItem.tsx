import { Link } from 'react-router-dom'
import { ChevronDown, ChevronRight } from 'lucide-react'
import * as Icons from 'lucide-react'
import type { SidebarItem as SidebarItemType } from '../../types/sidebar'

interface SidebarItemProps {
    item: SidebarItemType
    level: number
    isExpanded: boolean
    isActive: boolean
    onToggle?: () => void
}

export default function SidebarItem({
    item,
    level,
    isExpanded,
    isActive,
    onToggle
}: SidebarItemProps) {
    const hasChildren = item.children && item.children.length > 0
    const Icon = item.icon ? (Icons as any)[item.icon] : null

    const itemClasses = `
        sidebar-item
        sidebar-item--level-${level}
        ${isActive ? 'sidebar-item--active' : ''}
        ${hasChildren ? 'sidebar-item--parent' : ''}
    `.trim()

    const content = (
        <div className={itemClasses}>
            <div className="sidebar-item__content">
                {/* Иконка */}
                {Icon && level === 0 && (
                    <Icon className="sidebar-item__icon" size={18} />
                )}

                {/* Заголовок */}
                {item.path ? (
                    <Link
                        to={item.path}
                        className="sidebar-item__link"
                    >
                        {item.title}
                    </Link>
                ) : (
                    <span
                        className="sidebar-item__text"
                        onClick={() => {
                            // Если нет path но есть children, разворачиваем при клике
                            if (hasChildren) {
                                onToggle?.()
                            }
                        }}
                        style={{ cursor: hasChildren ? 'pointer' : 'default' }}
                    >
                        {item.title}
                    </span>
                )}

                {/* Кнопка разворачивания */}
                {hasChildren && (
                    <button
                        className="sidebar-item__toggle"
                        onClick={(e) => {
                            e.preventDefault()
                            onToggle?.()
                        }}
                        aria-label={isExpanded ? 'Collapse' : 'Expand'}
                    >
                        {isExpanded ? (
                            <ChevronDown size={16} />
                        ) : (
                            <ChevronRight size={16} />
                        )}
                    </button>
                )}
            </div>
        </div>
    )

    return (
        <div className="sidebar-item-wrapper">
            {content}
        </div>
    )
}
