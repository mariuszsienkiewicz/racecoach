import { Link } from 'react-router-dom'

const footerColumns = [
  {
    title: 'Product',
    links: [
      { href: '#sync', label: 'Device sync' },
      { href: '#analysis', label: 'Workout analysis' },
      { href: '#plans', label: 'Adaptive plans' },
      { href: '#coach', label: 'AI coach' },
    ],
  },
  {
    title: 'Athletes',
    links: [
      { href: '#sync', label: 'Garmin & Coros' },
      { href: '#sync', label: '.FIT uploads' },
      { href: '#plans', label: 'Race prep' },
      { href: '/login', label: 'Athlete login', isRoute: true },
      { href: '/register', label: 'Create account', isRoute: true },
    ],
  },
  {
    title: 'Company',
    links: [
      { href: '#top', label: 'About RaceCoach' },
      { href: 'mailto:hello@racecoach.app', label: 'Contact' },
    ],
  },
]

export function SiteFooter() {
  return (
    <footer className="relative border-t border-border/70 bg-[#eef3ee]/70">
      <div className="mx-auto grid w-full max-w-6xl gap-10 px-6 py-14 md:grid-cols-[1.4fr_1fr_1fr_1fr]">
        <div className="max-w-sm space-y-4">
          <p className="font-display text-2xl font-semibold tracking-tight">RaceCoach</p>
          <p className="text-sm leading-relaxed text-muted-foreground">
            The AI running coach that syncs your watches, analyzes every session, and keeps your
            training plan honest to how you actually train.
          </p>
        </div>

        {footerColumns.map((column) => (
          <div key={column.title} className="space-y-4">
            <p className="text-sm font-semibold tracking-wide text-foreground">{column.title}</p>
            <ul className="space-y-2.5">
              {column.links.map((link) => (
                <li key={`${column.title}-${link.label}`}>
                  {'isRoute' in link && link.isRoute ? (
                    <Link
                      to={link.href}
                      className="text-sm text-muted-foreground transition-colors hover:text-foreground"
                    >
                      {link.label}
                    </Link>
                  ) : (
                    <a
                      href={link.href}
                      className="text-sm text-muted-foreground transition-colors hover:text-foreground"
                    >
                      {link.label}
                    </a>
                  )}
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>

      <div className="mx-auto flex w-full max-w-6xl flex-col gap-2 border-t border-border/60 px-6 py-6 text-sm text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
        <p>© {new Date().getFullYear()} RaceCoach. Built for runners who want smarter feedback.</p>
        <p>Sync. Analyze. Adapt. Race.</p>
      </div>
    </footer>
  )
}
