import type {
  Activity,
  DashboardData,
  DeviceConnection,
  PlanDay,
  WeekSummary,
} from '@/types/dashboard'

const now = new Date()

function daysAgo(days: number, hour = 7, minute = 10): string {
  const date = new Date(now)
  date.setDate(date.getDate() - days)
  date.setHours(hour, minute, 0, 0)
  return date.toISOString()
}

function daysFromWeekStart(offset: number): string {
  const date = new Date(now)
  const day = date.getDay()
  const mondayOffset = day === 0 ? -6 : 1 - day
  date.setDate(date.getDate() + mondayOffset + offset)
  date.setHours(0, 0, 0, 0)
  return date.toISOString().slice(0, 10)
}

export const mockWeekSummary: WeekSummary = {
  weekLabel: 'This week',
  totalDistanceM: 58_400,
  totalDurationSec: 18_960,
  activitiesCount: 4,
  completedWorkouts: 4,
  plannedWorkouts: 6,
  loadScore: 72,
  readinessLabel: 'Productive fatigue',
  highlight: 'Strong long run on Sunday. Thursday threshold is still on track.',
}

export const mockActivities: Activity[] = [
  {
    id: 'act_001',
    title: 'Sunday long run',
    type: 'long_run',
    source: 'garmin',
    status: 'ready',
    startedAt: daysAgo(1, 8, 5),
    durationSec: 6120,
    distanceM: 21_100,
    avgPaceSecPerKm: 290,
    avgHeartRate: 148,
    elevationGainM: 186,
    summary: 'Solid aerobic execution. Final 4 km drifted 8s/km, still within plan.',
  },
  {
    id: 'act_002',
    title: 'Friday strides + easy',
    type: 'easy',
    source: 'coros',
    status: 'ready',
    startedAt: daysAgo(3, 17, 40),
    durationSec: 2700,
    distanceM: 8_200,
    avgPaceSecPerKm: 330,
    avgHeartRate: 132,
    elevationGainM: 42,
    summary: 'Easy volume completed as prescribed. Strides looked sharp.',
  },
  {
    id: 'act_003',
    title: 'Wednesday intervals',
    type: 'intervals',
    source: 'fit_upload',
    status: 'analyzing',
    startedAt: daysAgo(5, 6, 50),
    durationSec: 3480,
    distanceM: 11_400,
    avgPaceSecPerKm: 305,
    avgHeartRate: 161,
    elevationGainM: 68,
    summary: null,
  },
  {
    id: 'act_004',
    title: 'Tuesday recovery jog',
    type: 'recovery',
    source: 'garmin',
    status: 'ready',
    startedAt: daysAgo(6, 18, 15),
    durationSec: 1980,
    distanceM: 5_600,
    avgPaceSecPerKm: 354,
    avgHeartRate: 121,
    elevationGainM: 18,
    summary: 'Kept truly easy. Good response after Monday strength.',
  },
  {
    id: 'act_005',
    title: 'Park tempo',
    type: 'tempo',
    source: 'coros',
    status: 'ready',
    startedAt: daysAgo(9, 7, 20),
    durationSec: 3120,
    distanceM: 10_000,
    avgPaceSecPerKm: 278,
    avgHeartRate: 156,
    elevationGainM: 54,
    summary: 'Controlled threshold block. Second half was the strongest.',
  },
  {
    id: 'act_006',
    title: 'Imported trail loop',
    type: 'easy',
    source: 'fit_upload',
    status: 'uploaded',
    startedAt: daysAgo(11, 9, 0),
    durationSec: 4200,
    distanceM: 12_300,
    avgPaceSecPerKm: 341,
    avgHeartRate: 139,
    elevationGainM: 240,
    summary: null,
  },
]

export const mockPlanDays: PlanDay[] = [
  { date: daysFromWeekStart(0), title: 'Recovery jog', type: 'recovery', completed: true, adapted: false },
  { date: daysFromWeekStart(1), title: 'Easy + drills', type: 'easy', completed: true, adapted: false },
  { date: daysFromWeekStart(2), title: 'Intervals 6x800', type: 'intervals', completed: true, adapted: false },
  { date: daysFromWeekStart(3), title: 'Easy aerobic', type: 'easy', completed: false, adapted: true },
  { date: daysFromWeekStart(4), title: 'Threshold 3x10', type: 'tempo', completed: false, adapted: false },
  { date: daysFromWeekStart(5), title: 'Rest or mobility', type: 'recovery', completed: false, adapted: false },
  { date: daysFromWeekStart(6), title: 'Long run 22k', type: 'long_run', completed: true, adapted: true },
]

export const mockDevices: DeviceConnection[] = [
  {
    id: 'dev_garmin',
    provider: 'garmin',
    label: 'Garmin',
    connected: true,
    lastSyncAt: daysAgo(0, 8, 40),
  },
  {
    id: 'dev_coros',
    provider: 'coros',
    label: 'Coros',
    connected: false,
    lastSyncAt: null,
  },
  {
    id: 'dev_strava',
    provider: 'strava',
    label: 'Strava',
    connected: false,
    lastSyncAt: null,
  },
]

export const mockDashboardData: DashboardData = {
  week: mockWeekSummary,
  activities: mockActivities,
  planDays: mockPlanDays,
  devices: mockDevices,
}
