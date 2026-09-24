import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { Alert, Button, CircularProgress, IconButton, MenuItem, TextField, Tooltip } from '@mui/material'
import { ArrowDown, ArrowUp, Beaker, CirclePlus, Play, Route, Rows3 } from 'lucide-react'
import { PageHeader } from '@/components/common/PageHeader'
import { PlanStatusBadge } from '@/components/common/PlanStatusBadge'
import { useAuth } from '@/hooks/useAuth'
import { useAssessmentStore } from '@/stores/assessment'
import { useDiverStore } from '@/stores/diver'
import { usePlanStore } from '@/stores/plan'
import { useSegmentStore } from '@/stores/segment'
import type { CreateDivePlan } from '@/types/plan'
import type { CreateExposureSegment, SegmentType } from '@/types/segment'

const planInitial: CreateDivePlan = { plan_code: '', diver_profile_id: 0, worksite_pressure_bar: 1, breathing_mix: { o2: .21, he: 0, n2: .79 }, planned_at: new Date(Date.now() + 86_400_000).toISOString().slice(0, 16) }
const segmentInitial = { depth_m: 0, duration_min: 5, ascent_rate_mmin: 0, gas_mix: { o2: .21, he: 0, n2: .79 }, segment_type: 'bottom' as SegmentType, notes: '' }

export function PlansPage() {
  const { isPlanner } = useAuth()
  const plans = usePlanStore()
  const divers = useDiverStore()
  const segments = useSegmentStore()
  const runAssessment = useAssessmentStore((state) => state.run)
  const [planForm, setPlanForm] = useState(planInitial)
  const [segmentForm, setSegmentForm] = useState(segmentInitial)
  const [formMode, setFormMode] = useState<'plan' | 'segment' | null>(null)
  const [busy, setBusy] = useState(false)
  const [notice, setNotice] = useState<string | null>(null)
  const [localError, setLocalError] = useState<string | null>(null)
  useEffect(() => { void plans.load(); void divers.load() }, [plans.load, divers.load])
  useEffect(() => { if (plans.selected) void segments.load(plans.selected.id) }, [plans.selected?.id])
  const selected = plans.selected
  const choosePlan = async (id: number) => { await plans.select(id); await segments.load(id); setNotice(null) }
  const createPlan = async (event: FormEvent) => {
    event.preventDefault(); setBusy(true); setLocalError(null)
    try {
      const input = { ...planForm, planned_at: new Date(planForm.planned_at).toISOString() }
      await plans.create(input); setPlanForm(planInitial); setFormMode(null); setNotice('Draft plan created with input version 1.')
    } catch (error) { setLocalError(error instanceof Error ? error.message : 'Plan creation failed') }
    finally { setBusy(false) }
  }
  const createSegment = async (event: FormEvent) => {
    event.preventDefault(); if (!selected) return
    setBusy(true); setLocalError(null)
    const input: CreateExposureSegment = { ...segmentForm, plan_version: selected.version, sequence_no: segments.items.length + 1 }
    try {
      await segments.create(selected.id, input); await plans.select(selected.id); setSegmentForm({ ...segmentInitial, gas_mix: selected.breathing_mix }); setFormMode(null); setNotice(`Segment ${input.sequence_no} appended and input version advanced.`)
    } catch (error) { setLocalError(error instanceof Error ? error.message : 'Segment creation failed') }
    finally { setBusy(false) }
  }
  const move = async (index: number, direction: -1 | 1) => {
    if (!selected) return
    const target = index + direction
    if (target < 0 || target >= segments.items.length) return
    const ids = segments.items.map((item) => item.id); [ids[index], ids[target]] = [ids[target], ids[index]]
    setBusy(true); setLocalError(null)
    try { await segments.reorder(selected.id, ids, selected.version); await plans.select(selected.id); setNotice('Segment order and plan input version updated.') }
    catch (error) { setLocalError(error instanceof Error ? error.message : 'Reorder failed') }
    finally { setBusy(false) }
  }
  const model = async () => {
    if (!selected) return
    setBusy(true); setLocalError(null); setNotice(null)
    try { const result = await runAssessment(selected.id, selected.version); await plans.load(); setNotice(`Immutable assessment #${result.id} created with comparative index ${result.comparative_score.toFixed(1)}.`) }
    catch (error) { setLocalError(error instanceof Error ? error.message : 'Model run failed') }
    finally { setBusy(false) }
  }
  const mixTotal = useMemo(() => planForm.breathing_mix.o2 + planForm.breathing_mix.he, [planForm.breathing_mix])
  return (
    <div className="page">
      <PageHeader eyebrow="VERSIONED EXPOSURE INPUT" title="Plan assembly" detail="Build ordered depth and gas assumptions before deterministic offline modeling." actions={isPlanner && <Button variant="contained" startIcon={<CirclePlus size={17} />} onClick={() => setFormMode(formMode === 'plan' ? null : 'plan')}>New plan</Button>} />
      {(plans.error || segments.error || localError) && <Alert severity="error">{plans.error ?? segments.error ?? localError}</Alert>}
      {notice && <Alert severity="success" onClose={() => setNotice(null)}>{notice}</Alert>}
      {formMode === 'plan' && <form className="inline-form" onSubmit={createPlan}>
        <div className="section-title"><Route size={19} /><div><strong>New versioned plan</strong><span>Begins in draft</span></div></div>
        <TextField label="Plan code" value={planForm.plan_code} onChange={(event) => setPlanForm({ ...planForm, plan_code: event.target.value.toUpperCase() })} required />
        <TextField select label="Diver profile" value={planForm.diver_profile_id || ''} onChange={(event) => setPlanForm({ ...planForm, diver_profile_id: Number(event.target.value) })} required>{divers.items.map((profile) => <MenuItem key={profile.id} value={profile.id}>{profile.profile_code} · {profile.display_name}</MenuItem>)}</TextField>
        <TextField label="Worksite pressure (bar)" type="number" inputProps={{ step: .01, min: .7, max: 1.5 }} value={planForm.worksite_pressure_bar} onChange={(event) => setPlanForm({ ...planForm, worksite_pressure_bar: Number(event.target.value) })} />
        <TextField label="O2 fraction" type="number" inputProps={{ step: .01, min: .16, max: 1 }} value={planForm.breathing_mix.o2} onChange={(event) => { const o2 = Number(event.target.value); setPlanForm({ ...planForm, breathing_mix: { ...planForm.breathing_mix, o2, n2: Math.max(0, 1 - o2 - planForm.breathing_mix.he) } }) }} />
        <TextField label="He fraction" type="number" inputProps={{ step: .01, min: 0, max: .84 }} value={planForm.breathing_mix.he} onChange={(event) => { const he = Number(event.target.value); setPlanForm({ ...planForm, breathing_mix: { ...planForm.breathing_mix, he, n2: Math.max(0, 1 - planForm.breathing_mix.o2 - he) } }) }} error={mixTotal > 1} helperText={`N2 ${(Math.max(0, 1 - mixTotal) * 100).toFixed(0)}%`} />
        <TextField label="Planned at" type="datetime-local" value={planForm.planned_at} onChange={(event) => setPlanForm({ ...planForm, planned_at: event.target.value })} InputLabelProps={{ shrink: true }} />
        <div className="form-actions"><Button onClick={() => setFormMode(null)}>Cancel</Button><Button type="submit" variant="contained" disabled={busy || mixTotal > 1 || !planForm.diver_profile_id}>Create draft</Button></div>
      </form>}
      <div className="plan-workbench">
        <section className="record-list plan-index">
          <div className="list-heading"><span>{plans.items.length} PLANS</span><span>VERSION</span></div>
          {plans.loading && <div className="loading-state"><CircularProgress size={24} /></div>}
          {plans.items.map((plan) => <button className={`record-row ${selected?.id === plan.id ? 'selected' : ''}`} key={plan.id} onClick={() => void choosePlan(plan.id)}><div><strong>{plan.plan_code}</strong><span>{plan.diver_profile_code}</span></div><div className="record-meta"><PlanStatusBadge status={plan.plan_status} /><small>v{plan.version}</small></div></button>)}
        </section>
        <section className="sequence-board">
          {selected ? <>
            <div className="sequence-head"><div><span className="eyebrow">PLAN INPUT / V{selected.version}</span><h2>{selected.plan_code}</h2><p>{selected.diver_profile_code} · {selected.worksite_pressure_bar.toFixed(2)} bar · O2 {(selected.breathing_mix.o2 * 100).toFixed(0)} / He {(selected.breathing_mix.he * 100).toFixed(0)}</p></div><PlanStatusBadge status={selected.plan_status} /></div>
            <div className="sequence-toolbar"><div><Rows3 size={17} /><span>{segments.items.length} ordered segments</span></div>{isPlanner && selected.plan_status === 'draft' && <div><Button size="small" startIcon={<CirclePlus size={16} />} onClick={() => { setSegmentForm({ ...segmentInitial, gas_mix: selected.breathing_mix }); setFormMode(formMode === 'segment' ? null : 'segment') }}>Add segment</Button><Button size="small" variant="contained" startIcon={<Play size={16} />} onClick={() => void model()} disabled={busy || segments.items.length === 0}>Run model</Button></div>}</div>
            {formMode === 'segment' && <form className="segment-form" onSubmit={createSegment}>
              <span className="sequence-token">{String(segments.items.length + 1).padStart(2, '0')}</span>
              <TextField select label="Type" value={segmentForm.segment_type} onChange={(event) => setSegmentForm({ ...segmentForm, segment_type: event.target.value as SegmentType })}>{['descent', 'bottom', 'transit', 'ascent', 'surface'].map((value) => <MenuItem key={value} value={value}>{value}</MenuItem>)}</TextField>
              <TextField label="Depth m" type="number" inputProps={{ step: .1, min: 0, max: 120 }} value={segmentForm.depth_m} onChange={(event) => setSegmentForm({ ...segmentForm, depth_m: Number(event.target.value) })} />
              <TextField label="Duration min" type="number" inputProps={{ step: .1, min: .1 }} value={segmentForm.duration_min} onChange={(event) => setSegmentForm({ ...segmentForm, duration_min: Number(event.target.value) })} />
              <TextField label="Ascent m/min" type="number" inputProps={{ step: .1, min: 0, max: 18 }} value={segmentForm.ascent_rate_mmin} onChange={(event) => setSegmentForm({ ...segmentForm, ascent_rate_mmin: Number(event.target.value) })} />
              <TextField label="Notes" value={segmentForm.notes} onChange={(event) => setSegmentForm({ ...segmentForm, notes: event.target.value })} />
              <Button type="submit" variant="contained" disabled={busy}>Append</Button>
            </form>}
            <div className="segment-table"><div className="segment-row segment-labels"><span>SEQ</span><span>TYPE / NOTE</span><span>DEPTH</span><span>TIME</span><span>ASCENT</span><span>GAS</span><span>ORDER</span></div>{segments.items.map((segment, index) => <div className="segment-row" key={segment.id}><span className="sequence-token">{String(segment.sequence_no).padStart(2, '0')}</span><div><strong>{segment.segment_type}</strong><small>{segment.notes || 'No note'}</small></div><span>{segment.depth_m.toFixed(1)} m</span><span>{segment.duration_min.toFixed(1)} min</span><span>{segment.ascent_rate_mmin.toFixed(1)}</span><span>O2 {(segment.gas_mix.o2 * 100).toFixed(0)} / He {(segment.gas_mix.he * 100).toFixed(0)}</span><div className="row-actions"><Tooltip title="Move earlier"><span><IconButton size="small" onClick={() => void move(index, -1)} disabled={busy || index === 0 || selected.plan_status !== 'draft'}><ArrowUp size={16} /></IconButton></span></Tooltip><Tooltip title="Move later"><span><IconButton size="small" onClick={() => void move(index, 1)} disabled={busy || index === segments.items.length - 1 || selected.plan_status !== 'draft'}><ArrowDown size={16} /></IconButton></span></Tooltip></div></div>)}{segments.items.length === 0 && <div className="empty-state"><Beaker size={24} />No exposure segments recorded for this draft.</div>}</div>
          </> : <div className="empty-state">Select a plan to inspect its ordered input.</div>}
        </section>
      </div>
    </div>
  )
}
