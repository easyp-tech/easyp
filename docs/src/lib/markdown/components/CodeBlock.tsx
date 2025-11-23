import React, { useState, useEffect } from 'react';
import { Highlight } from 'prism-react-renderer';
import Prism from 'prismjs';
import { Copy, Check, Terminal } from 'lucide-react';
import type { CodeBlockProps } from '../types';
import { cyberTheme } from '../utils/prismTheme';

// Prevent Prism from automatically highlighting everything immediately
if (typeof window !== 'undefined') {
  (window as any).Prism = Prism;
}
Prism.manual = true;

// Singleton promise to ensure we only load languages once
let prismInitPromise: Promise<void> | null = null;
let prismReady = false;

// Try to load languages synchronously if possible
const initPrismSync = () => {
  if (typeof window === 'undefined') return false;

  try {
    (window as any).Prism = Prism;
    Prism.manual = true;

    // Try synchronous registration for common languages
    if (!Prism.languages.yaml) {
      // Basic YAML highlighting patterns
      Prism.languages.yaml = {
        'scalar': {
          pattern: /([^\s\x00-\x08\x0E-\x1F",[\]{}]|'(?:[^']|'')*'|"(?:\\.|[^\\"])*")(?:\s*(?:$|#))/,
          lookbehind: true,
          alias: 'string'
        },
        'comment': /#.*/,
        'key': {
          pattern: /(\s*(?:^|[:\-?\s])\s*)[^\s\x00-\x08\x0E-\x1F,[\]{}][^\s\x00-\x08\x0E-\x1F",[\]{}]*(?:\s+[^\s\x00-\x08\x0E-\x1F",[\]{}][^\s\x00-\x08\x0E-\x1F",[\]{}]*)*\s*(?=\s*:)/,
          lookbehind: true,
          alias: 'property'
        },
        'directive': {
          pattern: /^%\w.*$/m,
          alias: 'important'
        },
        'datetime': {
          pattern: /([:\-?\s]\s*)([0-9]{4}-[0-9]{1,2}-[0-9]{1,2}([Tt ][0-9]{1,2}:[0-9]{1,2}:[0-9]{1,2}(\.[0-9]+)?([ ]?[+-][0-9]{1,2}:[0-9]{1,2}|Z)?)?)/,
          lookbehind: true,
          alias: 'number'
        },
        'boolean': {
          pattern: /([:\-?\s]\s*)(?:true|false|yes|no|on|off)(?=\s*$)/m,
          lookbehind: true,
          alias: 'important'
        },
        'null': {
          pattern: /([:\-?\s]\s*)(?:null|~)(?=\s*$)/m,
          lookbehind: true,
          alias: 'important'
        },
        'string': {
          pattern: /([:\-?\s]\s*)(?:"(?:\\.|[^\\"])*"|'(?:[^']|'')*')(?=\s*$)/m,
          lookbehind: true
        },
        'number': {
          pattern: /([:\-?\s]\s*)[-+]?[0-9]+(?:\.[0-9]+)?(?:[eE][-+]?[0-9]+)?(?=\s*$)/m,
          lookbehind: true
        },
        'tag': /![^\s]+/,
        'important': /[&*][^\s]+/,
        'punctuation': /---|[[\]{},]/
      };
    }

    // Add basic Protobuf highlighting if not present
    if (!Prism.languages.protobuf) {
      Prism.languages.protobuf = {
        'comment': [
          /\/\/.*$/m,
          /\/\*[\s\S]*?\*\//
        ],
        'string': {
          pattern: /"(?:\\.|[^\\"\r\n])*"/,
          greedy: true
        },
        'keyword': /\b(?:syntax|import|package|option|message|enum|service|rpc|returns|repeated|optional|required|oneof|map|reserved|extensions|extend|group)\b/,
        'builtin': /\b(?:double|float|int32|int64|uint32|uint64|sint32|sint64|fixed32|fixed64|sfixed32|sfixed64|bool|string|bytes)\b/,
        'number': /\b(?:0x[\da-f]+|\d*\.?\d+(?:e[+-]?\d+)?)\b/i,
        'boolean': /\b(?:true|false)\b/,
        'operator': /[=;]/,
        'punctuation': /[{}[\](),]/
      };
    }

    // Register aliases
    if (Prism.languages.yaml) {
      Prism.languages.yml = Prism.languages.yaml;
    }
    if (Prism.languages.protobuf) {
      Prism.languages.proto = Prism.languages.protobuf;
    }

    prismReady = true;
    return true;
  } catch (error) {
    console.warn('Prism sync init failed:', error);
    return false;
  }
};

/**
 * Hook to load Prism languages exactly once
 */
const usePrism = () => {
  const [isReady, setIsReady] = useState(prismReady);

  useEffect(() => {
    if (typeof window === 'undefined') return;

    // Try sync init first
    if (!prismReady && initPrismSync()) {
      setIsReady(true);
      return;
    }

    if (!prismInitPromise) {
      // Initialize Prism
      (window as any).Prism = Prism;
      Prism.manual = true;

      // Load languages asynchronously
      prismInitPromise = Promise.all([
        // @ts-ignore
        import('prismjs/components/prism-protobuf'),
        // @ts-ignore
        import('prismjs/components/prism-yaml'),
        // @ts-ignore
        import('prismjs/components/prism-bash'),
        // @ts-ignore
        import('prismjs/components/prism-go'),
        // @ts-ignore
        import('prismjs/components/prism-json')
      ]).then(() => {
        // Register aliases
        if (Prism.languages.protobuf) {
          Prism.languages.proto = Prism.languages.protobuf;
        }
        if (Prism.languages.bash) {
          Prism.languages.sh = Prism.languages.bash;
          Prism.languages.shell = Prism.languages.bash;
        }
        if (Prism.languages.yaml) {
          Prism.languages.yml = Prism.languages.yaml;
        }

        prismReady = true;
      }).catch(error => {
        console.error('Prism async init failed:', error);
        // Fallback to sync init
        initPrismSync();
        prismReady = true;
      });
    }

    // Wait for initialization
    prismInitPromise.then(() => setIsReady(true));
  }, []);

  return isReady || prismReady;
};

interface CodeBlockComponentProps extends CodeBlockProps {
  children?: string;
  className?: string;
}

export function CodeBlock({
  children = '',
  className = '',
  language,
  hasFocus = false,
  focusLines = []
}: CodeBlockComponentProps) {
  const [isCopied, setIsCopied] = useState(false);
  usePrism();

  // Extract language from className (format: language-xxx)
  let extractedLanguage = language || className.replace(/language-/, '') || 'text';

  // Handle proto/protobuf variations
  if (extractedLanguage === 'proto') {
    extractedLanguage = 'protobuf';
  }

  // Enhanced fallback language detection
  if ((extractedLanguage === 'text' || extractedLanguage === '') && children) {
    const codeContent = children.toLowerCase().trim();
    const originalContent = children.trim();

    // Protobuf detection (flexible patterns)
    if (
      // Explicit protobuf syntax patterns
      /syntax\s*=\s*["']proto[23]?["']/i.test(originalContent) ||
      /import\s+["'][^"']+\.proto["']/i.test(originalContent) ||
      /package\s+[\w.]+;/i.test(originalContent) ||
      // Common protobuf constructs
      (codeContent.includes('message ') && /^\s*\w+\s+\w+\s+=\s+\d+;/m.test(originalContent)) ||
      (codeContent.includes('enum ') && /^\s*\w+\s*=\s*\d+;/m.test(originalContent)) ||
      (codeContent.includes('service ') && codeContent.includes('rpc ')) ||
      // Proto3 specific patterns
      /^\s*syntax\s*=.*proto/mi.test(originalContent)
    ) {
      extractedLanguage = 'protobuf';
    }
    // YAML detection (enhanced patterns)
    else if (
      // YAML keys
      codeContent.includes('version:') ||
      codeContent.includes('lint:') ||
      codeContent.includes('use:') ||
      codeContent.includes('generate:') ||
      codeContent.includes('deps:') ||
      codeContent.includes('plugins:') ||
      codeContent.includes('breaking:') ||
      codeContent.includes('ignore:') ||
      codeContent.includes('except:') ||
      codeContent.includes('ignore_only:') ||
      // Common YAML patterns
      /^\s*[\w-]+:\s*/m.test(originalContent) ||
      /^\s*-\s+/.test(originalContent) ||
      // YAML configuration suffixes
      codeContent.includes('_suffix:') ||
      codeContent.includes('_prefix:') ||
      // YAML boolean values
      /:\s*(true|false|yes|no)\s*$/m.test(originalContent)
    ) {
      extractedLanguage = 'yaml';
    }
    // Bash/Shell detection
    else if (
      codeContent.includes('easyp ') ||
      codeContent.startsWith('npm ') ||
      codeContent.startsWith('go ') ||
      codeContent.startsWith('$ ') ||
      codeContent.startsWith('brew ') ||
      codeContent.includes('#!/bin/') ||
      /^[\$#]\s+/.test(originalContent) ||
      codeContent.includes('docker ') ||
      codeContent.includes('git ')
    ) {
      extractedLanguage = 'bash';
    }
    // JSON detection
    else if (
      (codeContent.startsWith('{') && codeContent.endsWith('}')) ||
      (codeContent.startsWith('[') && codeContent.endsWith(']')) ||
      // JSON-like structure
      /^\s*["{[]/.test(originalContent)
    ) {
      extractedLanguage = 'json';
    }
    // Configuration detection (fallback to YAML for config-like content)
    else if (
      codeContent.includes('=') && codeContent.includes(':') ||
      /^\w+\s*[:=]\s*.+$/m.test(originalContent)
    ) {
      extractedLanguage = 'yaml';
    }
  }


  // Map language names for better compatibility with prism-react-renderer
  const languageMap: { [key: string]: string } = {
    'protobuf': 'proto',
    'yaml': 'yaml',
    'yml': 'yaml',
    'bash': 'bash',
    'shell': 'bash',
    'json': 'json'
  };

  // Force protobuf language to be available
  if (extractedLanguage === 'protobuf' && (!Prism.languages.protobuf || !Prism.languages.proto)) {
    // Ensure protobuf is available even if async loading failed
    initPrismSync();
  }

  // Use mapped language or fallback to original
  const finalLanguage = languageMap[extractedLanguage] || extractedLanguage;



  // Clean the code content
  const code = children.trim();

  // Parse focus lines from code if not provided
  const { cleanCode, parsedFocusLines } = React.useMemo(() => {
    if (focusLines.length > 0) {
      return { cleanCode: code, parsedFocusLines: focusLines };
    }
    return parseCodeFocus(code);
  }, [code, focusLines]);

  const shouldShowFocus = hasFocus || parsedFocusLines.length > 0;

  const handleCopy = () => {
    navigator.clipboard.writeText(cleanCode);
    setIsCopied(true);
    setTimeout(() => setIsCopied(false), 2000);
  };

  // Use standard Prism highlighting for all languages including protobuf
  return (
    <div className={`code-block group ${shouldShowFocus ? 'code-block--with-focus' : ''} ${className}`}>
      {/* Window Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-white/5 bg-slate-900/50">
        {/* Window Controls (Traffic Lights) */}
        <div className="flex items-center gap-2">
          <div className="w-3 h-3 rounded-full bg-slate-700/50 group-hover:bg-red-500/80 transition-colors duration-300" />
          <div className="w-3 h-3 rounded-full bg-slate-700/50 group-hover:bg-yellow-500/80 transition-colors duration-300" />
          <div className="w-3 h-3 rounded-full bg-slate-700/50 group-hover:bg-green-500/80 transition-colors duration-300" />
        </div>

        {/* Title / Language */}
        <div className="absolute left-1/2 -translate-x-1/2 flex items-center gap-2 text-xs font-medium text-slate-500 font-mono">
          <Terminal size={12} />
          <span className="uppercase tracking-wider">
            {finalLanguage === 'proto' || finalLanguage === 'protobuf' ? 'PROTOBUF' :
             finalLanguage === 'yaml' ? 'YAML' :
             finalLanguage === 'bash' ? 'BASH' :
             finalLanguage === 'json' ? 'JSON' :
             finalLanguage === 'go' ? 'GO' :
             finalLanguage.toUpperCase()}
          </span>
        </div>

        {/* Copy Button */}
        <button
          onClick={handleCopy}
          className="flex items-center gap-2 px-3 py-1.5 text-xs text-slate-400 hover:text-white bg-slate-800/50 hover:bg-slate-700/50 border border-slate-600/30 hover:border-slate-500/50 rounded-md transition-all duration-200"
        >
          {isCopied ? <Check size={12} /> : <Copy size={12} />}
          {isCopied ? 'Copied' : 'Copy'}
        </button>
      </div>

      {/* Code Content */}
      <Highlight
        theme={cyberTheme}
        code={cleanCode}
        language={finalLanguage as any}
        prism={Prism}
      >
        {({ className: highlightClassName, style, tokens, getLineProps, getTokenProps }) => (
          <pre
            className={`${highlightClassName} code-pre`}
            style={{ ...style, background: 'transparent' }}
          >
            {tokens.map((line, lineIndex) => {
              const lineNumber = lineIndex + 1;
              const isFocusLine = parsedFocusLines.includes(lineNumber);
              const isDimmedLine = shouldShowFocus && !isFocusLine;

              const lineProps = getLineProps({ line, key: lineIndex });

              return (
                <div
                  key={lineIndex}
                  className={`
                    code-line
                    ${isFocusLine ? 'line-focus' : ''}
                    ${isDimmedLine ? 'line-dimmed' : ''}
                  `.trim()}
                  data-line={lineNumber}
                >
                  <span className="line-number">{lineNumber}</span>
                  <span className="line-content">
                    {line.map((token, tokenIndex) => (
                      <span key={tokenIndex} {...getTokenProps({ token })} />
                    ))}
                  </span>
                </div>
              );
            })}
          </pre>
        )}
      </Highlight>
    </div>
  );

}

/**
 * Inline Code Component with Syntax Highlighting
 */
function InlineCode({ children }: { children: string }) {
  usePrism();

  // Heuristic to guess language
  let lang = 'text';
  if (children.includes('easyp') || children.startsWith('npm') || children.startsWith('go ')) {
    lang = 'bash';
  } else if (children.includes('message ') || children.includes('service ') || children.includes('rpc ') || /^[A-Z][a-zA-Z0-9_]*$/.test(children)) {
    lang = 'protobuf';
  }

  return (
    <Highlight
      theme={cyberTheme}
      code={children}
      language={lang as any}
      prism={Prism}
    >
      {({ className, style, tokens, getTokenProps }) => (
        <code className={`inline-code ${className}`} style={{ ...style, backgroundColor: 'rgba(30, 41, 59, 0.5)' }}>
          {tokens.map((line, i) => (
            <span key={i}>
              {line.map((token, key) => (
                <span key={key} {...getTokenProps({ token })} />
              ))}
            </span>
          ))}
        </code>
      )}
    </Highlight>
  );
}

/**
 * Parses code content to extract focus lines and clean code
 * @param code Raw code content
 * @returns Clean code and focus line numbers
 */
function parseCodeFocus(code: string): {
  cleanCode: string;
  parsedFocusLines: number[];
} {
  const lines = code.split('\n');
  const focusLines: number[] = [];
  const cleanLines: string[] = [];

  lines.forEach((line, index) => {
    const lineNumber = index + 1;

    // Check for focus comment patterns
    const focusPatterns = [
      /\/\/\s*\[!code\s+focus\]/i,
      /\/\*\s*\[!code\s+focus\]\s*\*\//i,
      /#\s*\[!code\s+focus\]/i,        // Python, shell comments
      /<!--\s*\[!code\s+focus\]\s*-->/i, // HTML comments
      /\/\/.*\[!code\s+focus\]/i,      // More flexible JS comments
    ];

    let hasFocusComment = false;
    let cleanLine = line;

    for (const pattern of focusPatterns) {
      if (pattern.test(line)) {
        hasFocusComment = true;
        // Remove the focus comment from the line
        cleanLine = line.replace(pattern, '').trim();
        break;
      }
    }

    if (hasFocusComment) {
      focusLines.push(lineNumber);
    }

    cleanLines.push(cleanLine);
  });

  return {
    cleanCode: cleanLines.join('\n'),
    parsedFocusLines: focusLines
  };
}

/**
 * Pre-processes code blocks to identify focus annotations
 * @param content Markdown content
 * @returns Content with focus data extracted
 */
export function preprocessCodeFocus(content: string): string {
  // This will be used to identify code blocks with focus before rendering
  return content;
}

/**
 * Custom code component for markdown-to-jsx
 */
export function MarkdownCodeBlock(props: any) {
  const { children, className, ...rest } = props;



  // Handle inline code
  if (!className) {
    return <InlineCode>{children}</InlineCode>;
  }

  return <CodeBlock className={className} {...rest}>{children}</CodeBlock>;
}

/**
 * Custom pre component for markdown-to-jsx
 */
export function MarkdownPre(props: any) {
  const { children, ...rest } = props;



  // If children is a code element (or our custom MarkdownCodeBlock), render as CodeBlock directly
  // This replaces the <pre> tag entirely with our CodeBlock component
  if (React.isValidElement(children)) {
    // Check if it's a standard 'code' tag or our custom MarkdownCodeBlock component
    // We check for MarkdownCodeBlock by reference since they are in the same file
    if (children.type === 'code' || children.type === MarkdownCodeBlock) {
      const codeProps = children.props as any;

      // IMPORTANT: We return CodeBlock directly. CodeBlock renders a <div> wrapper.
      // If we wrapped this in a <pre>, we would get <pre><div>...</div></pre> which is invalid HTML.
      // By returning CodeBlock directly, we replace the markdown <pre> with our <div class="code-block">.
      return (
        <CodeBlock
          className={codeProps.className}
          {...rest}
        >
          {codeProps.children}
        </CodeBlock>
      );
    }
  }

  // Fallback for other pre content
  return <pre {...rest}>{children}</pre>;
}

export default CodeBlock;
