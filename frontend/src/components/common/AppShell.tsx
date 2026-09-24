import { ClipboardCheck, FileClock, Gauge, LogOut, Route, UsersRound, Waves } from 'lucide-react'
import { IconButton, Tooltip } from '@mui/material'
import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { SafetyBoundaryBanner } from './SafetyBoundaryBanner'
import { useAuth } from '@/hooks/useAuth'

const links = [
  { to: '/divers', label: 'Divers', icon: UsersRound },
  { to: '/plans', label: 'Plans', icon: Route },
  { to: '/exposures', label: 'Exposure', icon: Waves },
  { to: '/assessments', label: 'Assessments', icon: ClipboardCheck },
]

export function AppShell() {
  const { user, logout, isSupervisor } = useAuth()
  const navigate = useNavigate()
  const signOut = () => { logout(); navigate('/login') }
  return (
    <div className="app-shell">
      <aside className="side-rail">
        <div className="brand-block"><Gauge size={23} /><div><strong>DIVE EXPOSURE</strong><span>review control / 01</span></div></div>
        <nav aria-label="Primary navigation">
          {links.map(({ to, label, icon: Icon }) => <NavLink key={to} to={to}><Icon size={18} /><span>{label}</span></NavLink>)}
          {isSupervisor && <NavLink to="/audit"><FileClock size={18} /><span>Audit</span></NavLink>}
        </nav>
        <div className="operator-block"><span>{user?.display_name}</span><small>{user?.role}</small><Tooltip title="Sign out"><IconButton aria-label="Sign out" onClick={signOut}><LogOut size={18} /></IconButton></Tooltip></div>
      </aside>
      <div className="workspace">
        <SafetyBoundaryBanner />
        <main><Outlet /></main>
      </div>
    </div>
  )
}
