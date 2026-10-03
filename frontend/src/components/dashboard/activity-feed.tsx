import { useEffect, useRef, useState } from 'react'
import { MessageSquare, RefreshCw } from 'lucide-react'

import { StatusBadge, typeLabel } from '@/components/dashboard/status-badge'
import { Button } from '@/components/ui/button'
import {
  formatActivityDate,
  formatDistance,
  formatDuration,
  formatPace,
} from '@/lib/format'
import { cn } from '@/lib/utils'
import type { Activity } from '@/types/dashboard'

type ActivityFeedProps = {
  activities: Activity[]
  onAskAi: (activity: Activity) => void
  onReprocess?: (activityId: string) => Promise<void>
}

const REVIEW_MESSAGES = [
  'Reading your session…',
  'Looking at pace and effort…',
  'Checking splits and rhythm…',
  'Comparing effort across the run…',
  'Noticing where you pushed harder…',
  'Weighing recovery against intensity…',
  'Finding the story in your data…',
  'Pulling out what mattered most…',
  'Turning metrics into insight…',
  'Shaping a short coaching note…',
] as const

function isPending(status: Activity['status']) {
  return status === 'uploaded' || status === 'analyzing'
}

function splitCoachSummary(text: string): { headline: string | null; body: string } {
  const trimmed = text.trim()
  const parts = trimmed.split(/\n\n+/)
  if (parts.length >= 2 && parts[0].length > 0 && parts[0].length <= 120) {
    return { headline: parts[0], body: parts.slice(1).join('\n\n').trim() }
  }
  return { headline: null, body: trimmed }
}

function ThinkingDots() {
  return (
    <span className="rc-dot-wave inline-flex items-center gap-1.5" aria-hidden="true">
      <span className="size-1.5 rounded-full bg-primary/70" />
      <span className="size-1.5 rounded-full bg-primary/70" />
      <span className="size-1.5 rounded-full bg-primary/70" />
    </span>
  )
}

function ReviewingStatus({ hasMetrics }: { hasMetrics: boolean }) {
  const [index, setIndex] = useState(0)

  useEffect(() => {
    if (!hasMetrics) {
      return
    }
    const timer = window.setInterval(() => {
      setIndex((current) => (current + 1) % REVIEW_MESSAGES.length)
    }, 3200)
    return () => window.clearInterval(timer)
  }, [hasMetrics])

  const message = hasMetrics ? REVIEW_MESSAGES[index] : 'Getting your session ready…'

  return (
    <div className="space-y-3">
      <div className="flex items-center gap-3 text-sm text-muted-foreground">
        <ThinkingDots />
        <span key={message} className="rc-insight-in">
          {message}
        </span>
      </div>
      <div className="h-px w-full max-w-sm overflow-hidden rounded-full bg-border/80">
        <div className="rc-breathe-bar h-full w-full rounded-full bg-primary/40" />
      </div>
    </div>
  )
}

function MetricChip({
  label,
  value,
  pending,
  reveal,
}: {
  label: string
  value: string
  pending: boolean
  reveal: boolean
}) {
  if (pending) {
    return (
      <span
        className="rc-shimmer inline-block h-5 w-[4.25rem] rounded-md border border-border/50"
        title={`Waiting for ${label}`}
        aria-label={`Waiting for ${label}`}
      />
    )
  }

  return (
    <span className={cn(reveal && 'rc-metric-in')} title={label}>
      {value}
    </span>
  )
}

function CoachInsight({ summary, animateIn }: { summary: string; animateIn: boolean }) {
  const { headline, body } = splitCoachSummary(summary)

  return (
    <div className={cn('max-w-2xl space-y-2', animateIn && 'rc-insight-in')}>
      {headline ? (
        <p className="text-sm font-semibold leading-snug text-foreground">{headline}</p>
      ) : null}
      <p className="text-sm leading-relaxed text-foreground/80">{body}</p>
    </div>
  )
}

function ActivityRow({
  activity,
  onAskAi,
  onReprocess,
}: {
  activity: Activity
  onAskAi: (activity: Activity) => void
  onReprocess?: (activityId: string) => Promise<void>
}) {
  const ready = activity.status === 'ready'
  const pending = isPending(activity.status)
  const hasMetrics = activity.distanceM > 0 && activity.durationSec > 0
  const prevStatus = useRef(activity.status)
  const [justReady, setJustReady] = useState(false)
  const [metricsReveal, setMetricsReveal] = useState(hasMetrics)
  const [reprocessing, setReprocessing] = useState(false)
  const [reprocessError, setReprocessError] = useState<string | null>(null)

  useEffect(() => {
    if (prevStatus.current !== 'ready' && activity.status === 'ready') {
      setJustReady(true)
      const timer = window.setTimeout(() => setJustReady(false), 1200)
      prevStatus.current = activity.status
      return () => window.clearTimeout(timer)
    }
    prevStatus.current = activity.status
  }, [activity.status])

  useEffect(() => {
    if (hasMetrics) {
      setMetricsReveal(true)
    }
  }, [hasMetrics])

  return (
    <article
      className={cn(
        'relative flex flex-col gap-5 py-8 transition-colors duration-500 lg:flex-row lg:items-start lg:gap-8',
        pending && 'bg-[#f7f4ec]/55',
        justReady && 'bg-primary/[0.04]',
      )}
    >
      <div className="min-w-0 flex-1 space-y-4">
        <div className="space-y-2">
          <div className="flex flex-wrap items-center gap-2.5">
            <h3 className="font-display text-lg font-semibold tracking-tight">{activity.title}</h3>
            <StatusBadge status={activity.status} />
            <span className="rounded-full bg-secondary px-2.5 py-1 text-xs font-medium text-secondary-foreground">
              {typeLabel(activity.type)}
            </span>
          </div>
          <p className="text-sm text-muted-foreground">
            {formatActivityDate(activity.startedAt)} · {activity.source.replaceAll('_', ' ')}
          </p>
        </div>

        <div className="flex min-h-5 flex-wrap items-center gap-x-5 gap-y-2 text-sm text-muted-foreground">
          <MetricChip
            label="Distance"
            value={formatDistance(activity.distanceM)}
            pending={pending && !hasMetrics}
            reveal={metricsReveal}
          />
          <MetricChip
            label="Duration"
            value={formatDuration(activity.durationSec)}
            pending={pending && !hasMetrics}
            reveal={metricsReveal}
          />
          <MetricChip
            label="Pace"
            value={formatPace(activity.avgPaceSecPerKm)}
            pending={pending && !hasMetrics}
            reveal={metricsReveal}
          />
          {activity.avgHeartRate ? (
            <MetricChip
              label="Heart rate"
              value={`${activity.avgHeartRate} bpm`}
              pending={false}
              reveal={metricsReveal}
            />
          ) : pending && !hasMetrics ? (
            <MetricChip label="Heart rate" value="" pending reveal={false} />
          ) : null}
        </div>

        {activity.status === 'failed' ? (
          <p className="text-sm text-destructive">
            Something went wrong with this session. Try uploading the file again.
          </p>
        ) : null}

        {pending ? <ReviewingStatus hasMetrics={hasMetrics} /> : null}

        {ready && activity.summary ? (
          <CoachInsight summary={activity.summary} animateIn={justReady} />
        ) : null}

        {ready && !activity.summary ? (
          <p className="text-sm text-muted-foreground">
            Session is ready. Ask the AI coach for a deeper debrief.
          </p>
        ) : null}

        {activity.reprocessAvailable ? (
          <div className="space-y-2 rounded-2xl bg-secondary/40 px-3 py-3">
            <p className="text-sm text-foreground/85">
              {activity.reprocessReason === 'stale_pipeline'
                ? 'Analysis got stuck, refresh to finish processing this session.'
                : 'Newer analysis is available, refresh for better pace splits and AI feedback.'}
            </p>
            <Button
              type="button"
              variant="secondary"
              size="sm"
              disabled={reprocessing || !onReprocess}
              onClick={() => {
                if (!onReprocess) {
                  return
                }
                setReprocessError(null)
                setReprocessing(true)
                void onReprocess(activity.id)
                  .catch((err) => {
                    setReprocessError(
                      err instanceof Error ? err.message : 'Could not refresh analysis.',
                    )
                  })
                  .finally(() => setReprocessing(false))
              }}
            >
              <RefreshCw className={cn('h-3.5 w-3.5', reprocessing && 'animate-spin')} />
              {reprocessing ? 'Refreshing…' : 'Refresh analysis'}
            </Button>
            {reprocessError ? (
              <p className="text-xs text-destructive" role="alert">
                {reprocessError}
              </p>
            ) : null}
          </div>
        ) : null}
      </div>

      <div className="lg:pt-1">
        <Button
          type="button"
          variant="outline"
          className={cn(
            'shrink-0 transition-opacity duration-300',
            pending && 'opacity-60',
            justReady && 'rc-insight-in',
          )}
          disabled={!ready}
          onClick={() => onAskAi(activity)}
        >
          <MessageSquare className="h-4 w-4" />
          {ready ? 'Ask AI' : pending ? 'Reviewing…' : 'Waiting…'}
        </Button>
      </div>
    </article>
  )
}

export function ActivityFeed({ activities, onAskAi, onReprocess }: ActivityFeedProps) {
  return (
    <section className="rounded-3xl border border-border/70 bg-background/75 p-6 backdrop-blur sm:p-8">
      <div className="flex items-end justify-between gap-4">
        <div>
          <p className="text-sm font-semibold uppercase tracking-[0.18em] text-primary">Activities</p>
          <h2 className="mt-2 font-display text-2xl font-semibold tracking-tight">
            Imported sessions
          </h2>
        </div>
        <p className="text-sm text-muted-foreground">{activities.length} shown</p>
      </div>

      <div className="mt-8 divide-y divide-border/60 border-y border-border/60">
        {activities.length === 0 ? (
          <p className="py-10 text-sm text-muted-foreground">
            No activities yet. Upload a `.fit` file to see your real imports here.
          </p>
        ) : (
          activities.map((activity) => (
            <ActivityRow
              key={activity.id}
              activity={activity}
              onAskAi={onAskAi}
              onReprocess={onReprocess}
            />
          ))
        )}
      </div>
    </section>
  )
}
