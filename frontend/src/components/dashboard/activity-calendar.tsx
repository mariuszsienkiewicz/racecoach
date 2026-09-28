import { useMemo, useState } from 'react'

import { Button } from '@/components/ui/button'
import { addMonths, startOfMonth, toDateKey } from '@/lib/format'
import type { Activity } from '@/types/dashboard'
import { cn } from '@/lib/utils'

type ActivityCalendarProps = {
  activities: Activity[]
  selectedDate: string | null
  onSelectDate: (dateKey: string | null) => void
}

const weekdayLabels = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']

export function ActivityCalendar({
  activities,
  selectedDate,
  onSelectDate,
}: ActivityCalendarProps) {
  const [month, setMonth] = useState(() => startOfMonth(new Date()))

  const activitiesByDay = useMemo(() => {
    const map = new Map<string, Activity[]>()
    for (const activity of activities) {
      const key = toDateKey(new Date(activity.startedAt))
      const bucket = map.get(key) ?? []
      bucket.push(activity)
      map.set(key, bucket)
    }
    return map
  }, [activities])

  const cells = useMemo(() => {
    const first = startOfMonth(month)
    const startOffset = (first.getDay() + 6) % 7
    const daysInMonth = new Date(month.getFullYear(), month.getMonth() + 1, 0).getDate()
    const totalCells = Math.ceil((startOffset + daysInMonth) / 7) * 7

    return Array.from({ length: totalCells }, (_, index) => {
      const dayNumber = index - startOffset + 1
      if (dayNumber < 1 || dayNumber > daysInMonth) {
        return null
      }
      const date = new Date(month.getFullYear(), month.getMonth(), dayNumber)
      return toDateKey(date)
    })
  }, [month])

  const monthLabel = new Intl.DateTimeFormat('en', {
    month: 'long',
    year: 'numeric',
  }).format(month)

  return (
    <section
      id="calendar"
      className="scroll-mt-8 rounded-3xl border border-border/70 bg-background/75 p-6 backdrop-blur sm:p-8"
    >
      <div className="flex items-center justify-between gap-3">
        <div>
          <p className="text-sm font-semibold uppercase tracking-[0.18em] text-primary">Calendar</p>
          <h2 className="mt-2 font-display text-2xl font-semibold tracking-tight">{monthLabel}</h2>
        </div>
        <div className="flex gap-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => setMonth((current) => addMonths(current, -1))}
          >
            Prev
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => setMonth(startOfMonth(new Date()))}
          >
            Today
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => setMonth((current) => addMonths(current, 1))}
          >
            Next
          </Button>
        </div>
      </div>

      <div className="mt-8 grid grid-cols-7 gap-2 text-center text-xs font-semibold uppercase tracking-[0.14em] text-muted-foreground">
        {weekdayLabels.map((label) => (
          <div key={label} className="py-1">
            {label}
          </div>
        ))}
      </div>

      <div className="mt-3 grid grid-cols-7 gap-2">
        {cells.map((dateKey, index) => {
          if (!dateKey) {
            return <div key={`empty-${index}`} className="min-h-16 rounded-2xl bg-transparent" />
          }

          const dayActivities = activitiesByDay.get(dateKey) ?? []
          const isSelected = selectedDate === dateKey
          const isToday = dateKey === toDateKey(new Date())

          return (
            <button
              key={dateKey}
              type="button"
              onClick={() => onSelectDate(isSelected ? null : dateKey)}
              className={cn(
                'min-h-16 rounded-2xl border px-2 py-2 text-left transition-colors',
                isSelected
                  ? 'border-primary bg-primary text-primary-foreground'
                  : 'border-border/60 bg-[#f7f4ec]/55 hover:border-primary/40 hover:bg-accent/60',
                isToday && !isSelected && 'ring-1 ring-primary/30',
              )}
            >
              <div className="flex items-center justify-between gap-1">
                <span className="text-sm font-semibold">{Number(dateKey.slice(-2))}</span>
                {dayActivities.length > 0 ? (
                  <span
                    className={cn(
                      'text-[10px] font-semibold uppercase tracking-wide',
                      isSelected ? 'text-primary-foreground/80' : 'text-primary',
                    )}
                  >
                    {dayActivities.length}
                  </span>
                ) : null}
              </div>
              <div className="mt-2 flex flex-wrap gap-1">
                {dayActivities.slice(0, 3).map((activity) => (
                  <span
                    key={activity.id}
                    className={cn(
                      'h-1.5 w-1.5 rounded-full',
                      isSelected ? 'bg-primary-foreground' : 'bg-primary',
                    )}
                  />
                ))}
              </div>
            </button>
          )
        })}
      </div>

      {selectedDate ? (
        <p className="mt-4 text-sm text-muted-foreground">
          Showing activities for{' '}
          <span className="font-medium text-foreground">
            {new Intl.DateTimeFormat('en', {
              weekday: 'long',
              month: 'long',
              day: 'numeric',
            }).format(new Date(`${selectedDate}T12:00:00`))}
          </span>
          . Click the day again to clear the filter.
        </p>
      ) : (
        <p className="mt-4 text-sm text-muted-foreground">
          Select a day to filter imported activities in the feed below.
        </p>
      )}
    </section>
  )
}
