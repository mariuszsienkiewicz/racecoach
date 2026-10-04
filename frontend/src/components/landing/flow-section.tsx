import { Link } from 'react-router-dom'

import { Reveal } from '@/components/landing/reveal'
import { Button } from '@/components/ui/button'

const flow = [
  {
    title: 'Bring workouts in',
    copy: 'Automatic pulls from Coros, Garmin, and more, plus manual `.fit` uploads when you need them.',
  },
  {
    title: 'Read the session',
    copy: 'RaceCoach analyzes execution, pacing, and training stress so you know what the workout really said.',
  },
  {
    title: 'Adapt and ask',
    copy: 'Your plan updates from completed work, and the AI coach is there when you want a second opinion.',
  },
]

export function FlowSection() {
  return (
    <section className="relative border-t border-border/50 py-24 sm:py-28">
      <div className="mx-auto w-full max-w-6xl px-6">
        <Reveal>
          <p className="text-sm font-semibold uppercase tracking-[0.22em] text-primary">
            From watch to coaching
          </p>
          <h2 className="mt-4 max-w-3xl font-display text-3xl font-semibold tracking-tight text-foreground sm:text-5xl">
            A continuous loop around every run.
          </h2>
        </Reveal>

        <div className="mt-14 grid gap-10 md:grid-cols-3">
          {flow.map((step, index) => (
            <Reveal key={step.title} delayMs={index * 90}>
              <p className="font-display text-sm font-semibold text-primary">0{index + 1}</p>
              <h3 className="mt-3 font-display text-2xl font-semibold tracking-tight">{step.title}</h3>
              <p className="mt-3 text-base leading-relaxed text-muted-foreground">{step.copy}</p>
            </Reveal>
          ))}
        </div>
      </div>
    </section>
  )
}

export function FinalCtaSection() {
  return (
    <section className="relative overflow-hidden border-t border-border/50 py-24 sm:py-28">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_50%_0%,rgba(34,120,80,0.16),transparent_55%),linear-gradient(180deg,#f4f1e8_0%,#e4ede6_100%)]" />
      <div className="relative mx-auto flex w-full max-w-6xl flex-col items-start gap-6 px-6 sm:items-center sm:text-center">
        <Reveal>
          <h2 className="max-w-3xl font-display text-3xl font-semibold tracking-tight text-foreground sm:text-5xl">
            Ready for coaching that keeps up with your training?
          </h2>
          <p className="mx-auto mt-5 max-w-2xl text-lg leading-relaxed text-muted-foreground">
            Sync your devices, review automatic analysis, follow an adaptive plan, and chat with an
            AI coach that already knows your week.
          </p>
          <div className="mt-8 flex flex-wrap gap-3 sm:justify-center">
            <Button asChild size="lg">
              <Link to="/register">Open RaceCoach</Link>
            </Button>
            <Button asChild size="lg" variant="outline">
              <a href="#sync">Explore the product</a>
            </Button>
          </div>
        </Reveal>
      </div>
    </section>
  )
}
