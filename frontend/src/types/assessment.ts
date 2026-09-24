import type { DivePlan } from './plan'
import type { ExposureSegment } from './segment'
import type { DiverProfile } from './diver'
import type { RiskBand, RiskFlag } from './risk'

export interface LoadPoint {
  sequence_no: number
  depth_m: number
  elapsed_min: number
  ambient_bar: number
  inspired_n2_bar: number
  inspired_he_bar: number
  n2_load_bar: number
  he_load_bar: number
  total_inert_bar: number
  relative_change: number
}

export interface CompartmentCurve {
  name: string
  n2_half_time_min: number
  he_half_time_min: number
  points: LoadPoint[]
}

export interface ModelAssumptions {
  purpose: string
  pressure_conversion: string
  water_vapor_bar: number
  kinetics: string
  configured_compartments: Array<{ name: string; n2_half_time_min: number; he_half_time_min: number }>
  prohibited_use: string[]
}

export interface InputSnapshot {
  plan: DivePlan & { breathing_mix_json?: string }
  diver: DiverProfile
  segments: ExposureSegment[]
  algorithm_version: string
  safety_boundary: string
}

export interface DecompressionAssessment {
  id: number
  plan_id: number
  assessment_status: string
  algorithm_version: string
  input_snapshot: InputSnapshot
  compartment_loads: CompartmentCurve[]
  risk_flags: RiskFlag[]
  highest_risk_band: RiskBand
  comparative_score: number
  assumptions: ModelAssumptions
  created_at: string
  reviewed_at?: string
  safety_disclaimer: string
}

export interface AssessmentComparison {
  left: DecompressionAssessment
  right: DecompressionAssessment
  score_delta: number
  flag_delta: number
  summary: string[]
  disclaimer: string
}
