export type Role = 'planner' | 'supervisor' | 'admin'

export interface User {
  id: number
  username: string
  display_name: string
  role: Role
}

export interface LoginResponse {
  token: string
  expires_at: string
  user: User
}
