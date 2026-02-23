import fs from 'node:fs/promises'
import path from 'node:path'
import matter from 'gray-matter'

const DOCS_ROOT = path.resolve('public/docs')
const OUTPUT_ROOT = path.resolve('public/search')

const LANG_CONFIGS = [
  { sourceDir: 'guide', outputFile: 'index.en.json' },
  { sourceDir: 'ru-guide', outputFile: 'index.ru.json' },
]

const EXCERPT_LENGTH = 220
const MIN_CHUNK_LENGTH = 30

async function getMarkdownFiles(dirPath) {
  const entries = await fs.readdir(dirPath, { withFileTypes: true })
  const files = await Promise.all(
    entries.map(async (entry) => {
      const fullPath = path.join(dirPath, entry.name)
      if (entry.isDirectory()) {
        return getMarkdownFiles(fullPath)
      }
      if (entry.isFile() && entry.name.endsWith('.md')) {
        return [fullPath]
      }
      return []
    }),
  )

  return files.flat().sort((a, b) => a.localeCompare(b))
}

function normalizeWhitespace(value) {
  return value.replace(/\r/g, '\n').replace(/\s+/g, ' ').trim()
}

function stripMarkdown(markdown) {
  return normalizeWhitespace(
    markdown
      .replace(/```[\s\S]*?```/g, ' ')
      .replace(/~~~[\s\S]*?~~~/g, ' ')
      .replace(/<!--[\s\S]*?-->/g, ' ')
      .replace(/!\[([^\]]*)\]\([^)]+\)/g, '$1')
      .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
      .replace(/\[([^\]]+)\]\[[^\]]*\]/g, '$1')
      .replace(/^#{1,6}\s+/gm, '')
      .replace(/^>\s?/gm, '')
      .replace(/^[-*+]\s+/gm, '')
      .replace(/^\d+\.\s+/gm, '')
      .replace(/`([^`]+)`/g, '$1')
      .replace(/\|/g, ' ')
      .replace(/<[^>]+>/g, ' ')
      .replace(/\n+/g, ' '),
  )
}

function stripHeadingFormatting(heading) {
  return normalizeWhitespace(
    heading
      .replace(/`([^`]+)`/g, '$1')
      .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
      .replace(/\*\*([^*]+)\*\*/g, '$1')
      .replace(/\*([^*]+)\*/g, '$1')
      .replace(/<[^>]+>/g, ' '),
  )
}

function generateHeadingId(text) {
  return text
    .toLowerCase()
    .replace(/[^\w\s-]/g, '')
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')
}

function splitIntoSections(content) {
  const lines = content.split('\n')
  const sections = []
  let current = { heading: null, anchor: '', lines: [] }

  for (const line of lines) {
    const headingMatch = line.match(/^(#{2,6})\s+(.+)$/)
    if (headingMatch) {
      if (current.heading || current.lines.length > 0) {
        sections.push(current)
      }

      const heading = stripHeadingFormatting(headingMatch[2].trim())
      current = {
        heading,
        anchor: generateHeadingId(heading),
        lines: [],
      }
      continue
    }

    current.lines.push(line)
  }

  if (current.heading || current.lines.length > 0) {
    sections.push(current)
  }

  return sections
    .map((section) => ({
      ...section,
      text: section.lines.join('\n').trim(),
    }))
    .filter((section) => section.heading || section.text.length > 0)
}

function splitSectionToChunks(sectionText) {
  const parts = sectionText
    .split(/\n{2,}/)
    .map((part) => part.trim())
    .filter(Boolean)

  return parts.length > 0 ? parts : [sectionText]
}

function extractTitle(frontmatter, content, relativePath) {
  if (typeof frontmatter.title === 'string' && frontmatter.title.trim()) {
    return frontmatter.title.trim()
  }

  const h1Match = content.match(/^#\s+(.+)$/m)
  if (h1Match?.[1]) {
    return stripHeadingFormatting(h1Match[1])
  }

  const fileName = path.basename(relativePath, '.md')
  return fileName
    .replace(/[-_]/g, ' ')
    .replace(/\b\w/g, (char) => char.toUpperCase())
}

function toDocsRoute(relativePath) {
  const withoutExt = relativePath.replace(/\.md$/, '')
  return `/docs/guide/${withoutExt.split(path.sep).join('/')}`
}

function buildChunkedDocuments(markdown, relativePath) {
  const parsed = matter(markdown)
  const content = parsed.content || markdown
  const basePath = toDocsRoute(relativePath)
  const pageId = relativePath.replace(/\.md$/, '').split(path.sep).join('/')
  const title = extractTitle(parsed.data, content, relativePath)
  const sections = splitIntoSections(content)

  const documents = []
  let chunkIndex = 0

  for (const section of sections) {
    const rawChunks = splitSectionToChunks(section.text || '')
    for (const rawChunk of rawChunks) {
      const chunkSource = section.heading
        ? `${section.heading}\n${rawChunk}`
        : rawChunk

      const cleanContent = stripMarkdown(chunkSource)
      if (cleanContent.length < MIN_CHUNK_LENGTH) {
        continue
      }

      const excerptSource =
        typeof parsed.data.description === 'string' && parsed.data.description.trim()
          ? parsed.data.description.trim()
          : cleanContent
      const excerpt = excerptSource.slice(0, EXCERPT_LENGTH)
      const pathWithHash = section.anchor ? `${basePath}#${section.anchor}` : basePath

      documents.push({
        id: `${pageId}::${section.anchor || 'top'}::${chunkIndex}`,
        title,
        section: section.heading || undefined,
        path: pathWithHash,
        headings: section.heading ? [section.heading] : [],
        excerpt,
        content: cleanContent,
      })
      chunkIndex += 1
    }
  }

  if (documents.length > 0) {
    return documents
  }

  const fallbackContent = stripMarkdown(content)
  return [{
    id: `${pageId}::top::0`,
    title,
    section: undefined,
    path: basePath,
    headings: [],
    excerpt: fallbackContent.slice(0, EXCERPT_LENGTH),
    content: fallbackContent,
  }]
}

async function buildIndexForLanguage(sourceDir) {
  const languageRoot = path.join(DOCS_ROOT, sourceDir)
  const markdownFiles = await getMarkdownFiles(languageRoot)

  const nestedDocuments = await Promise.all(
    markdownFiles.map(async (absoluteFilePath) => {
      const markdown = await fs.readFile(absoluteFilePath, 'utf8')
      const relativePath = path.relative(languageRoot, absoluteFilePath)
      return buildChunkedDocuments(markdown, relativePath)
    }),
  )

  return nestedDocuments.flat()
}

async function main() {
  await fs.mkdir(OUTPUT_ROOT, { recursive: true })

  for (const config of LANG_CONFIGS) {
    const documents = await buildIndexForLanguage(config.sourceDir)
    const outputPath = path.join(OUTPUT_ROOT, config.outputFile)
    await fs.writeFile(outputPath, JSON.stringify(documents), 'utf8')
    console.log(`Generated ${config.outputFile}: ${documents.length} documents`)
  }
}

main().catch((error) => {
  console.error('Failed to generate search index:', error)
  process.exit(1)
})
