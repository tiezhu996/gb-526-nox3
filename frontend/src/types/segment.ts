import type { GasMix } from './plan'

export type SegmentType = 'descent' | 'bottom' | 'transit' | 'ascent' | 'surface'

export interface ExposureSegment {
  id: number
  plan_id: number
  sequence_no: number
  depth_m: number
  duration_min: number
  ascent_rate_mmin: number
  gas_mix: GasMix
  segment_type: SegmentType
  notes: string
  created_at: string
  updated_at: string
}

export interface CreateExposureSegment {
  plan_version: number
  sequence_no: number
  depth_m: number
  duration_min: number
  ascent_rate_mmin: number
  gas_mix: GasMix
  segment_type: SegmentType
  notes: string
}
