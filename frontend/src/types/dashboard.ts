export type ActivitySource = 'garmin' | 'coros' | 'fit_upload' | 'manual'

export type ActivityStatus = 'uploaded' | 'analyzing' | 'ready' | 'failed'

export type ActivityType = 'easy' | 'tempo' | 'intervals' | 'long_run' | 'recovery' | 'race'

export type ReprocessReason =
  | 'outdated_analysis'
  | 'stale_pipeline'
  | 'pipeline_in_progress'
  | 'already_current'
  | 'outside_window'
  | 'not_ready'
  | 'missing_fit'

export type Activity = {
  id: string
  title: string
  type: ActivityType
  source: ActivitySource
  status: ActivityStatus
  startedAt: string
  durationSec: number
  distanceM: number
  avgPaceSecPerKm: number | null
  avgHeartRate: number | null
  elevationGainM: number | null
  summary: string | null
  metricsObjectKey?: string | null
  featuresObjectKey?: string | null
  analysisVersion?: number | null
  reprocessAvailable?: boolean
  reprocessReason?: ReprocessReason | null
}

export type WeekSummary = {
  weekLabel: string
  totalDistanceM: number
  totalDurationSec: number
  activitiesCount: number
  completedWorkouts: number
  plannedWorkouts: number
  loadScore: number
  readinessLabel: string
  highlight: string
}

export type PlanDay = {
  date: string
  title: string
  type: ActivityType
  completed: boolean
  adapted: boolean
}

export type DeviceConnection = {
  id: string
  provider: 'garmin' | 'coros' | 'strava'
  label: string
  connected: boolean
  lastSyncAt: string | null
}

export type ChatMessage = {
  id: string
  role: 'athlete' | 'coach'
  content: string
  createdAt: string
}

export type DashboardData = {
  week: WeekSummary
  activities: Activity[]
  planDays: PlanDay[]
  devices: DeviceConnection[]
}
