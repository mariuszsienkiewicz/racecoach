import { Navigate, useNavigate } from 'react-router-dom'

import { AthleteProfileForm } from '@/components/athlete/athlete-profile-form'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { useAuth } from '@/context/auth-context'

export function OnboardingPage() {
  const { user, loading, refreshUser } = useAuth()
  const navigate = useNavigate()

  if (!loading && !user) {
    return <Navigate to="/login" replace />
  }

  if (!loading && user?.onboardingCompleted) {
    return <Navigate to="/dashboard" replace />
  }

  return (
    <main className="relative flex min-h-svh items-center justify-center overflow-hidden px-4 py-10">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top,_rgba(34,120,80,0.18),_transparent_55%),linear-gradient(160deg,#f7f4ec_0%,#e8efe8_45%,#d7e4da_100%)]" />
      <div className="relative w-full max-w-2xl">
        <Card className="border-border/70 bg-card/90 backdrop-blur">
          <CardHeader>
            <CardTitle className="font-display text-3xl">Set up your coaching profile</CardTitle>
            <CardDescription>
              Tell RaceCoach how to read your heart-rate zones. Quick setup estimates them from HRmax
              (or age). Advanced athletes can paste exact bpm bounds from a watch.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <AthleteProfileForm
              submitLabel="Save and open dashboard"
              completeOnboarding
              onSaved={async () => {
                await refreshUser()
                navigate('/dashboard', { replace: true })
              }}
            />
          </CardContent>
        </Card>
      </div>
    </main>
  )
}
