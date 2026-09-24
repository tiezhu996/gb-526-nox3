import { ChevronDown, FlaskConical, ShieldX } from 'lucide-react'
import type { ModelAssumptions } from '@/types/assessment'

export function AssumptionPanel({ assumptions }: { assumptions?: ModelAssumptions }) {
  if (!assumptions) return <div className="empty-inline">No model assumptions are available for this selection.</div>
  return (
    <details className="assumption-panel" open>
      <summary><FlaskConical size={18} /> Model assumptions <ChevronDown size={17} className="summary-chevron" /></summary>
      <div className="assumption-grid">
        <div><span>Purpose</span><strong>{assumptions.purpose}</strong></div>
        <div><span>Pressure</span><strong>{assumptions.pressure_conversion}</strong></div>
        <div><span>Kinetics</span><strong>{assumptions.kinetics}</strong></div>
        <div><span>Water vapor</span><strong>{assumptions.water_vapor_bar.toFixed(4)} bar</strong></div>
      </div>
      <div className="prohibited-use"><ShieldX size={17} /><span>{assumptions.prohibited_use.join(' · ')}</span></div>
    </details>
  )
}
