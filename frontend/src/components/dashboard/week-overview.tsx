import type { WeekSummary } from '@/types/dashboard'
import { formatDistance, formatDuration } from '@/lib/format'

type WeekOverviewProps = {
  week: WeekSummary
  refreshing?: boolean
}

export function WeekOverview({ week, refreshing = false }: WeekOverviewProps) {
  const stats = [
    { label: 'Distance', value: formatDistance(week.totalDistanceM) },
    { label: 'Time', value: formatDuration(week.totalDurationSec) },
    {
      label: 'Sessions',
      value: `${week.completedWorkouts}/${week.activitiesCount || week.plannedWorkouts}`,
    },
    { label: 'Load', value: `${week.loadScore}` },
  ]

  return (
    <section className="space-y-7">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <p className="text-sm font-semibold uppercase tracking-[0.2em] text-primary">
            {week.weekLabel}
          </p>
          <h1 className="mt-2 font-display text-3xl font-semibold tracking-tight sm:text-4xl">
            Coaching desk
          </h1>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          {refreshing ? (
            <p className="rounded-full border border-border/70 bg-background/70 px-3 py-1.5 text-sm font-medium text-muted-foreground">
              Updating…
            </p>
          ) : null}
          <p className="rounded-full border border-border/70 bg-background/70 px-3 py-1.5 text-sm font-medium text-muted-foreground">
            {week.readinessLabel}
          </p>
        </div>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {stats.map((stat) => (
          <div
            key={stat.label}
            className="rounded-2xl border border-border/70 bg-background/75 px-5 py-5 backdrop-blur"
          >
            <p className="text-xs font-semibold uppercase tracking-[0.16em] text-muted-foreground">
              {stat.label}
            </p>
            <p className="mt-3 font-display text-2xl font-semibold tracking-tight">{stat.value}</p>
          </div>
        ))}
      </div>

      <p className="max-w-3xl text-base leading-relaxed text-muted-foreground">{week.highlight}</p>
    </section>
  )
}
