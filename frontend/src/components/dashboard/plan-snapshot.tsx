import { formatShortDate } from '@/lib/format'
import type { PlanDay } from '@/types/dashboard'
import { typeLabel } from '@/components/dashboard/status-badge'
import { ComingSoonBadge } from '@/components/ui/coming-soon-badge'
import { cn } from '@/lib/utils'

type PlanSnapshotProps = {
  planDays: PlanDay[]
}

export function PlanSnapshot({ planDays }: PlanSnapshotProps) {
  return (
    <section
      id="plan"
      className="scroll-mt-8 rounded-3xl border border-border/70 bg-background/75 p-6 backdrop-blur sm:p-8"
    >
      <div className="flex flex-wrap items-center gap-2">
        <p className="text-sm font-semibold uppercase tracking-[0.18em] text-primary">Plan</p>
        <ComingSoonBadge />
      </div>
      <h2 className="mt-2 font-display text-2xl font-semibold tracking-tight">This week's plan</h2>
      <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
        Preview of the adaptive schedule UI. Live plan generation and coach adaptations will land
        here later - the layout stays so we can wire the API without rebuilding the desk.
      </p>

      <ol className="mt-8 space-y-3" aria-hidden="true">
        {planDays.map((day) => (
          <li
            key={day.date}
            className={cn(
              'flex items-start justify-between gap-4 rounded-2xl border px-4 py-3 opacity-80',
              day.completed
                ? 'border-primary/20 bg-primary/5'
                : 'border-border/70 bg-[#f7f4ec]/35',
            )}
          >
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.14em] text-muted-foreground">
                {formatShortDate(day.date)}
              </p>
              <p className="mt-1 font-medium">{day.title}</p>
              <p className="text-sm text-muted-foreground">{typeLabel(day.type)}</p>
            </div>
            <div className="flex flex-col items-end gap-1 text-xs font-semibold uppercase tracking-[0.12em]">
              <span className={day.completed ? 'text-primary' : 'text-muted-foreground'}>
                {day.completed ? 'Done' : 'Planned'}
              </span>
              {day.adapted ? <span className="text-amber-800">Adapted</span> : null}
            </div>
          </li>
        ))}
      </ol>
    </section>
  )
}
