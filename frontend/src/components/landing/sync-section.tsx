import { Reveal } from '@/components/landing/reveal'
import { ComingSoonBadge } from '@/components/ui/coming-soon-badge'

const sources = [
  {
    name: 'Coros',
    detail: 'Automatic activity pull after every session.',
    status: 'soon' as const,
  },
  {
    name: 'Garmin',
    detail: 'Keep your watch workflow, RaceCoach stays in sync.',
    status: 'soon' as const,
  },
  {
    name: '.FIT upload',
    detail: 'Drop a file anytime when a device sync is not enough.',
    status: 'live' as const,
  },
]

export function SyncSection() {
  return (
    <section id="sync" className="relative scroll-mt-24 border-t border-border/50 py-24 sm:py-28">
      <div className="mx-auto grid w-full max-w-6xl gap-14 px-6 lg:grid-cols-[1.1fr_0.9fr] lg:items-end">
        <Reveal>
          <p className="text-sm font-semibold uppercase tracking-[0.22em] text-primary">Data in</p>
          <h2 className="mt-4 max-w-xl font-display text-3xl font-semibold tracking-tight text-foreground sm:text-5xl">
            Every run lands in one place, automatically.
          </h2>
          <p className="mt-5 max-w-lg text-lg leading-relaxed text-muted-foreground">
            `.fit` upload is live today. Coros, Garmin, and other providers keep this same layout -
            OAuth sync plugs in without rebuilding the intake story.
          </p>
        </Reveal>

        <Reveal delayMs={120}>
          <ul className="space-y-0 divide-y divide-border/70 border-y border-border/70">
            {sources.map((source) => (
              <li key={source.name} className="flex items-start justify-between gap-6 py-5">
                <div>
                  <p className="font-display text-xl font-semibold tracking-tight">{source.name}</p>
                  <p className="mt-1 text-sm leading-relaxed text-muted-foreground">{source.detail}</p>
                </div>
                {source.status === 'live' ? (
                  <span className="mt-1 text-xs font-semibold uppercase tracking-[0.18em] text-primary">
                    Live
                  </span>
                ) : (
                  <ComingSoonBadge className="mt-1 shrink-0" />
                )}
              </li>
            ))}
          </ul>
        </Reveal>
      </div>
    </section>
  )
}
