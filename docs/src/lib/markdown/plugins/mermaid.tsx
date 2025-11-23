import React from 'react';
import type { MarkdownProcessorOptions } from '../types';
import Mermaid from '../../../components/Mermaid';

/**
 * Processes Mermaid code blocks in markdown content
 * Converts ```mermaid blocks to Mermaid React components
 *
 * @param content The markdown content to process
 * @param options Processing options
 * @returns Processed content with Mermaid placeholders
 */
export function processMermaid(
  content: string,
  options?: MarkdownProcessorOptions
): string {
  // Find all mermaid code blocks
  const mermaidRegex = /```mermaid\n([\s\S]*?)\n```/g;
  let processedContent = content;
  let match;
  let diagramIndex = 0;

  while ((match = mermaidRegex.exec(content)) !== null) {
    const [fullMatch, diagramCode] = match;
    const diagramId = `mermaid-diagram-${diagramIndex++}`;

    // Clean up the diagram code
    const cleanDiagramCode = diagramCode.trim();

    // Create placeholder that will be replaced during rendering
    const placeholder = `<MERMAID_DIAGRAM_${diagramId}>${cleanDiagramCode}</MERMAID_DIAGRAM_${diagramId}>`;

    processedContent = processedContent.replace(fullMatch, placeholder);
  }

  return processedContent;
}

/**
 * Creates a component override for rendering Mermaid diagrams in markdown-to-jsx
 *
 * @returns Component overrides object for markdown-to-jsx
 */
export function createMermaidOverrides() {
  return {
    // Handle custom Mermaid placeholder elements
    MERMAID_DIAGRAM_0: ({ children }: { children: string }) => (
      <Mermaid chart={children} id="mermaid-diagram-0" />
    ),
    MERMAID_DIAGRAM_1: ({ children }: { children: string }) => (
      <Mermaid chart={children} id="mermaid-diagram-1" />
    ),
    MERMAID_DIAGRAM_2: ({ children }: { children: string }) => (
      <Mermaid chart={children} id="mermaid-diagram-2" />
    ),
    MERMAID_DIAGRAM_3: ({ children }: { children: string }) => (
      <Mermaid chart={children} id="mermaid-diagram-3" />
    ),
    MERMAID_DIAGRAM_4: ({ children }: { children: string }) => (
      <Mermaid chart={children} id="mermaid-diagram-4" />
    ),
    // Add more as needed or use dynamic generation
  };
}

/**
 * Creates dynamic Mermaid overrides for any number of diagrams
 *
 * @param content Processed markdown content
 * @returns Dynamic component overrides
 */
export function createDynamicMermaidOverrides(content: string) {
  const overrides: Record<string, React.ComponentType<{ children: string }>> = {};

  // Find all mermaid placeholders
  const placeholderRegex = /<MERMAID_DIAGRAM_(mermaid-diagram-\d+)>([\s\S]*?)<\/MERMAID_DIAGRAM_\1>/g;
  let match;

  while ((match = placeholderRegex.exec(content)) !== null) {
    const [, diagramId, diagramCode] = match;
    const componentName = `MERMAID_DIAGRAM_${diagramId.replace(/-/g, '_')}`;

    overrides[componentName] = ({ children }: { children: string }) => (
      <Mermaid chart={children || diagramCode} id={diagramId} />
    );
  }

  return overrides;
}

/**
 * Alternative approach: Replace mermaid blocks with HTML that can be processed by markdown-to-jsx
 *
 * @param content The markdown content to process
 * @returns Content with mermaid blocks replaced by custom HTML elements
 */
export function processMermaidAsHtml(content: string): string {
  // More flexible regex that handles different line endings and whitespace
  const mermaidRegex = /```mermaid\r?\n([\s\S]*?)\r?\n```/g;
  let processedContent = content;
  let diagramIndex = 0;



  processedContent = processedContent.replace(mermaidRegex, (match, diagramCode) => {
    const diagramId = `mermaid-diagram-${diagramIndex++}`;
    const cleanDiagramCode = diagramCode.trim();



    // Create a div element that can be overridden by markdown-to-jsx
    return `<div class="mermaid-diagram" data-id="${diagramId}" data-chart="${encodeURIComponent(cleanDiagramCode)}"></div>`;
  });

  return processedContent;
}

/**
 * Creates component overrides for the HTML approach
 */
export function createMermaidHtmlOverrides() {
  return {
    div: (props: any) => {
      // Check if this is a mermaid diagram div
      if (props.className === 'mermaid-diagram' && props['data-chart']) {
        const chart = props['data-chart'] ? decodeURIComponent(props['data-chart']) : '';
        return <Mermaid chart={chart} id={props['data-id']} />;
      }
      // Return normal div for non-mermaid divs
      return <div {...props} />;
    }
  };
}

/**
 * Validates Mermaid syntax
 *
 * @param diagramCode Mermaid diagram code
 * @returns Validation result
 */
export function validateMermaidSyntax(diagramCode: string): {
  isValid: boolean;
  errors: string[];
  warnings: string[];
} {
  const errors: string[] = [];
  const warnings: string[] = [];

  try {
    // Basic syntax checks
    const trimmedCode = diagramCode.trim();

    if (!trimmedCode) {
      errors.push('Empty diagram code');
      return { isValid: false, errors, warnings };
    }

    // Check for common diagram types
    const supportedTypes = [
      'graph', 'flowchart', 'sequenceDiagram', 'classDiagram',
      'stateDiagram', 'gantt', 'pie', 'gitgraph', 'erDiagram',
      'journey', 'requirementDiagram'
    ];

    const firstLine = trimmedCode.split('\n')[0].toLowerCase();
    const hasKnownType = supportedTypes.some(type =>
      firstLine.includes(type.toLowerCase())
    );

    if (!hasKnownType) {
      warnings.push('Diagram type not recognized, ensure it starts with a valid Mermaid diagram declaration');
    }

    // Check for balanced brackets/parentheses
    const brackets = { '(': 0, '[': 0, '{': 0 };
    for (const char of trimmedCode) {
      if (char === '(') brackets['(']++;
      if (char === ')') brackets['(']--;
      if (char === '[') brackets['[']++;
      if (char === ']') brackets['[']--;
      if (char === '{') brackets['{']++;
      if (char === '}') brackets['{']--;
    }

    Object.entries(brackets).forEach(([bracket, count]) => {
      if (count !== 0) {
        errors.push(`Unbalanced ${bracket} brackets`);
      }
    });

  } catch (error) {
    errors.push(`Syntax validation error: ${(error as Error).message}`);
  }

  return {
    isValid: errors.length === 0,
    errors,
    warnings
  };
}

/**
 * Extracts all Mermaid diagrams from content
 *
 * @param content Markdown content
 * @returns Array of diagram objects
 */
export function extractMermaidDiagrams(content: string): Array<{
  id: string;
  code: string;
  type: string;
  validation: ReturnType<typeof validateMermaidSyntax>;
}> {
  const diagrams: Array<{
    id: string;
    code: string;
    type: string;
    validation: ReturnType<typeof validateMermaidSyntax>;
  }> = [];

  const mermaidRegex = /```mermaid\n([\s\S]*?)\n```/g;
  let match;
  let index = 0;

  while ((match = mermaidRegex.exec(content)) !== null) {
    const [, diagramCode] = match;
    const cleanCode = diagramCode.trim();
    const firstLine = cleanCode.split('\n')[0].toLowerCase();

    // Determine diagram type
    let type = 'unknown';
    if (firstLine.includes('graph') || firstLine.includes('flowchart')) type = 'flowchart';
    else if (firstLine.includes('sequencediagram')) type = 'sequence';
    else if (firstLine.includes('classdiagram')) type = 'class';
    else if (firstLine.includes('statediagram')) type = 'state';
    else if (firstLine.includes('gantt')) type = 'gantt';
    else if (firstLine.includes('pie')) type = 'pie';
    else if (firstLine.includes('gitgraph')) type = 'gitgraph';
    else if (firstLine.includes('erdiagram')) type = 'er';
    else if (firstLine.includes('journey')) type = 'journey';
    else if (firstLine.includes('requirementdiagram')) type = 'requirement';

    diagrams.push({
      id: `mermaid-diagram-${index++}`,
      code: cleanCode,
      type,
      validation: validateMermaidSyntax(cleanCode)
    });
  }

  return diagrams;
}

export default {
  processMermaid,
  processMermaidAsHtml,
  createMermaidOverrides,
  createDynamicMermaidOverrides,
  createMermaidHtmlOverrides,
  validateMermaidSyntax,
  extractMermaidDiagrams
};
