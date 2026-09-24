import { useEffect, useMemo, useState } from 'react'
import { Alert, Button, MenuItem, TextField } from '@mui/material'
import { ArrowRight, CheckCheck, GitCompareArrows, ShieldAlert, Send } from 'lucide-react'
import { AssumptionPanel } from '@/components/common/AssumptionPanel'
import { PageHeader } from '@/components/common/PageHeader'
import { PlanStatusBadge } from '@/components/common/PlanStatusBadge'
import { getPlan } from '@/api/plan'
import { useAuth } from '@/hooks/useAuth'
import { useAssessmentPolling } from '@/hooks/useAssessmentPolling'
import { useAssessmentStore } from '@/stores/assessment'
import { usePlanStore } from '@/stores/plan'

export function AssessmentsPage() {
  const { isPlanner, isSupervisor } = useAuth()
  const plans = usePlanStore()
  const assessments = useAssessmentStore()
  const [compareId, setCompareId] = useState<number>(0)
  const [reason, setReason] = useState('Reviewed training assumptions and versioned model evidence.')
  const [busy, setBusy] = useState(false)
  const [notice, setNotice] = useState<string | null>(null)
  const [localError, setLocalError] = useState<string | null>(null)
  useEffect(() => { void assessments.load(); void plans.load() }, [assessments.load, plans.load])
  useAssessmentPolling(true)
  const selected = assessments.selected
  const selectedPlan = useMemo(() => plans.items.find((plan) => plan.id === selected?.plan_id), [plans.items, selected?.plan_id])
  const transition = async (kind: 'submit' | 'approve') => {
    if (!selected) return
    setBusy(true); setLocalError(null); setNotice(null)
    try {
      const plan = await getPlan(selected.plan_id)
      if (kind === 'submit') await assessments.submit(selected.id, plan.version, reason)
      else await assessments.approve(selected.id, plan.version, reason)
      await plans.load(); setNotice(kind === 'submit' ? 'Assessment submitted for human supervisor review.' : 'Assessment approved for training comparison; no operational clearance was issued.')
    } catch (error) { setLocalError(error instanceof Error ? error.message : 'Review action failed') }
    finally { setBusy(false) }
  }
  return (
    <div className="page">
      <PageHeader eyebrow="IMMUTABLE MODEL RUNS" title="Assessment review" detail="Compare fixed snapshots, inspect risk evidence, and record explicit human decisions." />
      {(assessments.error || plans.error || localError) && <Alert severity="error">{assessments.error ?? plans.error ?? localError}</Alert>}
      {notice && <Alert severity="success" onClose={() => setNotice(null)}>{notice}</Alert>}
      <div className="assessment-layout">
        <section className="assessment-queue">
          <div className="list-heading"><span>{assessments.items.length} RUNS</span><span>INDEX</span></div>
          {assessments.items.map((item) => <button key={item.id} className={`assessment-row ${selected?.id === item.id ? 'selected' : ''}`} onClick={() => void assessments.select(item.id)}><div><strong>#{item.id} · {plans.items.find((plan) => plan.id === item.plan_id)?.plan_code ?? `Plan ${item.plan_id}`}</strong><span>{item.algorithm_version}</span></div><div><PlanStatusBadge status={item.assessment_status} /><b>{item.comparative_score.toFixed(1)}</b></div></button>)}
          {!assessments.items.length && <div className="empty-state">No immutable assessments recorded.</div>}
        </section>
        <section className="assessment-detail">
          {selected ? <>
            <div className="assessment-title"><div><span className="eyebrow">ASSESSMENT #{selected.id}</span><h2>{selectedPlan?.plan_code ?? `Plan ${selected.plan_id}`}</h2><p>Created {new Date(selected.created_at).toLocaleString()} · input snapshot preserved</p></div><div className="score-dial"><span>COMPARATIVE INDEX</span><strong>{selected.comparative_score.toFixed(1)}</strong><small>{selected.highest_risk_band} · not a safety score</small></div></div>
            <div className="review-bar"><PlanStatusBadge status={selected.assessment_status} /><TextField label="Review reason" value={reason} onChange={(event) => setReason(event.target.value)} fullWidth />{isPlanner && selected.assessment_status === 'modeled' && <Button variant="contained" startIcon={<Send size={17} />} disabled={busy || reason.length < 3} onClick={() => void transition('submit')}>Submit</Button>}{isSupervisor && selected.assessment_status === 'pending_supervisor_review' && <Button variant="contained" color="secondary" startIcon={<CheckCheck size={17} />} disabled={busy || reason.length < 3} onClick={() => void transition('approve')}>Approve training</Button>}</div>
            <section className="risk-section"><div className="subheading">Risk evidence <span>{selected.risk_flags.length}</span></div><div className="risk-list">{selected.risk_flags.map((flag) => <article className={`risk-row risk-${flag.band}`} key={flag.code}><ShieldAlert size={18} /><div><strong>{flag.code.replaceAll('_', ' ')}</strong><p>{flag.message}</p><small>{flag.evidence}</small></div><span>{flag.band}</span></article>)}</div></section>
            <div className="compartment-grid">{selected.compartment_loads.map((curve) => { const last = curve.points.at(-1); return <div key={curve.name}><span>{curve.name}</span><strong>{last?.total_inert_bar.toFixed(3)} bar</strong><small>N2 t½ {curve.n2_half_time_min} · He t½ {curve.he_half_time_min}</small></div> })}</div>
            <AssumptionPanel assumptions={selected.assumptions} />
            <section className="compare-panel"><div className="section-title"><GitCompareArrows size={18} /><div><strong>Compare immutable runs</strong><span>Difference is descriptive, not relative safety</span></div></div><TextField select label="Other assessment" value={compareId || ''} onChange={(event) => setCompareId(Number(event.target.value))} sx={{ minWidth: 220 }}>{assessments.items.filter((item) => item.id !== selected.id).map((item) => <MenuItem key={item.id} value={item.id}>#{item.id} · index {item.comparative_score.toFixed(1)}</MenuItem>)}</TextField><Button startIcon={<ArrowRight size={16} />} disabled={!compareId} onClick={() => void assessments.compare(selected.id, compareId)}>Compare</Button>{assessments.comparison && <div className="comparison-result"><strong>{assessments.comparison.score_delta >= 0 ? '+' : ''}{assessments.comparison.score_delta.toFixed(2)} index</strong><span>{assessments.comparison.flag_delta >= 0 ? '+' : ''}{assessments.comparison.flag_delta} flags</span><p>{assessments.comparison.summary.join(' ')}</p></div>}</section>
            <p className="disclaimer-line">{selected.safety_disclaimer}</p>
          </> : <div className="empty-state">Select an assessment to inspect its immutable evidence.</div>}
        </section>
      </div>
    </div>
  )
}
