import { useAuthStore } from '@/stores/auth'

export function useAuth() {
  const { user, logout } = useAuthStore()
  return {
    user,
    logout,
    isPlanner: user?.role === 'planner' || user?.role === 'admin',
    isSupervisor: user?.role === 'supervisor' || user?.role === 'admin',
  }
}
