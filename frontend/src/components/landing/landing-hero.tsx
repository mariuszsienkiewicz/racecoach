import { Link } from 'react-router-dom'

import { Button } from '@/components/ui/button'
import { useAuth } from '@/context/auth-context'

function HeroTrailVisual() {
  return (
    <div className="pointer-events-none absolute inset-x-0 bottom-0 h-[48vh] min-h-[280px] w-full overflow-hidden sm:h-[52vh]">
      <svg
        className="absolute inset-0 h-full w-full"
        viewBox="0 0 1440 520"
        preserveAspectRatio="xMidYMax slice"
        aria-hidden="true"
      >
        <defs>
          <linearGradient id="rc-trail-fill" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="rgba(34,120,80,0.22)" />
            <stop offset="55%" stopColor="rgba(34,120,80,0.08)" />
            <stop offset="100%" stopColor="rgba(34,120,80,0)" />
          </linearGradient>
          <linearGradient id="rc-trail-line" x1="0" y1="0" x2="1" y2="0">
            <stop offset="0%" stopColor="rgba(28,92,62,0.15)" />
            <stop offset="35%" stopColor="rgba(28,92,62,0.9)" />
            <stop offset="100%" stopColor="rgba(28,92,62,0.25)" />
          </linearGradient>
        </defs>

        <path
          d="M0,360 C180,300 260,250 420,270 C580,290 640,360 780,330 C920,300 980,210 1120,230 C1260,250 1340,310 1440,280 L1440,520 L0,520 Z"
          fill="url(#rc-trail-fill)"
        />

        <path
          className="rc-draw-path"
          d="M-20,390 C140,330 240,250 390,265 C560,282 620,360 760,335 C900,310 970,205 1125,225 C1255,242 1345,300 1460,270"
          fill="none"
          stroke="url(#rc-trail-line)"
          strokeWidth="3.5"
          strokeLinecap="round"
        />

        <g className="rc-pulse-soft" fill="rgba(28,92,62,0.85)">
          <circle cx="390" cy="265" r="5" />
          <circle cx="760" cy="335" r="5" />
          <circle cx="1125" cy="225" r="5" />
        </g>

        <g fill="none" stroke="rgba(28,92,62,0.12)" strokeWidth="1">
          <path d="M80,120 C220,90 320,140 460,110" />
          <path d="M980,90 C1100,70 1220,120 1360,95" />
          <path d="M540,80 C680,55 760,95 880,70" />
        </g>
      </svg>

      <div className="absolute inset-x-0 bottom-0 h-24 bg-gradient-to-t from-[#d8e4db] via-transparent to-transparent" />
    </div>
  )
}

export function LandingHero() {
  const { user } = useAuth()
  const primaryHref = user ? '/dashboard' : '/register'
  const primaryLabel = user ? 'Open dashboard' : 'Start coaching'

  return (
    <section className="relative isolate flex min-h-[calc(100svh-4.5rem)] flex-col justify-start overflow-hidden pt-10 sm:pt-16">
      <div className="rc-drift pointer-events-none absolute inset-0 -z-10 bg-[radial-gradient(circle_at_18%_12%,rgba(34,120,80,0.22),transparent_42%),radial-gradient(circle_at_88%_8%,rgba(20,60,40,0.16),transparent_34%),linear-gradient(180deg,#f8f5ee_0%,#e8efe8_52%,#d6e3d9_100%)]" />

      <div className="relative z-10 mx-auto w-full max-w-6xl px-6 pb-[46vh] sm:pb-[48vh]">
        <p className="rc-fade-up text-sm font-semibold uppercase tracking-[0.28em] text-primary">
          AI running coach
        </p>
        <h1 className="rc-fade-up rc-fade-up-delay-1 mt-4 font-display text-6xl font-semibold leading-[0.95] tracking-tight text-foreground sm:text-7xl lg:text-8xl">
          RaceCoach
        </h1>
        <p className="rc-fade-up rc-fade-up-delay-2 mt-6 max-w-xl text-lg leading-relaxed text-muted-foreground sm:text-xl">
          Sync every run, get clear execution feedback, and let your plan adapt when life, or a
          hard session, changes the week.
        </p>
        <div className="rc-fade-up rc-fade-up-delay-3 mt-8 flex flex-wrap items-center gap-3">
          <Button asChild size="lg">
            <Link to={primaryHref}>{primaryLabel}</Link>
          </Button>
          <Button asChild size="lg" variant="outline">
            <a href="#sync">See how it works</a>
          </Button>
        </div>
      </div>

      <HeroTrailVisual />
    </section>
  )
}
