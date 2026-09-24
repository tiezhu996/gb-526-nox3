import type { GasMix } from './plan'

export interface DiverProfile {
  id: number
  profile_code: string
  display_name: string
  qualification_level: 'trainee' | 'commercial' | 'advanced' | 'supervisor'
  default_o2_fraction: number
  default_he_fraction: number
  default_n2_fraction: number
  profile_status: 'active' | 'inactive' | 'training_hold'
  limits_note: string
  version: number
  created_at: string
  updated_at: string
}

export interface CreateDiverProfile {
  profile_code: string
  display_name: string
  qualification_level: DiverProfile['qualification_level']
  default_o2_fraction: number
  default_he_fraction: number
  profile_status: DiverProfile['profile_status']
  limits_note: string
}

export const profileGas = (profile: DiverProfile): GasMix => ({
  o2: profile.default_o2_fraction,
  he: profile.default_he_fraction,
  n2: profile.default_n2_fraction,
})
