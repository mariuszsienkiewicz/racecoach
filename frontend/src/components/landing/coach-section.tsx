import { Reveal } from '@/components/landing/reveal'

const prompts = [
  'Why did my pace fall apart in the last repeat?',
  'Should I push the long run this weekend?',
  'Rewrite Thursday if I only have 45 minutes.',
]

export function CoachSection() {
  return (
    <section
      id="coach"
      className="relative scroll-mt-24 overflow-hidden border-t border-border/50 bg-[#e7efe8]/60 py-24 sm:py-28"
    >
      <div className="mx-auto grid w-full max-w-6xl gap-14 px-6 lg:grid-cols-[0.95fr_1.05fr] lg:items-center">
        <Reveal>
          <p className="text-sm font-semibold uppercase tracking-[0.22em] text-primary">AI coach</p>
          <h2 className="mt-4 font-display text-3xl font-semibold tracking-tight text-foreground sm:text-5xl">
            Talk through the week with a coach that knows your data.
          </h2>
          <p className="mt-5 text-lg leading-relaxed text-muted-foreground">
            Ask why a session felt off, how to rearrange a busy week, or what to prioritize before
            race day. The AI coach answers in context of your synced workouts, recent analysis, and
            current plan, not generic running advice.
          </p>
        </Reveal>

        <Reveal delayMs={120}>
          <div className="space-y-4">
            {prompts.map((prompt, index) => (
              <div
                key={prompt}
                className="rounded-2xl border border-border/70 bg-background/80 px-5 py-4 shadow-[0_10px_30px_-24px_rgba(20,60,40,0.45)]"
                style={{ marginLeft: index === 1 ? '1.5rem' : index === 2 ? '0.5rem' : 0 }}
              >
                <p className="text-xs font-semibold uppercase tracking-[0.16em] text-primary">
                  Athlete
                </p>
                <p className="mt-2 text-base leading-relaxed text-foreground">{prompt}</p>
              </div>
            ))}
            <div className="rounded-2xl border border-primary/20 bg-primary px-5 py-4 text-primary-foreground">
              <p className="text-xs font-semibold uppercase tracking-[0.16em] text-primary-foreground/75">
                RaceCoach
              </p>
              <p className="mt-2 text-base leading-relaxed">
                Your last two threshold sessions finished strong, but Tuesday’s long run elevated
                overnight recovery load. Keep Friday’s quality work, and move the long run to
                Sunday with an easier warmup.
              </p>
            </div>
          </div>
        </Reveal>
      </div>
    </section>
  )
}
