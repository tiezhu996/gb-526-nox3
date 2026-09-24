import { useEffect, useMemo, useState } from 'react'
import { Alert, Button, TextField } from '@mui/material'
import { FileClock, RefreshCw, Search, ShieldCheck } from 'lucide-react'
import { listAuditEvents } from '@/api/audit'
import { errorMessage } from '@/api/client'
import { PageHeader } from '@/components/common/PageHeader'
import type { AuditEvent } from '@/types/audit'

export function AuditPage() {
  const [items, setItems] = useState<AuditEvent[]>([])
  const [filter, setFilter] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const load = async () => {
    setLoading(true); setError(null)
    try { setItems((await listAuditEvents()).items) }
    catch (caught) { setError(errorMessage(caught)) }
    finally { setLoading(false) }
  }
  useEffect(() => { void load() }, [])
  const visible = useMemo(() => {
    const needle = filter.trim().toLowerCase()
    return needle ? items.filter((item) => [item.actor_username, item.action, item.entity_type, item.request_id].some((value) => value.toLowerCase().includes(needle))) : items
  }, [filter, items])
  return (
    <div className="page">
      <PageHeader eyebrow="APPEND-ONLY TRACE" title="Audit evidence" detail="Profile changes, input versions, model runs, state transitions, and human review actions." actions={<Button startIcon={<RefreshCw size={16} />} onClick={() => void load()} disabled={loading}>Refresh</Button>} />
      {error && <Alert severity="error">{error}</Alert>}
      <div className="audit-tools"><Search size={18} /><TextField placeholder="Filter actor, action, entity, or request ID" value={filter} onChange={(event) => setFilter(event.target.value)} fullWidth /><span>{visible.length} / {items.length}</span></div>
      <section className="audit-ledger">
        <div className="audit-row audit-labels"><span>TIME / REQUEST</span><span>ACTOR</span><span>ACTION</span><span>ENTITY</span><span>BEFORE → AFTER</span></div>
        {visible.map((event) => <div className="audit-row" key={event.id}><div><strong>{new Date(event.created_at).toLocaleString()}</strong><code>{event.request_id}</code></div><div><ShieldCheck size={16} /><strong>{event.actor_username}</strong></div><span>{event.action}</span><span>{event.entity_type} #{event.entity_id}</span><div className="audit-change"><small>{event.before_summary || '∅'}</small><span>→</span><strong>{event.after_summary || '∅'}</strong></div></div>)}
        {!visible.length && <div className="empty-state"><FileClock size={24} />No matching audit events.</div>}
      </section>
    </div>
  )
}
