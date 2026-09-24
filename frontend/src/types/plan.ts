export type PlanStatus =
  | 'draft'
  | 'modeled'
  | 'pending_supervisor_review'
  | 'approved_for_training'
  | 'archived'

export const PLAN_STATUSES: PlanStatus[] = [
  'draft',
  'modeled',
  'pending_supervisor_review',
  'approved_for_training',
  'archived',
]

export interface GasMix {
  o2: number
  he: number
  n2: number
}

export interface DivePlan {
  id: number
  plan_code: string
  diver_profile_id: number
  diver_profile_code?: string
  worksite_pressure_bar: number
  breathing_mix: GasMix
  plan_status: PlanStatus
  created_by: number
  reviewed_by?: number
  version: number
  planned_at: string
  created_at: string
  updated_at: string
}

export interface CreateDivePlan {
  plan_code: string
  diver_profile_id: number
  worksite_pressure_bar: number
  breathing_mix: GasMix
  planned_at: string
}
