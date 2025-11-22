export interface MonitorResult {
  name: string
  url: string
  up: boolean
  status: number
  timestamp: string
}

export interface MonitorData {
  generated_at: string
  results: MonitorResult[]
  history: Record<string, { up: boolean; timestamp: string; responseTime?: number }[]>
}

const SERVICES = [
  { name: "API Gateway", url: "https://api.example.com/health" },
  { name: "Auth Service", url: "https://auth.example.com/status" },
  { name: "Database Primary", url: "postgres://db-prod-1" },
  { name: "Redis Cache", url: "redis://cache-cluster" },
  { name: "Frontend App", url: "https://app.example.com" },
  { name: "Marketing Site", url: "https://example.com" },
  { name: "Payment Processor", url: "https://pay.example.com/ping" },
  { name: "Email Worker", url: "worker://email-queue" },
  { name: "Search Index", url: "https://search.example.com/_cat/health" },
  { name: "CDN Edge", url: "https://cdn.example.com/status" },
]

export function generateMockData(): MonitorData {
  const now = new Date()
  const results: MonitorResult[] = SERVICES.map((service) => {
    // Random status, mostly up
    const isUp = Math.random() > 0.05
    return {
      name: service.name,
      url: service.url,
      up: isUp,
      status: isUp ? 200 : 503,
      timestamp: now.toISOString(),
    }
  })

  // Generate mock history for sparklines
  const history: Record<string, { up: boolean; timestamp: string; responseTime?: number }[]> = {}

  SERVICES.forEach((service) => {
    const points = []
    for (let i = 0; i < 60; i++) {
      // Mostly up history
      const isUp = Math.random() > 0.02
      points.push({
        up: isUp,
        timestamp: new Date(now.getTime() - (60 - i) * 60000).toISOString(),
        // Mock response time between 50ms and 200ms
        responseTime: isUp ? 50 + Math.random() * 150 : 0,
      })
    }
    history[service.name] = points
  })

  return {
    generated_at: now.toLocaleString(),
    results,
    history,
  }
}
