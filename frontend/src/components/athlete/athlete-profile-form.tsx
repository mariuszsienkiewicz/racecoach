import { useEffect, useState, type FormEvent } from 'react'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useAuth } from '@/context/auth-context'
import {
  previewAthleteZones,
  saveAthleteProfile,
} from '@/lib/api'
import type {
  AthleteProfile,
  AthleteSex,
  HrMaxSource,
  HrZone,
  PrimaryGoal,
  ZoneMethod,
} from '@/types/athlete-profile'
import { cn } from '@/lib/utils'

type Mode = 'quick' | 'custom'

type AthleteProfileFormProps = {
  initial?: AthleteProfile | null
  submitLabel: string
  completeOnboarding?: boolean
  onSaved: (profile: AthleteProfile) => void
}

const goals: { value: PrimaryGoal; label: string }[] = [
  { value: 'general', label: 'General fitness' },
  { value: '5k', label: '5K' },
  { value: '10k', label: '10K' },
  { value: 'half', label: 'Half marathon' },
  { value: 'marathon', label: 'Marathon' },
]

function emptyCustomZones(): HrZone[] {
  return [1, 2, 3, 4, 5].map((zone) => ({ zone, minBpm: 0, maxBpm: 0 }))
}

export function AthleteProfileForm({
  initial,
  submitLabel,
  completeOnboarding = true,
  onSaved,
}: AthleteProfileFormProps) {
  const { token } = useAuth()
  const [mode, setMode] = useState<Mode>(
    initial?.zoneMethod === 'custom' ? 'custom' : 'quick',
  )
  const [birthYear, setBirthYear] = useState(initial?.birthYear?.toString() ?? '')
  const [hrMax, setHrMax] = useState(initial?.hrMax ? String(initial.hrMax) : '')
  const [hrRest, setHrRest] = useState(initial?.hrRest ? String(initial.hrRest) : '')
  const [useKarvonen, setUseKarvonen] = useState(initial?.zoneMethod === 'hrr_pct')
  const [primaryGoal, setPrimaryGoal] = useState<PrimaryGoal>(initial?.primaryGoal ?? 'general')
  const [sex, setSex] = useState<AthleteSex | ''>(initial?.sex ?? '')
  const [zones, setZones] = useState<HrZone[]>(
    initial?.zones?.length === 5 ? initial.zones : emptyCustomZones(),
  )
  const [previewZones, setPreviewZones] = useState<HrZone[]>(initial?.zones ?? [])
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (mode !== 'quick' || !token) {
      return
    }
    const year = birthYear ? Number(birthYear) : undefined
    const max = hrMax ? Number(hrMax) : undefined
    if (!year && !max) {
      return
    }
    const handle = window.setTimeout(() => {
      void previewAthleteZones(token, {
        birthYear: year,
        hrMax: max,
        hrRest: hrRest ? Number(hrRest) : null,
        zoneMethod: useKarvonen ? 'hrr_pct' : 'hrmax_pct',
      })
        .then((result) => {
          setPreviewZones(result.zones)
          if (!hrMax && result.hrMax) {
            setHrMax(String(result.hrMax))
          }
        })
        .catch(() => {
          /* preview is best-effort while typing */
        })
    }, 350)
    return () => window.clearTimeout(handle)
  }, [mode, birthYear, hrMax, hrRest, useKarvonen, token])

  async function onSubmit(event: FormEvent) {
    event.preventDefault()
    if (!token) {
      return
    }
    setError(null)
    setSubmitting(true)
    try {
      const year = birthYear ? Number(birthYear) : null
      let nextHrMax = hrMax ? Number(hrMax) : 0
      let hrMaxSource: HrMaxSource = 'manual'
      if (!nextHrMax && year) {
        nextHrMax = 220 - (new Date().getFullYear() - year)
        hrMaxSource = 'estimated_age'
      }

      const zoneMethod: ZoneMethod =
        mode === 'custom' ? 'custom' : useKarvonen ? 'hrr_pct' : 'hrmax_pct'

      const payload = {
        birthYear: year,
        sex: sex || null,
        primaryGoal,
        hrMax: nextHrMax,
        hrRest: hrRest ? Number(hrRest) : null,
        hrMaxSource,
        zoneMethod,
        zones: mode === 'custom' ? zones : undefined,
        completeOnboarding,
      }

      const result = await saveAthleteProfile(token, payload)
      onSaved(result.profile)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not save profile.')
    } finally {
      setSubmitting(false)
    }
  }

  const shownZones = mode === 'custom' ? zones : previewZones

  return (
    <form className="space-y-6" onSubmit={onSubmit}>
      <div className="flex flex-wrap gap-2">
        <Button
          type="button"
          size="sm"
          variant={mode === 'quick' ? 'default' : 'outline'}
          onClick={() => setMode('quick')}
        >
          Quick setup
        </Button>
        <Button
          type="button"
          size="sm"
          variant={mode === 'custom' ? 'default' : 'outline'}
          onClick={() => setMode('custom')}
        >
          I know my zones
        </Button>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-2">
          <Label htmlFor="birthYear">Birth year</Label>
          <Input
            id="birthYear"
            inputMode="numeric"
            placeholder="1990"
            value={birthYear}
            onChange={(e) => setBirthYear(e.target.value)}
          />
          <p className="text-xs text-muted-foreground">Used to estimate HRmax if you leave it blank.</p>
        </div>
        <div className="space-y-2">
          <Label htmlFor="hrMax">Max HR (bpm)</Label>
          <Input
            id="hrMax"
            inputMode="numeric"
            placeholder="190"
            value={hrMax}
            onChange={(e) => setHrMax(e.target.value)}
            required={mode === 'custom' || !birthYear}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="hrRest">Resting HR (optional)</Label>
          <Input
            id="hrRest"
            inputMode="numeric"
            placeholder="48"
            value={hrRest}
            onChange={(e) => setHrRest(e.target.value)}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="goal">Primary goal</Label>
          <select
            id="goal"
            className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs"
            value={primaryGoal}
            onChange={(e) => setPrimaryGoal(e.target.value as PrimaryGoal)}
          >
            {goals.map((goal) => (
              <option key={goal.value} value={goal.value}>
                {goal.label}
              </option>
            ))}
          </select>
        </div>
        <div className="space-y-2">
          <Label htmlFor="sex">Sex (optional)</Label>
          <select
            id="sex"
            className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs"
            value={sex}
            onChange={(e) => setSex(e.target.value as AthleteSex | '')}
          >
            <option value="">Prefer not to say</option>
            <option value="female">Female</option>
            <option value="male">Male</option>
            <option value="other">Other</option>
            <option value="unspecified">Unspecified</option>
          </select>
        </div>
      </div>

      {mode === 'quick' ? (
        <label className="flex items-start gap-3 text-sm text-muted-foreground">
          <input
            type="checkbox"
            className="mt-1"
            checked={useKarvonen}
            onChange={(e) => setUseKarvonen(e.target.checked)}
            disabled={!hrRest}
          />
          <span>
            Use heart-rate reserve (Karvonen) when resting HR is set - usually more realistic than %
            of HRmax alone.
          </span>
        </label>
      ) : null}

      <div className="space-y-3">
        <div>
          <p className="text-sm font-medium text-foreground">Your zones</p>
          <p className="text-xs text-muted-foreground">
            Coaching judges intensity by these zones, not absolute bpm.
          </p>
        </div>
        <div className="overflow-hidden rounded-2xl border border-border/70">
          <table className="w-full text-sm">
            <thead className="bg-secondary/50 text-left text-xs uppercase tracking-wide text-muted-foreground">
              <tr>
                <th className="px-3 py-2">Zone</th>
                <th className="px-3 py-2">Min</th>
                <th className="px-3 py-2">Max</th>
              </tr>
            </thead>
            <tbody>
              {(shownZones.length ? shownZones : emptyCustomZones()).map((zone, index) => (
                <tr key={zone.zone} className="border-t border-border/60">
                  <td className="px-3 py-2 font-medium">Z{zone.zone}</td>
                  <td className="px-3 py-2">
                    {mode === 'custom' ? (
                      <Input
                        inputMode="numeric"
                        className="h-8 w-24"
                        value={zones[index]?.minBpm || ''}
                        onChange={(e) => {
                          const next = [...zones]
                          next[index] = {
                            ...next[index],
                            minBpm: Number(e.target.value) || 0,
                          }
                          setZones(next)
                        }}
                      />
                    ) : (
                      <span className={cn(!zone.minBpm && 'text-muted-foreground')}>
                        {zone.minBpm || '-'}
                      </span>
                    )}
                  </td>
                  <td className="px-3 py-2">
                    {mode === 'custom' ? (
                      <Input
                        inputMode="numeric"
                        className="h-8 w-24"
                        value={zones[index]?.maxBpm || ''}
                        onChange={(e) => {
                          const next = [...zones]
                          next[index] = {
                            ...next[index],
                            maxBpm: Number(e.target.value) || 0,
                          }
                          setZones(next)
                        }}
                      />
                    ) : (
                      <span className={cn(!zone.maxBpm && 'text-muted-foreground')}>
                        {zone.maxBpm || '-'}
                      </span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {error ? <p className="text-sm text-destructive">{error}</p> : null}

      <Button type="submit" disabled={submitting} className="w-full sm:w-auto">
        {submitting ? 'Saving…' : submitLabel}
      </Button>
    </form>
  )
}
