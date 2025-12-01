export function parseBadGoodSections(markdown: string) {
    // Check if markdown contains both ### Bad and ### Good sections
    const hasBadSection = /###\s+Bad/i.test(markdown)
    const hasGoodSection = /###\s+Good/i.test(markdown)

    if (!hasBadSection || !hasGoodSection) {
        return { hasBadGood: false, badContent: '', goodContent: '', restContent: markdown }
    }

    // Split by ### Bad
    const parts = markdown.split(/###\s+Bad/i)
    if (parts.length < 2) {
        return { hasBadGood: false, badContent: '', goodContent: '', restContent: markdown }
    }

    const beforeBad = parts[0]
    const afterBad = parts[1]

    // Split the afterBad by ### Good
    const goodParts = afterBad.split(/###\s+Good/i)
    if (goodParts.length < 2) {
        return { hasBadGood: false, badContent: '', goodContent: '', restContent: markdown }
    }

    const badContent = goodParts[0].trim()
    const goodAndRest = goodParts[1]

    // Find where the Good section ends (next ### heading or end of string)
    const nextHeadingMatch = goodAndRest.match(/\n###\s+/)
    let goodContent: string
    let afterGood: string

    if (nextHeadingMatch) {
        const splitIndex = nextHeadingMatch.index!
        goodContent = goodAndRest.substring(0, splitIndex).trim()
        afterGood = goodAndRest.substring(splitIndex).trim()
    } else {
        goodContent = goodAndRest.trim()
        afterGood = ''
    }

    // Reconstruct the rest of the content (before Bad + after Good)
    const restContent = (beforeBad + (afterGood ? '\n\n' + afterGood : '')).trim()

    return {
        hasBadGood: true,
        badContent,
        goodContent,
        restContent
    }
}
