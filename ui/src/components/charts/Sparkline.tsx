/**
 * Sparkline component for rendering simple line charts
 * Uses SVG with smooth bezier curves
 */
interface SparklineProps {
  /** Array of numeric values to plot */
  data: number[];
  /** Stroke color (CSS color string) */
  color?: string;
  /** Height in pixels */
  height?: number;
  /** Width in pixels */
  width?: number;
}

/**
 * Generates a smooth SVG path using catmull-rom spline interpolation
 * Converts to cubic bezier curves for smooth lines
 */
function generateSmoothPath(points: { x: number; y: number }[]): string {
  if (points.length < 2) return '';

  if (points.length === 2) {
    return `M ${points[0].x} ${points[0].y} L ${points[1].x} ${points[1].y}`;
  }

  let path = `M ${points[0].x} ${points[0].y}`;

  for (let i = 0; i < points.length - 1; i++) {
    const p0 = points[i === 0 ? i : i - 1];
    const p1 = points[i];
    const p2 = points[i + 1];
    const p3 = points[i + 2 < points.length ? i + 2 : i + 1];

    const cp1x = p1.x + (p2.x - p0.x) / 6;
    const cp1y = p1.y + (p2.y - p0.y) / 6;
    const cp2x = p2.x - (p3.x - p1.x) / 6;
    const cp2y = p2.y - (p3.y - p1.y) / 6;

    path += ` C ${cp1x} ${cp1y}, ${cp2x} ${cp2y}, ${p2.x} ${p2.y}`;
  }

  return path;
}

export default function Sparkline({
  data,
  color = 'currentColor',
  height = 24,
  width = 256
}: SparklineProps) {
  if (!data || data.length < 2) return null;

  const padding = 4;
  const effectiveWidth = width - padding * 2;
  const effectiveHeight = height - padding * 2;

  const max = Math.max(...data);
  const min = 0;
  const range = max - min || 1;

  const points = data.map((value, index) => ({
    x: padding + (index / (data.length - 1)) * effectiveWidth,
    y: padding + effectiveHeight - ((value - min) / range) * effectiveHeight
  }));

  const path = generateSmoothPath(points);

  return (
    <svg
      width={width}
      height={height}
      viewBox={`0 0 ${width} ${height}`}
      preserveAspectRatio="xMidYMid meet"
    >
      <path
        d={path}
        fill="none"
        stroke={color}
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}
