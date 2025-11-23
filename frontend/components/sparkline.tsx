interface SparklineProps {
  data: { up: boolean; timestamp: string; responseTime?: number }[];
  color?: string;
  width?: number;
  height?: number;
}

export function Sparkline({
  data,
  color = "#10b981",
  width = 160,
  height = 40,
}: SparklineProps) {
  if (!data || data.length === 0) {
    return (
      <div className="h-full w-full bg-muted/20 flex items-center justify-center text-[10px] text-muted-foreground">
        No Data
      </div>
    );
  }

  const values = data.map((d) => (d.up ? d.responseTime || 100 : 0));
  const max = Math.max(...values, 1);
  const min = 0;

  const points = values
    .map((val, i) => {
      const x = (i / (values.length - 1)) * width;
      const y = height - ((val - min) / (max - min)) * height;
      return `${x},${y}`;
    })
    .join(" ");

  const areaPoints = `0,${height} ${points} ${width},${height}`;

  const gradientId = `gradient-${Math.random().toString(36).substr(2, 9)}`;

  return (
    <svg
      width="100%"
      height="100%"
      viewBox={`0 0 ${width} ${height}`}
      preserveAspectRatio="none"
      className="overflow-visible"
    >
      <defs>
        <linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor={color} stopOpacity={0.2} />
          <stop offset="100%" stopColor={color} stopOpacity={0} />
        </linearGradient>
      </defs>

      <polygon points={areaPoints} fill={`url(#${gradientId})`} />

      <polyline
        points={points}
        fill="none"
        stroke={color}
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />

      {values.length > 0 && (
        <circle
          cx={width}
          cy={
            height - ((values[values.length - 1] - min) / (max - min)) * height
          }
          r="2"
          fill={color}
        />
      )}
    </svg>
  );
}
