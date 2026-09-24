import { create } from 'zustand'
import { login as loginRequest } from '@/api/auth'
import { errorMessage } from '@/api/client'
import type { User } from '@/types/auth'

const storedUser = localStorage.getItem('dive-control-user')

interface AuthStore {
  user: User | null
  loading: boolean
  error: string | null
  login: (username: string, password: string) => Promise<void>
  logout: () => void
}

export const useAuthStore = create<AuthStore>((set) => ({
  user: storedUser ? JSON.parse(storedUser) as User : null,
  loading: false,
  error: null,
  login: async (username, password) => {
    set({ loading: true, error: null })
    try {
      const response = await loginRequest(username, password)
      localStorage.setItem('dive-control-token', response.token)
      localStorage.setItem('dive-control-user', JSON.stringify(response.user))
      set({ user: response.user, loading: false })
    } catch (error) {
      set({ error: errorMessage(error), loading: false })
      throw error
    }
  },
  logout: () => {
    localStorage.removeItem('dive-control-token')
    localStorage.removeItem('dive-control-user')
    set({ user: null, error: null })
  },
}))
