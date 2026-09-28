import { Reveal } from '@/components/landing/reveal'

const signals = [
  {
    title: 'Execution quality',
    copy: 'Did the intervals hit the right effort, or did the session drift?',
  },
  {
    title: 'Pacing & consistency',
    copy: 'See where splits held strong and where fatigue or surges showed up.',
  },
  {
    title: 'Load awareness',
    copy: 'Understand how today sits against recent training stress and recovery.',
  },
  {
    title: 'Clear next action',
    copy: 'Leave each workout with specific feedback, not a wall of charts.',
  },
]

export function AnalysisSection() {
  return (
    <section
      id="analysis"
      className="relative scroll-mt-24 overflow-hidden border-t border-border/50 bg-[#e8efe8]/55 py-24 sm:py-28"
    >
      <div className="pointer-events-none absolute inset-y-0 right-0 w-1/2 bg-[radial-gradient(circle_at_80%_40%,rgba(34,120,80,0.14),transparent_55%)]" />

      <div className="relative mx-auto w-full max-w-6xl px-6">
        <Reveal>
          <p className="text-sm font-semibold uppercase tracking-[0.22em] text-primary">Analysis</p>
          <h2 className="mt-4 max-w-3xl font-display text-3xl font-semibold tracking-tight text-foreground sm:text-5xl">
            Upload once. Get coaching-grade feedback.
          </h2>
          <p className="mt-5 max-w-2xl text-lg leading-relaxed text-muted-foreground">
            RaceCoach analyzes each activity automatically, calling out execution, pacing patterns,
            and what actually matters for tomorrow's training.
          </p>
        </Reveal>

        <div className="mt-14 grid gap-x-10 gap-y-10 sm:grid-cols-2">
          {signals.map((signal, index) => (
            <Reveal key={signal.title} delayMs={index * 80}>
              <div className="max-w-md">
                <p className="font-display text-xs font-semibold uppercase tracking-[0.2em] text-primary/80">
                  0{index + 1}
                </p>
                <h3 className="mt-3 font-display text-2xl font-semibold tracking-tight">
                  {signal.title}
                </h3>
                <p className="mt-2 text-base leading-relaxed text-muted-foreground">{signal.copy}</p>
              </div>
            </Reveal>
          ))}
        </div>
      </div>
    </section>
  )
}
