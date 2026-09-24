import { AlertTriangle, ShieldCheck } from 'lucide-react'

export function SafetyBoundaryBanner() {
  return (
    <section className="safety-boundary" aria-label="Training use boundary">
      <div className="safety-icon"><ShieldCheck size={20} aria-hidden="true" /></div>
      <div>
        <strong>TRAINING DECISION SUPPORT</strong>
        <span>Offline comparison only. Not medical advice, a certified dive table, live life support, or executable decompression instruction.</span>
      </div>
      <AlertTriangle size={18} aria-label="Human supervisor approval required" />
    </section>
  )
}
