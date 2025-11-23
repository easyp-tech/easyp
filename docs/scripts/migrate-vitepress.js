#!/usr/bin/env node

const fs = require('fs').promises;
const path = require('path');
const matter = require('gray-matter');

// Configuration
const CONFIG = {
    sourceDir: path.join(__dirname, '../public/docs'),
    targetDir: path.join(__dirname, '../public/docs-migrated'),
    backupDir: path.join(__dirname, '../public/docs-backup'),
    dryRun: process.argv.includes('--dry-run'),
    verbose: process.argv.includes('--verbose'),
    stats: {
        processed: 0,
        skipped: 0,
        errors: 0,
        modified: 0
    }
};

// Color codes for console output
const colors = {
    reset: '\x1b[0m',
    bright: '\x1b[1m',
    green: '\x1b[32m',
    yellow: '\x1b[33m',
    red: '\x1b[31m',
    blue: '\x1b[34m',
    cyan: '\x1b[36m'
};

// Logging helpers
const log = {
    info: (msg) => console.log(`${colors.blue}ℹ${colors.reset} ${msg}`),
    success: (msg) => console.log(`${colors.green}✓${colors.reset} ${msg}`),
    warning: (msg) => console.log(`${colors.yellow}⚠${colors.reset} ${msg}`),
    error: (msg) => console.error(`${colors.red}✗${colors.reset} ${msg}`),
    verbose: (msg) => CONFIG.verbose && console.log(`${colors.cyan}›${colors.reset} ${msg}`)
};

/**
 * Extract title from markdown content
 */
function extractTitle(content) {
    const match = content.match(/^#\s+(.+)$/m);
    return match ? match[1].trim() : 'Untitled';
}

/**
 * Convert VitePress container blocks to HTML
 */
function convertContainerBlocks(content) {
    const containerTypes = ['tip', 'warning', 'danger', 'info', 'details'];
    let modified = false;

    containerTypes.forEach(type => {
        const regex = new RegExp(`:::\\s*${type}(?:\\s+(.*))?\\n([\\s\\S]*?)\\n:::`, 'g');
        const matches = content.match(regex);
        if (matches) {
            modified = true;
            content = content.replace(regex, (match, title, body) => {
                const titleHtml = title ? `\n<p class="custom-block-title">${title}</p>\n` : '';
                return `<div class="${type} custom-block">${titleHtml}${body}\n</div>`;
            });
        }
    });

    return { content, modified };
}

/**
 * Convert VitePress code group blocks
 */
function convertCodeGroups(content) {
    let modified = false;

    // Convert code-group blocks
    const codeGroupRegex = /:::code-group\n([\s\S]*?)\n:::/g;
    const matches = content.match(codeGroupRegex);

    if (matches) {
        modified = true;
        content = content.replace(codeGroupRegex, (match, groupContent) => {
            // Parse individual code blocks within the group
            const codeBlocks = [];
            const blockRegex = /```(\w+)(?:\s+\[(.+?)\])?\n([\s\S]*?)```/g;
            let blockMatch;

            while ((blockMatch = blockRegex.exec(groupContent)) !== null) {
                const [, lang, label, code] = blockMatch;
                codeBlocks.push({
                    lang,
                    label: label || lang,
                    code: code.trim()
                });
            }

            // Convert to tabs structure (will need custom component)
            if (codeBlocks.length > 0) {
                let result = '<div class="code-group">\n';
                result += '<div class="code-group-tabs">\n';
                codeBlocks.forEach((block, i) => {
                    result += `  <button class="code-group-tab${i === 0 ? ' active' : ''}" data-tab="${i}">${block.label}</button>\n`;
                });
                result += '</div>\n';
                codeBlocks.forEach((block, i) => {
                    result += `<div class="code-group-content${i === 0 ? ' active' : ''}" data-content="${i}">\n\n`;
                    result += '```' + block.lang + '\n';
                    result += block.code + '\n';
                    result += '```\n\n</div>\n';
                });
                result += '</div>';
                return result;
            }

            return match;
        });
    }

    return { content, modified };
}

/**
 * Add or update frontmatter
 */
function ensureFrontmatter(content, filePath) {
    const parsed = matter(content);
    let modified = false;

    // Add title if missing
    if (!parsed.data.title) {
        parsed.data.title = extractTitle(parsed.content);
        modified = true;
    }

    // Add toc if content has multiple headings and no explicit toc setting
    if (parsed.data.toc === undefined) {
        const headingCount = (parsed.content.match(/^#{2,4}\s+/gm) || []).length;
        if (headingCount >= 3 || parsed.content.includes('[[toc]]')) {
            parsed.data.toc = true;
            modified = true;
        }
    }

    // Add description if missing (extract from first paragraph)
    if (!parsed.data.description) {
        const firstParagraph = parsed.content.match(/^(?!#)(?!\s*$)(.+)$/m);
        if (firstParagraph) {
            parsed.data.description = firstParagraph[1].substring(0, 160).trim();
            modified = true;
        }
    }

    return {
        content: matter.stringify(parsed.content, parsed.data),
        modified
    };
}

/**
 * Convert VitePress specific syntax
 */
function convertVitePressSpecificSyntax(content) {
    let modified = false;
    let result = content;

    // Convert badge syntax
    if (result.includes('<Badge')) {
        modified = true;
        result = result.replace(
            /<Badge\s+type="(\w+)"(?:\s+text="([^"]+)")?\s*\/>/g,
            '<span class="badge badge-$1">$2</span>'
        );
    }

    // Convert custom anchor links {#custom-id}
    const anchorRegex = /^(#{1,6}\s+.+?)\s*\{#([\w-]+)\}\s*$/gm;
    if (anchorRegex.test(result)) {
        modified = true;
        result = result.replace(anchorRegex, '$1');
        log.verbose('Note: Custom anchor IDs detected. These will be auto-generated from heading text.');
    }

    // Convert VitePress table of contents placeholder
    if (result.includes('[[toc]]') || result.includes('[toc]')) {
        modified = true;
        result = result.replace(/\[?\[toc\]\]?/gi, '[[toc]]');
    }

    return { content: result, modified };
}

/**
 * Process a single markdown file
 */
async function processMarkdownFile(filePath) {
    try {
        const relativePath = path.relative(CONFIG.sourceDir, filePath);
        log.verbose(`Processing: ${relativePath}`);

        const content = await fs.readFile(filePath, 'utf-8');
        let processedContent = content;
        let wasModified = false;

        // Apply transformations
        const containerResult = convertContainerBlocks(processedContent);
        processedContent = containerResult.content;
        wasModified = wasModified || containerResult.modified;

        const codeGroupResult = convertCodeGroups(processedContent);
        processedContent = codeGroupResult.content;
        wasModified = wasModified || codeGroupResult.modified;

        const syntaxResult = convertVitePressSpecificSyntax(processedContent);
        processedContent = syntaxResult.content;
        wasModified = wasModified || syntaxResult.modified;

        const frontmatterResult = ensureFrontmatter(processedContent, filePath);
        processedContent = frontmatterResult.content;
        wasModified = wasModified || frontmatterResult.modified;

        // Save the processed file
        const targetPath = path.join(CONFIG.targetDir, relativePath);

        if (!CONFIG.dryRun) {
            await fs.mkdir(path.dirname(targetPath), { recursive: true });
            await fs.writeFile(targetPath, processedContent, 'utf-8');
        }

        if (wasModified) {
            CONFIG.stats.modified++;
            log.success(`Modified: ${relativePath}`);
        } else {
            CONFIG.stats.skipped++;
            CONFIG.verbose && log.info(`Unchanged: ${relativePath}`);
        }

        CONFIG.stats.processed++;

    } catch (error) {
        CONFIG.stats.errors++;
        log.error(`Error processing ${filePath}: ${error.message}`);
    }
}

/**
 * Recursively find all markdown files
 */
async function findMarkdownFiles(dir) {
    const files = [];
    const entries = await fs.readdir(dir, { withFileTypes: true });

    for (const entry of entries) {
        const fullPath = path.join(dir, entry.name);
        if (entry.isDirectory()) {
            files.push(...await findMarkdownFiles(fullPath));
        } else if (entry.isFile() && entry.name.endsWith('.md')) {
            files.push(fullPath);
        }
    }

    return files;
}

/**
 * Create backup of source directory
 */
async function createBackup() {
    if (!CONFIG.dryRun) {
        log.info('Creating backup...');
        try {
            // Remove old backup if exists
            await fs.rm(CONFIG.backupDir, { recursive: true, force: true });
            // Copy source to backup
            await fs.cp(CONFIG.sourceDir, CONFIG.backupDir, { recursive: true });
            log.success(`Backup created at: ${CONFIG.backupDir}`);
        } catch (error) {
            log.error(`Failed to create backup: ${error.message}`);
            throw error;
        }
    }
}

/**
 * Main migration function
 */
async function migrate() {
    console.log(`${colors.bright}${colors.blue}
╔══════════════════════════════════════════════╗
║     VitePress → React Markdown Migration     ║
╚══════════════════════════════════════════════╝
${colors.reset}`);

    // Parse command line arguments
    if (process.argv.includes('--help')) {
        console.log(`
Usage: node migrate-vitepress.js [options]

Options:
  --dry-run    Preview changes without writing files
  --verbose    Show detailed output
  --help       Show this help message

Source: ${CONFIG.sourceDir}
Target: ${CONFIG.targetDir}
        `);
        process.exit(0);
    }

    if (CONFIG.dryRun) {
        log.warning('DRY RUN MODE - No files will be modified');
    }

    try {
        // Check if source directory exists
        await fs.access(CONFIG.sourceDir);

        // Create backup
        await createBackup();

        // Find all markdown files
        log.info('Scanning for markdown files...');
        const files = await findMarkdownFiles(CONFIG.sourceDir);
        log.info(`Found ${files.length} markdown files`);

        // Process each file
        for (const file of files) {
            await processMarkdownFile(file);
        }

        // Print summary
        console.log(`\n${colors.bright}Migration Summary:${colors.reset}`);
        console.log(`${colors.green}✓ Processed:${colors.reset} ${CONFIG.stats.processed}`);
        console.log(`${colors.yellow}✎ Modified:${colors.reset} ${CONFIG.stats.modified}`);
        console.log(`${colors.cyan}⊘ Skipped:${colors.reset} ${CONFIG.stats.skipped}`);
        if (CONFIG.stats.errors > 0) {
            console.log(`${colors.red}✗ Errors:${colors.reset} ${CONFIG.stats.errors}`);
        }

        if (!CONFIG.dryRun) {
            log.success(`\nMigration complete! Files saved to: ${CONFIG.targetDir}`);
            log.info('Original files backed up to: ' + CONFIG.backupDir);
        } else {
            log.info('\nRun without --dry-run to apply changes');
        }

    } catch (error) {
        log.error(`Migration failed: ${error.message}`);
        process.exit(1);
    }
}

// Run migration
migrate();
