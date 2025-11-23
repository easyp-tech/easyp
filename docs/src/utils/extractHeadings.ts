export interface Heading {
    id: string
    text: string
    level: number
}

export function extractHeadings(content: string): Heading[] {
    const headings: Heading[] = []
    const lines = content.split('\n')

    // Simple regex for ## and ### headings
    // Ignoring # (h1) as it's usually the page title
    const headingRegex = /^(#{2,3})\s+(.+)$/

    // Helper to generate ID from text
    const generateId = (text: string) => {
        return text
            .toLowerCase()
            .replace(/[^\w\s-]/g, '')
            .replace(/\s+/g, '-')
    }

    lines.forEach(line => {
        const match = line.match(headingRegex)
        if (match) {
            const level = match[1].length
            const text = match[2].trim()
            const id = generateId(text)
            headings.push({ id, text, level })
        }
    })

    return headings
}
