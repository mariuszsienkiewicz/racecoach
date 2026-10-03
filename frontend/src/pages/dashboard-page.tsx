import { useMemo, useState } from 'react'

import { AiCoachDrawer } from '@/components/dashboard/ai-coach-drawer'
import { ActivityCalendar } from '@/components/dashboard/activity-calendar'
import { ActivityFeed } from '@/components/dashboard/activity-feed'
import { DashboardShell } from '@/components/dashboard/dashboard-shell'
import { ImportPanel } from '@/components/dashboard/import-panel'
import { PlanSnapshot } from '@/components/dashboard/plan-snapshot'
import { WeekOverview } from '@/components/dashboard/week-overview'
import { useDashboard } from '@/hooks/use-dashboard'
import { toDateKey } from '@/lib/format'
import type { Activity } from '@/types/dashboard'
import { Button } from '@/components/ui/button'

export function DashboardPage() {
  const { data, loading, refreshing, error, refresh, uploadFit, reprocessActivity } =
    useDashboard()
  const [selectedDate, setSelectedDate] = useState<string | null>(null)
  const [coachActivity, setCoachActivity] = useState<Activity | null>(null)

  const filteredActivities = useMemo(() => {
    if (!data) {
      return []
    }
    if (!selectedDate) {
      return data.activities
    }
    return data.activities.filter(
      (activity) => toDateKey(new Date(activity.startedAt)) === selectedDate,
    )
  }, [data, selectedDate])

  if (loading && !data) {
    return (
      <DashboardShell>
        <p className="text-sm text-muted-foreground">Loading your activities…</p>
      </DashboardShell>
    )
  }

  if (error && !data) {
    return (
      <DashboardShell>
        <div className="space-y-3">
          <p className="text-sm text-destructive">{error}</p>
          <Button type="button" variant="outline" onClick={() => void refresh()}>
            Try again
          </Button>
        </div>
      </DashboardShell>
    )
  }

  if (!data) {
    return (
      <DashboardShell>
        <p className="text-sm text-muted-foreground">Dashboard unavailable.</p>
      </DashboardShell>
    )
  }

  return (
    <DashboardShell>
      <div className="space-y-10">
        <WeekOverview week={data.week} refreshing={refreshing} />

        <div className="grid gap-8 xl:grid-cols-[1.35fr_0.9fr] xl:gap-10">
          <div className="space-y-8">
            <ActivityCalendar
              activities={data.activities}
              selectedDate={selectedDate}
              onSelectDate={setSelectedDate}
            />
            <ActivityFeed
              activities={filteredActivities}
              onAskAi={setCoachActivity}
              onReprocess={reprocessActivity}
            />
          </div>

          <div className="space-y-8">
            <ImportPanel
              devices={data.devices}
              onUpload={async (file) => {
                await uploadFit(file)
              }}
            />
            <PlanSnapshot planDays={data.planDays} />
          </div>
        </div>
      </div>

      <AiCoachDrawer activity={coachActivity} onClose={() => setCoachActivity(null)} />
    </DashboardShell>
  )
}
