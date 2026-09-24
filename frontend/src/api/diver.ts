import { request } from './client'
import type { Page } from '@/types/common'
import type { CreateDiverProfile, DiverProfile } from '@/types/diver'
import type { DivePlan } from '@/types/plan'

export const listDivers = () => request<Page<DiverProfile>>('/divers?size=100')
export const createDiver = (input: CreateDiverProfile) => request<DiverProfile>('/divers', { method: 'POST', body: JSON.stringify(input) })
export const updateDiver = (id: number, input: Omit<CreateDiverProfile, 'profile_code'> & { version: number }) => request<DiverProfile>(`/divers/${id}`, { method: 'PUT', body: JSON.stringify(input) })
export const listDiverPlans = (id: number) => request<Page<DivePlan>>(`/divers/${id}/plans?size=100`)
