import { useEffect, useMemo, useState, type FormEvent } from 'react'
import { Alert, Button, CircularProgress, MenuItem, TextField } from '@mui/material'
import { CirclePlus, FlaskConical, IdCard, UserRoundCheck } from 'lucide-react'
import { PageHeader } from '@/components/common/PageHeader'
import { PlanStatusBadge } from '@/components/common/PlanStatusBadge'
import { useAuth } from '@/hooks/useAuth'
import { useDiverStore } from '@/stores/diver'
import type { CreateDiverProfile } from '@/types/diver'

const initial: CreateDiverProfile = { profile_code: '', display_name: '', qualification_level: 'commercial', default_o2_fraction: 0.21, default_he_fraction: 0, profile_status: 'active', limits_note: '' }

export function DiversPage() {
  const { isPlanner } = useAuth()
  const { items, relatedPlans, selectedId, loading, error, load, select, create } = useDiverStore()
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState(initial)
  const [saving, setSaving] = useState(false)
  const [localError, setLocalError] = useState<string | null>(null)
  useEffect(() => { void load() }, [load])
  const selected = useMemo(() => items.find((item) => item.id === selectedId) ?? null, [items, selectedId])
  const submit = async (event: FormEvent) => {
    event.preventDefault(); setSaving(true); setLocalError(null)
    try { await create(form); setForm(initial); setShowForm(false) }
    catch (caught) { setLocalError(caught instanceof Error ? caught.message : 'Profile creation failed') }
    finally { setSaving(false) }
  }
  return (
    <div className="page">
      <PageHeader eyebrow="MINIMAL TRAINING RECORDS" title="Diver profiles" detail="Qualification and default gas assumptions without diagnostic medical records." actions={isPlanner && <Button variant="contained" startIcon={<CirclePlus size={17} />} onClick={() => setShowForm((value) => !value)}>New profile</Button>} />
      {(error || localError) && <Alert severity="error">{error ?? localError}</Alert>}
      {showForm && <form className="inline-form" onSubmit={submit}>
        <div className="section-title"><IdCard size={19} /><div><strong>New training profile</strong><span>Only minimum non-medical planning data</span></div></div>
        <TextField label="Profile code" value={form.profile_code} onChange={(event) => setForm({ ...form, profile_code: event.target.value.toUpperCase() })} required />
        <TextField label="Display name" value={form.display_name} onChange={(event) => setForm({ ...form, display_name: event.target.value })} required />
        <TextField select label="Qualification" value={form.qualification_level} onChange={(event) => setForm({ ...form, qualification_level: event.target.value as CreateDiverProfile['qualification_level'] })}>{['trainee', 'commercial', 'advanced', 'supervisor'].map((value) => <MenuItem key={value} value={value}>{value}</MenuItem>)}</TextField>
        <TextField label="O2 fraction" type="number" inputProps={{ step: .01, min: .16, max: 1 }} value={form.default_o2_fraction} onChange={(event) => setForm({ ...form, default_o2_fraction: Number(event.target.value) })} />
        <TextField label="He fraction" type="number" inputProps={{ step: .01, min: 0, max: .84 }} value={form.default_he_fraction} onChange={(event) => setForm({ ...form, default_he_fraction: Number(event.target.value) })} />
        <TextField className="form-wide" label="Training limits note" value={form.limits_note} onChange={(event) => setForm({ ...form, limits_note: event.target.value })} />
        <div className="form-actions"><Button onClick={() => setShowForm(false)}>Cancel</Button><Button type="submit" variant="contained" disabled={saving}>{saving ? 'Saving…' : 'Create profile'}</Button></div>
      </form>}
      <div className="master-detail">
        <section className="record-list" aria-label="Diver profile list">
          <div className="list-heading"><span>{items.length} PROFILES</span><span>STATUS / MIX</span></div>
          {loading && <div className="loading-state"><CircularProgress size={24} /></div>}
          {!loading && items.map((profile) => <button key={profile.id} className={`record-row ${profile.id === selectedId ? 'selected' : ''}`} onClick={() => void select(profile.id)}>
            <div><strong>{profile.profile_code}</strong><span>{profile.display_name}</span></div>
            <div className="record-meta"><span className={`text-status ${profile.profile_status}`}>{profile.profile_status}</span><small>O2 {(profile.default_o2_fraction * 100).toFixed(0)} / He {(profile.default_he_fraction * 100).toFixed(0)}</small></div>
          </button>)}
          {!loading && items.length === 0 && <div className="empty-state">No training profiles recorded.</div>}
        </section>
        <section className="detail-panel">
          {selected ? <>
            <div className="detail-kicker"><UserRoundCheck size={18} /><span>PROFILE {selected.version.toString().padStart(2, '0')}</span></div>
            <h2>{selected.display_name}</h2><p className="muted">{selected.profile_code} · {selected.qualification_level}</p>
            <div className="gas-composition"><FlaskConical size={20} /><div><span>OXYGEN<strong>{(selected.default_o2_fraction * 100).toFixed(0)}%</strong></span><span>HELIUM<strong>{(selected.default_he_fraction * 100).toFixed(0)}%</strong></span><span>NITROGEN<strong>{(selected.default_n2_fraction * 100).toFixed(0)}%</strong></span></div></div>
            <div className="note-block"><span>TRAINING LIMIT NOTE</span><p>{selected.limits_note || 'No additional training limit note recorded.'}</p></div>
            <div className="related-list"><div className="subheading">Associated plans <span>{relatedPlans.length}</span></div>{relatedPlans.map((plan) => <div className="related-row" key={plan.id}><div><strong>{plan.plan_code}</strong><small>{new Date(plan.planned_at).toLocaleDateString()}</small></div><PlanStatusBadge status={plan.plan_status} /></div>)}{relatedPlans.length === 0 && <div className="empty-inline">No associated plans.</div>}</div>
          </> : <div className="empty-state">Select a profile to inspect its planning assumptions.</div>}
        </section>
      </div>
    </div>
  )
}
