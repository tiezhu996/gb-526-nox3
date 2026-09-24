import { useEffect, useMemo, useState } from 'react'
import { Alert, MenuItem, TextField } from '@mui/material'
import { Activity, Layers3, Timer, Waves } from 'lucide-react'
import { AssumptionPanel } from '@/components/common/AssumptionPanel'
import { ExposureChart } from '@/components/common/ExposureChart'
import { PageHeader } from '@/components/common/PageHeader'
import { PlanStatusBadge } from '@/components/common/PlanStatusBadge'
import { useAssessmentStore } from '@/stores/assessment'
import { usePlanStore } from '@/stores/plan'
import { useSegmentStore } from '@/stores/segment'

export function ExposuresPage() {
  const plans = usePlanStore()
  const segments = useSegmentStore()
  const assessments = useAssessmentStore()
  const [planId, setPlanId] = useState<number>(0)
  useEffect(() => { void plans.load() }, [plans.load])
  useEffect(() => {
    const id = planId || plans.items[0]?.id
    if (!id) return
    if (!planId) setPlanId(id)
    void segments.load(id); void assessments.load(id); void plans.select(id)
  }, [planId, plans.items.length])
  const latest = assessments.items[0] ?? null
  const selectedPlan = plans.items.find((plan) => plan.id === planId) ?? plans.selected
  const elapsed = useMemo(() => segments.items.reduce((total, segment) => total + segment.duration_min, 0), [segments.items])
  const maxDepth = useMemo(() => Math.max(0, ...segments.items.map((segment) => segment.depth_m)), [segments.items])
  return (
    <div className="page">
      <PageHeader eyebrow="DEPTH / TIME EVIDENCE" title="Exposure profile" detail="Input geometry and modeled compartment response share one traceable timeline." actions={<TextField select label="Plan" value={planId || ''} onChange={(event) => setPlanId(Number(event.target.value))} sx={{ minWidth: 230 }}>{plans.items.map((plan) => <MenuItem key={plan.id} value={plan.id}>{plan.plan_code}</MenuItem>)}</TextField>} />
      {(plans.error || segments.error || assessments.error) && <Alert severity="error">{plans.error ?? segments.error ?? assessments.error}</Alert>}
      {selectedPlan && <div className="exposure-strip"><div><Waves size={19} /><span>PLAN<strong>{selectedPlan.plan_code}</strong></span></div><div><Timer size={19} /><span>ELAPSED<strong>{elapsed.toFixed(1)} min</strong></span></div><div><Activity size={19} /><span>MAX DEPTH<strong>{maxDepth.toFixed(1)} m</strong></span></div><div><Layers3 size={19} /><span>MODEL<strong>{latest?.algorithm_version ?? 'NOT RUN'}</strong></span></div><PlanStatusBadge status={selectedPlan.plan_status} /></div>}
      <section className="chart-section">
        <div className="section-heading"><div><span className="eyebrow">ACTUAL API SERIES</span><h2>Exposure and inert-load trace</h2></div><div className="chart-key"><span className="depth-key">Depth input</span><span className="load-key">Compartment loads</span></div></div>
        {segments.items.length ? <ExposureChart segments={segments.items} curves={latest?.compartment_loads} height={410} /> : <div className="empty-state">This plan has no exposure segments.</div>}
      </section>
      <div className="exposure-lower">
        <section className="profile-ledger"><div className="subheading">Segment ledger <span>{segments.items.length}</span></div>{segments.items.map((segment) => <div className="ledger-row" key={segment.id}><span>{String(segment.sequence_no).padStart(2, '0')}</span><strong>{segment.segment_type}</strong><span>{segment.depth_m.toFixed(1)} m</span><span>{segment.duration_min.toFixed(1)} min</span><span>O2 {(segment.gas_mix.o2 * 100).toFixed(0)} / He {(segment.gas_mix.he * 100).toFixed(0)}</span></div>)}</section>
        <section><AssumptionPanel assumptions={latest?.assumptions} /></section>
      </div>
      {latest && <p className="disclaimer-line">{latest.safety_disclaimer}</p>}
    </div>
  )
}
