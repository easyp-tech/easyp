export interface SidebarItem {
    title: string
    path?: string
    icon?: string
    children?: SidebarItem[]
}

export type SidebarConfig = SidebarItem[]
