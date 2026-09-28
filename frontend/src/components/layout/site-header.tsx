import { Link } from 'react-router-dom'

import { Button } from '@/components/ui/button'
import { useAuth } from '@/context/auth-context'
import { cn } from '@/lib/utils'

const navLinks = [
  { href: '#sync', label: 'Sync' },
  { href: '#analysis', label: 'Analysis' },
  { href: '#plans', label: 'Plans' },
  { href: '#coach', label: 'AI Coach' },
]

type SiteHeaderProps = {
  className?: string
}

export function SiteHeader({ className }: SiteHeaderProps) {
  const { user, loading, logout } = useAuth()

  return (
    <header
      className={cn(
        'relative z-20 mx-auto flex w-full max-w-6xl items-center justify-between gap-4 px-6 py-5',
        className,
      )}
    >
      <a href="#top" className="font-display text-xl font-semibold tracking-tight text-foreground">
        RaceCoach
      </a>

      <nav className="hidden items-center gap-7 md:flex" aria-label="Primary">
        {navLinks.map((link) => (
          <a
            key={link.href}
            href={link.href}
            className="text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
          >
            {link.label}
          </a>
        ))}
      </nav>

      <div className="flex items-center gap-3">
        {loading ? (
          <Button variant="outline" disabled>
            Checking session…
          </Button>
        ) : user ? (
          <>
            <Button asChild variant="outline">
              <Link to="/dashboard">Dashboard</Link>
            </Button>
            <Button variant="ghost" onClick={logout}>
              Log out
            </Button>
          </>
        ) : (
          <>
            <Button asChild variant="ghost" className="hidden sm:inline-flex">
              <Link to="/login">Sign in</Link>
            </Button>
            <Button asChild>
              <Link to="/login">Get started</Link>
            </Button>
          </>
        )}
      </div>
    </header>
  )
}
