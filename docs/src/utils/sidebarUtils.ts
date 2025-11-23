import type { SidebarItem, SidebarConfig } from '../types/sidebar'
import sidebarEn from '../config/sidebar.en.json'
import sidebarRu from '../config/sidebar.ru.json'

const STORAGE_KEY = 'docs-sidebar-expanded'

/**
 * Загружает конфигурацию sidebar
 */
export function loadSidebarConfig(): SidebarConfig {
    // Определяем язык: сначала localStorage, затем i18next (если доступен), fallback 'en'
    let lang = 'en'
    try {
        const stored = localStorage.getItem('language')
        if (stored) {
            lang = stored
        } else if ((window as any).i18next?.language) {
            lang = (window as any).i18next.language
        }
    } catch {
        // ignore any access errors
    }

    // Выбор конфигурации по языку с fallback на английский
    if (lang === 'ru') {
        return (sidebarRu as SidebarConfig)
    }
    return (sidebarEn as SidebarConfig)
}

/**
 * Находит элемент sidebar по пути
 */
export function findItemByPath(
    config: SidebarConfig,
    path: string
): SidebarItem | null {
    for (const item of config) {
        if (item.path === path) {
            return item
        }
        if (item.children) {
            const found = findItemByPath(item.children, path)
            if (found) return found
        }
    }
    return null
}

/**
 * Получает все родительские пути для заданного пути
 * Используется для автоматического разворачивания секций
 */
export function getParentPaths(
    config: SidebarConfig,
    targetPath: string
): string[] {
    const parents: string[] = []

    function findParents(items: SidebarConfig, currentPath: string[] = []): boolean {
        for (const item of items) {
            const newPath = item.path ? [...currentPath, item.path] : currentPath

            if (item.path === targetPath) {
                parents.push(...currentPath)
                return true
            }

            if (item.children) {
                if (findParents(item.children, newPath)) {
                    if (item.path) {
                        parents.push(item.path)
                    }
                    return true
                }
            }
        }
        return false
    }

    findParents(config)
    return parents
}

/**
 * Генерирует уникальный ID для элемента sidebar на основе его пути
 */
export function getItemId(item: SidebarItem, index: number): string {
    return item.path || `section-${item.title.toLowerCase().replace(/\s+/g, '-')}-${index}`
}

/**
 * Сохраняет состояние развернутых секций в localStorage
 */
export function saveExpandedState(sections: Set<string>): void {
    try {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(Array.from(sections)))
    } catch (error) {
        console.error('Failed to save sidebar state:', error)
    }
}

/**
 * Загружает состояние развернутых секций из localStorage
 */
export function loadExpandedState(): Set<string> {
    try {
        const stored = localStorage.getItem(STORAGE_KEY)
        if (stored) {
            return new Set(JSON.parse(stored))
        }
    } catch (error) {
        console.error('Failed to load sidebar state:', error)
    }
    return new Set()
}

/**
 * Проверяет, является ли путь активным (точное совпадение или начало текущего пути)
 */
export function isPathActive(itemPath: string | undefined, currentPath: string): boolean {
    if (!itemPath) return false
    return currentPath === itemPath || currentPath.startsWith(itemPath + '/')
}

/**
 * Returns the previous and next navigation items for the current path
 */
export function getPrevNext(
    config: SidebarConfig,
    currentPath: string
): { prev: SidebarItem | null; next: SidebarItem | null } {
    const flatItems: SidebarItem[] = []

    function flatten(items: SidebarConfig) {
        for (const item of items) {
            if (item.path) {
                flatItems.push(item)
            }
            if (item.children) {
                flatten(item.children)
            }
        }
    }

    flatten(config)

    const index = flatItems.findIndex(item => item.path === currentPath)
    if (index === -1) {
        return { prev: null, next: null }
    }

    return {
        prev: index > 0 ? flatItems[index - 1] : null,
        next: index < flatItems.length - 1 ? flatItems[index + 1] : null
    }
}
