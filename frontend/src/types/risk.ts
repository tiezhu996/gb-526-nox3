export type RiskBand = 'informational' | 'caution' | 'elevated' | 'invalid'

export interface RiskFlag {
  code: string
  band: RiskBand
  message: string
  evidence: string
}
