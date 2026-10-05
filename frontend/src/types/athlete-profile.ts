export type ZoneMethod = 'hrmax_pct' | 'hrr_pct' | 'custom'

export type HrMaxSource = 'manual' | 'estimated_age'

export type PrimaryGoal = 'general' | '5k' | '10k' | 'half' | 'marathon'

export type AthleteSex = 'male' | 'female' | 'other' | 'unspecified'

export type HrZone = {
  zone: number
  minBpm: number
  maxBpm: number
}

export type AthleteProfile = {
  birthYear: number | null
  sex: AthleteSex | null
  primaryGoal: PrimaryGoal
  hrMax: number
  hrRest: number | null
  hrMaxSource: HrMaxSource
  zoneMethod: ZoneMethod
  zones: HrZone[]
  zonesUpdatedAt: string
  onboardingCompleted: boolean
  onboardingCompletedAt: string | null
}

export type AthleteProfilePayload = {
  birthYear?: number | null
  sex?: AthleteSex | null
  primaryGoal?: PrimaryGoal
  hrMax?: number
  hrRest?: number | null
  hrMaxSource?: HrMaxSource
  zoneMethod?: ZoneMethod
  zones?: HrZone[]
  completeOnboarding?: boolean
}
