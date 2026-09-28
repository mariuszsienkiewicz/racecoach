import type { Activity, WeekSummary } from '@/types/dashboard'

function startOfWeekMonday(date: Date): Date {
  const d = new Date(date)
  const day = d.getDay()
  const mondayOffset = day === 0 ? -6 : 1 - day
  d.setHours(0, 0, 0, 0)
  d.setDate(d.getDate() + mondayOffset)
  return d
}

function endOfWeekSunday(weekStart: Date): Date {
  const d = new Date(weekStart)
  d.setDate(d.getDate() + 7)
  return d
}

/** Metrics are usable once distance is present (status may still be analyzing). */
function hasMetrics(activity: Activity): boolean {
  return activity.distanceM > 0
}

export function buildWeekSummary(activities: Activity[], now = new Date()): WeekSummary {
  const weekStart = startOfWeekMonday(now)
  const weekEnd = endOfWeekSunday(weekStart)

  const thisWeek = activities.filter((activity) => {
    const started = new Date(activity.startedAt)
    return started >= weekStart && started < weekEnd
  })

  const withMetrics = thisWeek.filter(hasMetrics)
  const fullyReady = thisWeek.filter((activity) => activity.status === 'ready')
  const pending = thisWeek.filter(
    (activity) => activity.status === 'uploaded' || activity.status === 'analyzing',
  )
  const failed = thisWeek.filter((activity) => activity.status === 'failed')

  const totalDistanceM = withMetrics.reduce((sum, a) => sum + a.distanceM, 0)
  const totalDurationSec = withMetrics.reduce((sum, a) => sum + a.durationSec, 0)

  let highlight = 'Upload a .fit file to see your week fill in with real sessions.'
  if (fullyReady.length > 0) {
    const longest = [...fullyReady].sort((a, b) => b.distanceM - a.distanceM)[0]
    highlight = `Longest fully analyzed session: ${longest.title}.`
  } else if (withMetrics.length > 0) {
    highlight = `${withMetrics.length} session${withMetrics.length === 1 ? '' : 's'} have metrics; full analysis still running.`
  } else if (pending.length > 0) {
    highlight = `${pending.length} session${pending.length === 1 ? '' : 's'} still being analyzed.`
  } else if (failed.length > 0) {
    highlight = `${failed.length} import${failed.length === 1 ? '' : 's'} failed, re-upload to retry.`
  } else if (thisWeek.length === 0 && activities.length > 0) {
    highlight = 'No sessions started this week yet. Earlier imports are in the feed below.'
  }

  const readinessLabel =
    pending.length > 0
      ? 'Pipeline running'
      : fullyReady.length > 0
        ? 'Analysis complete'
        : activities.length > 0
          ? 'Quiet week'
          : 'No imports yet'

  return {
    weekLabel: 'This week',
    totalDistanceM,
    totalDurationSec,
    activitiesCount: thisWeek.length,
    completedWorkouts: fullyReady.length,
    plannedWorkouts: Math.max(thisWeek.length, fullyReady.length),
    loadScore: Math.min(100, Math.round(totalDistanceM / 1000) * 4),
    readinessLabel,
    highlight,
  }
}

export function hasPendingActivities(activities: Activity[]): boolean {
  return activities.some(
    (activity) => activity.status === 'uploaded' || activity.status === 'analyzing',
  )
}
