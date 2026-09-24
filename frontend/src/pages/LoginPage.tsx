import { useState, type FormEvent } from 'react'
import { Alert, Button, CircularProgress, TextField } from '@mui/material'
import { Anchor, ArrowRight, CircleGauge, ShieldCheck } from 'lucide-react'
import { Navigate, useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/auth'

export function LoginPage() {
  const [username, setUsername] = useState('planner')
  const [password, setPassword] = useState('planner123')
  const { user, login, loading, error } = useAuthStore()
  const navigate = useNavigate()
  if (user) return <Navigate to="/plans" replace />
  const submit = async (event: FormEvent) => {
    event.preventDefault()
    try { await login(username, password); navigate('/plans') } catch { /* store renders the API error */ }
  }
  return (
    <main className="login-screen">
      <section className="login-identity">
        <div className="login-mark"><Anchor size={25} /><span>CDC / TRAINING CONTROL</span></div>
        <div className="depth-ruler" aria-hidden="true"><span>00</span><span>15</span><span>30</span><span>45</span><span>60 m</span></div>
        <div className="login-title"><span>OFFLINE EXPOSURE MODEL</span><h1>Dive Exposure<br />Review Control</h1><p>Traceable planning assumptions, deterministic compartment loads, and explicit supervisor decisions.</p></div>
        <div className="boundary-statement"><ShieldCheck size={21} /><span>Training comparison only<br /><small>No live equipment connection or executable decompression instruction</small></span></div>
      </section>
      <section className="login-form-wrap">
        <form className="login-form" onSubmit={submit}>
          <CircleGauge size={28} />
          <div><span className="eyebrow">AUTHORIZED WORKSTATION</span><h2>Open review session</h2></div>
          {error && <Alert severity="error">{error}</Alert>}
          <TextField label="Username" value={username} onChange={(event) => setUsername(event.target.value)} autoComplete="username" required />
          <TextField label="Password" type="password" value={password} onChange={(event) => setPassword(event.target.value)} autoComplete="current-password" required />
          <Button type="submit" variant="contained" endIcon={loading ? <CircularProgress size={16} color="inherit" /> : <ArrowRight size={17} />} disabled={loading}>Enter control room</Button>
          <div className="demo-accounts"><span>Planner</span><code>planner / planner123</code><span>Supervisor</span><code>supervisor / supervisor123</code></div>
        </form>
      </section>
    </main>
  )
}
