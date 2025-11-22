interface SparklineProps {
  data: { up: boolean; timestamp: string; responseTime?: number }[]
  color?: string
  width?: number
  height?: number
}

export function Sparkline({ data, color = "#10b981", width = 160, height = 40 }: SparklineProps) {
  // If no data, render empty state
  if (!data || data.length === 0) {
    return (
      <div className="h-full w-full bg-muted/20 flex items-center justify-center text-[10px] text-muted-foreground">
        No Data
      </div>
    )
  }

  // Normalize data to fit in the SVG
  const values = data.map((d) => (d.up ? d.responseTime || 100 : 0))
  const max = Math.max(...values, 1)
  const min = 0 // Always start from 0 for area charts

  // Create points for the line
  const points = values
    .map((val, i) => {
      const x = (i / (values.length - 1)) * width
      // Invert Y because SVG coordinates start from top
      const y = height - ((val - min) / (max - min)) * height
      return `${x},${y}`
    })
    .join(" ")

  // Create points for the filled area (close the loop at the bottom)
  const areaPoints = `0,${height} ${points} ${width},${height}`

  // Unique ID for the gradient to avoid conflicts if multiple charts are on page
  const gradientId = `gradient-${Math.random().toString(36).substr(2, 9)}`

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

      {/* Area fill */}
      <polygon points={areaPoints} fill={`url(#${gradientId})`} />

      {/* Line */}
      <polyline
        points={points}
        fill="none"
        stroke={color}
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />

      {/* Dot at the end */}
      {values.length > 0 && (
        <circle
          cx={width}
          cy={height - ((values[values.length - 1] - min) / (max - min)) * height}
          r="2"
          fill={color}
        />
      )}
    </svg>
  )
}
