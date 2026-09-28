import type { ActivityStatus, ActivityType } from '@/types/dashboard'
import { cn } from '@/lib/utils'

const statusStyles: Record<ActivityStatus, string> = {
  uploaded: 'bg-secondary text-secondary-foreground',
  analyzing: 'bg-amber-100 text-amber-900',
  ready: 'bg-primary/15 text-primary',
  failed: 'bg-destructive/10 text-destructive',
}

const statusLabels: Record<ActivityStatus, string> = {
  uploaded: 'Queued',
  analyzing: 'Reviewing',
  ready: 'Ready',
  failed: 'Failed',
}

const typeLabels: Record<ActivityType, string> = {
  easy: 'Easy',
  tempo: 'Tempo',
  intervals: 'Intervals',
  long_run: 'Long run',
  recovery: 'Recovery',
  race: 'Race',
}

export function StatusBadge({ status }: { status: ActivityStatus }) {
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-semibold tracking-wide',
        statusStyles[status],
      )}
    >
      {status === 'analyzing' || status === 'uploaded' ? (
        <span className="relative flex size-1.5">
          <span className="absolute inline-flex size-full animate-ping rounded-full bg-current opacity-40" />
          <span className="relative inline-flex size-1.5 rounded-full bg-current" />
        </span>
      ) : null}
      {statusLabels[status]}
    </span>
  )
}

export function typeLabel(type: ActivityType): string {
  return typeLabels[type]
}
