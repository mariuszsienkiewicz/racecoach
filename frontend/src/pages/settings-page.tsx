import { useEffect, useState } from 'react'

import { AthleteProfileForm } from '@/components/athlete/athlete-profile-form'
import { DashboardShell } from '@/components/dashboard/dashboard-shell'
import { useAuth } from '@/context/auth-context'
import { fetchAthleteProfile } from '@/lib/api'
import type { AthleteProfile } from '@/types/athlete-profile'

export function SettingsPage() {
  const { token, refreshUser } = useAuth()
  const [profile, setProfile] = useState<AthleteProfile | null>(null)
  const [loading, setLoading] = useState(true)
  const [savedNote, setSavedNote] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!token) {
      return
    }
    let cancelled = false
    void fetchAthleteProfile(token)
      .then((result) => {
        if (!cancelled) {
          setProfile(result.profile)
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Could not load profile.')
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false)
        }
      })
    return () => {
      cancelled = true
    }
  }, [token])

  return (
    <DashboardShell>
      <section className="space-y-6">
        <div>
          <p className="text-sm font-semibold uppercase tracking-[0.18em] text-primary">Settings</p>
          <h1 className="mt-2 font-display text-3xl font-semibold tracking-tight">Athlete profile</h1>
          <p className="mt-2 max-w-2xl text-muted-foreground">
            Update HRmax, resting HR, or custom zone bounds. New analysis uses these zones so easy
            work is not mislabeled just because absolute bpm looks high.
          </p>
        </div>

        {loading ? <p className="text-sm text-muted-foreground">Loading profile…</p> : null}
        {error ? <p className="text-sm text-destructive">{error}</p> : null}
        {savedNote ? <p className="text-sm text-primary">{savedNote}</p> : null}

        {!loading ? (
          <div className="rounded-3xl border border-border/70 bg-background/75 p-6 sm:p-8">
            <AthleteProfileForm
              initial={profile}
              submitLabel="Save profile"
              completeOnboarding
              onSaved={async (next) => {
                setProfile(next)
                setSavedNote('Profile saved. Refresh analysis on recent activities to apply new zones.')
                await refreshUser()
              }}
            />
          </div>
        ) : null}
      </section>
    </DashboardShell>
  )
}
