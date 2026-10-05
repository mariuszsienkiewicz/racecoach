import type { ReactNode } from 'react'
import { Link, NavLink } from 'react-router-dom'

import { Button } from '@/components/ui/button'
import { useAuth } from '@/context/auth-context'
import { cn } from '@/lib/utils'

const navItems = [
  { to: '/dashboard', label: 'Overview', end: true },
  { to: '/dashboard#calendar', label: 'Calendar', end: false },
  { to: '/dashboard#import', label: 'Import', end: false },
  { to: '/dashboard#plan', label: 'Plan', end: false },
  { to: '/settings', label: 'Profile & zones', end: true },
]

type DashboardShellProps = {
  children: ReactNode
}

export function DashboardShell({ children }: DashboardShellProps) {
  const { user, logout } = useAuth()

  return (
    <div className="min-h-svh bg-[linear-gradient(180deg,#f8f5ee_0%,#e8efe8_45%,#dce8df_100%)]">
      <div className="mx-auto flex w-full max-w-7xl flex-col gap-8 px-4 py-6 sm:px-6 lg:flex-row lg:gap-10 lg:px-8 lg:py-10">
        <aside className="lg:sticky lg:top-8 lg:flex lg:h-[calc(100svh-4rem)] lg:w-60 lg:shrink-0 lg:flex-col">
          <div className="rounded-3xl border border-border/70 bg-background/75 p-6 shadow-[0_20px_50px_-40px_rgba(20,60,40,0.45)] backdrop-blur">
            <Link to="/" className="font-display text-2xl font-semibold tracking-tight">
              RaceCoach
            </Link>
            <p className="mt-2 text-sm text-muted-foreground">Athlete dashboard</p>

            <nav className="mt-10 space-y-1.5" aria-label="Dashboard">
              {navItems.map((item) => (
                <NavLink
                  key={item.label}
                  to={item.to}
                  end={item.end}
                  className={({ isActive }) =>
                    cn(
                      'block rounded-xl px-3 py-2.5 text-sm font-medium transition-colors',
                      isActive && item.end
                        ? 'bg-primary text-primary-foreground'
                        : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground',
                    )
                  }
                >
                  {item.label}
                </NavLink>
              ))}
            </nav>

            <div className="mt-10 border-t border-border/70 pt-6">
              <p className="truncate text-sm font-medium">{user?.email}</p>
              <div className="mt-4 flex flex-col gap-2">
                <Button variant="ghost" size="sm" onClick={logout}>
                  Log out
                </Button>
              </div>
            </div>
          </div>
        </aside>

        <div className="min-w-0 flex-1 pb-12">{children}</div>
      </div>
    </div>
  )
}
