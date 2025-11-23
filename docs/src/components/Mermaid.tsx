import React, { useEffect, useRef, useState } from 'react';
import mermaid from 'mermaid';

interface MermaidProps {
  chart: string;
  id?: string;
}

const Mermaid: React.FC<MermaidProps> = ({ chart, id }) => {
  const ref = useRef<HTMLDivElement>(null);
  const [isInitialized, setIsInitialized] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!isInitialized) {
      mermaid.initialize({
        startOnLoad: false,
        theme: 'base',
        securityLevel: 'loose',
        themeVariables: {
          // Background colors
          background: '#020617',           // Slate 950 - main background
          primaryColor: '#3b82f6',         // Blue 500 - primary
          primaryTextColor: '#ffffff',     // White text
          primaryBorderColor: '#60a5fa',   // Blue 400 - primary glow

          // Secondary colors
          secondaryColor: '#8b5cf6',       // Violet 500 - secondary
          secondaryBorderColor: '#a78bfa', // Violet 400

          // Tertiary colors
          tertiaryColor: '#1e293b',        // Slate 800 - surface highlight
          tertiaryTextColor: '#e2e8f0',    // Slate 200
          tertiaryBorderColor: '#334155',  // Slate 700 - border

          // Text colors
          textColor: '#e2e8f0',           // Slate 200
          secondaryTextColor: '#cbd5e1',   // Slate 300

          // Background variations
          primaryColorLight: '#60a5fa',    // Blue 400
          primaryColorDark: '#1d4ed8',     // Blue 700

          // Line colors
          lineColor: '#334155',            // Slate 700 - border
          edgeLabelBackground: '#0f172a',  // Slate 900 - surface

          // Node colors
          mainBkg: '#1e293b',              // Slate 800 - surface highlight
          nodeBkg: '#1e293b',
          nodeTextColor: '#e2e8f0',
          nodeBorder: '#60a5fa',           // Blue 400 - primary glow

          // Cluster colors
          clusterBkg: '#0f172a',           // Slate 900 - surface
          clusterBorder: '#334155',        // Slate 700 - border

          // Active states
          activeTaskBkgColor: '#3b82f6',   // Blue 500 - primary
          activeTaskBorderColor: '#60a5fa', // Blue 400 - primary glow

          // Grid and axis
          gridColor: '#334155',            // Slate 700 - border
          section0: '#1e293b',             // Slate 800
          section1: '#0f172a',             // Slate 900
          section2: '#020617',             // Slate 950
          section3: '#1e293b',             // Slate 800

          // Special elements
          fillType0: '#3b82f6',            // Blue 500 - primary
          fillType1: '#8b5cf6',            // Violet 500 - secondary
          fillType2: '#06b6d4',            // Cyan 500
          fillType3: '#10b981',            // Emerald 500
          fillType4: '#f59e0b',            // Amber 500
          fillType5: '#ef4444',            // Red 500
          fillType6: '#ec4899',            // Pink 500
          fillType7: '#8b5cf6',            // Violet 500

          // Actor colors for sequence diagrams
          actorBkg: '#1e293b',             // Slate 800
          actorBorder: '#60a5fa',          // Blue 400
          actorTextColor: '#e2e8f0',       // Slate 200
          actorLineColor: '#334155',       // Slate 700

          // Sequence diagram colors
          activationBkgColor: '#3b82f6',   // Blue 500
          activationBorderColor: '#60a5fa', // Blue 400
          signalColor: '#e2e8f0',          // Slate 200
          signalTextColor: '#e2e8f0',      // Slate 200

          // Note colors
          noteBkgColor: '#0f172a',         // Slate 900
          noteBorderColor: '#8b5cf6',      // Violet 500
          noteTextColor: '#e2e8f0',        // Slate 200

          // Loop colors
          loopTextColor: '#e2e8f0',        // Slate 200

          // Label colors
          labelColor: '#e2e8f0',           // Slate 200
          labelTextColor: '#020617',       // Slate 950 (for contrast on light backgrounds)
          labelBoxBkgColor: '#3b82f6',     // Blue 500
          labelBoxBorderColor: '#60a5fa',  // Blue 400

          // Error colors
          errorBkgColor: '#ef4444',        // Red 500
          errorTextColor: '#ffffff',

          // Task colors for Gantt charts
          cScale0: '#3b82f6',              // Blue 500
          cScale1: '#8b5cf6',              // Violet 500
          cScale2: '#06b6d4',              // Cyan 500
          cScale3: '#10b981',              // Emerald 500
          cScale4: '#f59e0b',              // Amber 500
          cScale5: '#ef4444',              // Red 500

          // Pie chart colors
          pie1: '#3b82f6',                 // Blue 500
          pie2: '#8b5cf6',                 // Violet 500
          pie3: '#06b6d4',                 // Cyan 500
          pie4: '#10b981',                 // Emerald 500
          pie5: '#f59e0b',                 // Amber 500
          pie6: '#ef4444',                 // Red 500
          pie7: '#ec4899',                 // Pink 500
          pie8: '#84cc16',                 // Lime 500
          pie9: '#6366f1',                 // Indigo 500
          pie10: '#14b8a6',                // Teal 500
          pie11: '#f97316',                // Orange 500
          pie12: '#a855f7',                // Purple 500
          pieTitleTextSize: '18px',
          pieTitleTextColor: '#e2e8f0',    // Slate 200
          pieSectionTextSize: '14px',
          pieSectionTextColor: '#ffffff',
          pieLegendTextSize: '14px',
          pieLegendTextColor: '#e2e8f0',   // Slate 200
          pieStrokeColor: '#020617',       // Slate 950
          pieStrokeWidth: '1px',
          pieOuterStrokeWidth: '2px',
          pieOuterStrokeColor: '#334155',  // Slate 700
          pieOpacity: '0.9',

          // Git graph colors
          git0: '#3b82f6',                 // Blue 500
          git1: '#8b5cf6',                 // Violet 500
          git2: '#06b6d4',                 // Cyan 500
          git3: '#10b981',                 // Emerald 500
          git4: '#f59e0b',                 // Amber 500
          git5: '#ef4444',                 // Red 500
          git6: '#ec4899',                 // Pink 500
          git7: '#84cc16',                 // Lime 500
          gitBranchLabel0: '#e2e8f0',      // Slate 200
          gitBranchLabel1: '#e2e8f0',
          gitBranchLabel2: '#e2e8f0',
          gitBranchLabel3: '#e2e8f0',
          gitBranchLabel4: '#e2e8f0',
          gitBranchLabel5: '#e2e8f0',
          gitBranchLabel6: '#e2e8f0',
          gitBranchLabel7: '#e2e8f0',
        },
        flowchart: {
          curve: 'cardinal',
          padding: 20,
          nodeSpacing: 50,
          rankSpacing: 50,
          diagramPadding: 8,
          htmlLabels: true,
        },
        sequence: {
          diagramMarginX: 50,
          diagramMarginY: 10,
          actorMargin: 50,
          width: 150,
          height: 65,
          boxMargin: 10,
          boxTextMargin: 5,
          noteMargin: 10,
          messageMargin: 35,
          mirrorActors: true,
          bottomMarginAdj: 1,
          useMaxWidth: true,
          rightAngles: false,
          showSequenceNumbers: false,
        },
        gantt: {
          titleTopMargin: 25,
          barHeight: 20,
          fontsize: 11,
          sectionFontSize: 11,
          gridLineStartPadding: 35,
          bottomPadding: 25,
          leftPadding: 75,
          topPadding: 50,
          rightPadding: 75,
        },
        journey: {
          diagramMarginX: 50,
          diagramMarginY: 10,
          leftMargin: 150,
          width: 150,
          height: 50,
          boxMargin: 10,
          boxTextMargin: 5,
          noteMargin: 10,
          messageMargin: 35,
          bottomMarginAdj: 1,
        },
        pie: {
          textPosition: 0.75,
        },
        requirement: {
          rect_fill: '#1e293b',            // Slate 800
          text_color: '#e2e8f0',           // Slate 200
          rect_border_size: '1px',
          rect_border_color: '#334155',    // Slate 700
          rect_min_width: 200,
          rect_min_height: 200,
          fontSize: 14,
          fontWeight: 'normal',
        },
        gitGraph: {
          diagramPadding: 8,
          nodeLabel: {
            width: 75,
            height: 100,
            x: -25,
            y: 0,
          },
        },
      });
      setIsInitialized(true);
    }
  }, [isInitialized]);

  useEffect(() => {
    if (!isInitialized || !ref.current) return;

    const renderDiagram = async () => {
      try {
        setError(null);
        const element = ref.current;
        if (!element) return;

        // Clear previous content
        element.innerHTML = '';

        // Generate unique ID for the diagram
        const diagramId = id || `mermaid-${Math.random().toString(36).substr(2, 9)}`;

        // Render the diagram
        const { svg, bindFunctions } = await mermaid.render(diagramId, chart);

        element.innerHTML = svg;

        // Bind any interactive functions if they exist
        if (bindFunctions) {
          bindFunctions(element);
        }
      } catch (err) {
        console.error('Mermaid rendering error:', err);
        setError(err instanceof Error ? err.message : 'Failed to render diagram');
      }
    };

    renderDiagram();
  }, [chart, id, isInitialized]);

  if (error) {
    return (
      <div className="border border-red-500/30 bg-red-500/10 p-4 rounded-lg backdrop-blur-sm">
        <div className="text-red-300 text-sm font-medium">
          Error rendering diagram
        </div>
        <div className="text-red-400 text-sm mt-1">
          {error}
        </div>
        <details className="mt-2">
          <summary className="text-red-300 text-xs cursor-pointer hover:text-red-200">
            Show diagram code
          </summary>
          <pre className="text-red-400 text-xs mt-1 whitespace-pre-wrap font-mono bg-slate-900/50 p-2 rounded border border-red-500/20 overflow-x-auto">
            {chart}
          </pre>
        </details>
      </div>
    );
  }

  return (
    <div className="mermaid-container my-8">
      <div
        ref={ref}
        className="flex justify-center items-center min-h-[200px] bg-slate-950/50 rounded-xl border border-slate-800/50 p-6 backdrop-blur-sm"
        style={{
          filter: 'drop-shadow(0 4px 20px rgba(59, 130, 246, 0.1))',
        }}
      />
    </div>
  );
};

export default Mermaid;
