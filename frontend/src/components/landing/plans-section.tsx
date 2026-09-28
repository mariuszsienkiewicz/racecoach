import { Reveal } from '@/components/landing/reveal'
import { ComingSoonBadge } from '@/components/ui/coming-soon-badge'

export function PlansSection() {
  return (
    <section id="plans" className="relative scroll-mt-24 border-t border-border/50 py-24 sm:py-28">
      <div className="mx-auto grid w-full max-w-6xl gap-12 px-6 lg:grid-cols-2 lg:gap-16">
        <Reveal>
          <div className="flex flex-wrap items-center gap-3">
            <p className="text-sm font-semibold uppercase tracking-[0.22em] text-primary">
              Adaptive plans
            </p>
            <ComingSoonBadge />
          </div>
          <h2 className="mt-4 font-display text-3xl font-semibold tracking-tight text-foreground sm:text-5xl">
            Build a plan. Watch it get smarter.
          </h2>
          <p className="mt-5 text-lg leading-relaxed text-muted-foreground">
            Create a structured training plan for your next race or fitness block. As you complete
            workouts, RaceCoach adjusts upcoming sessions so the plan stays realistic, harder when
            you are responding well, lighter when the data says you need recovery.
          </p>
        </Reveal>

        <Reveal delayMs={120}>
          <div className="relative overflow-hidden rounded-3xl border border-border/70 bg-[#f7f4ec]/80 px-6 py-8 opacity-90 sm:px-8">
            <div className="absolute inset-0 bg-[linear-gradient(135deg,rgba(34,120,80,0.08),transparent_45%)]" />
            <div className="relative space-y-6" aria-hidden="true">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">
                  This week
                </p>
                <p className="mt-2 font-display text-2xl font-semibold tracking-tight">
                  Plan reshaped after Tuesday’s long run
                </p>
              </div>

              <ol className="space-y-4">
                <li className="flex gap-4">
                  <span className="mt-1 h-2.5 w-2.5 shrink-0 rounded-full bg-primary" />
                  <div>
                    <p className="font-medium">Wednesday threshold → easy aerobic</p>
                    <p className="text-sm text-muted-foreground">
                      Elevated fatigue markers after a strong long run.
                    </p>
                  </div>
                </li>
                <li className="flex gap-4">
                  <span className="mt-1 h-2.5 w-2.5 shrink-0 rounded-full bg-primary/50" />
                  <div>
                    <p className="font-medium">Friday intervals kept</p>
                    <p className="text-sm text-muted-foreground">
                      Execution quality stayed on target earlier in the week.
                    </p>
                  </div>
                </li>
                <li className="flex gap-4">
                  <span className="mt-1 h-2.5 w-2.5 shrink-0 rounded-full bg-primary" />
                  <div>
                    <p className="font-medium">Sunday long run nudged +2 km</p>
                    <p className="text-sm text-muted-foreground">
                      Volume response looks durable enough to progress.
                    </p>
                  </div>
                </li>
              </ol>
            </div>
          </div>
        </Reveal>
      </div>
    </section>
  )
}
