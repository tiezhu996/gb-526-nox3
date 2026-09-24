import { request } from './client'
import type { Page } from '@/types/common'
import type { CreateDivePlan, DivePlan, PlanStatus } from '@/types/plan'

export const listPlans = () => request<Page<DivePlan>>('/plans?size=100')
export const getPlan = (id: number) => request<DivePlan>(`/plans/${id}`)
export const createPlan = (input: CreateDivePlan) => request<DivePlan>('/plans', { method: 'POST', body: JSON.stringify(input) })
export const archivePlan = (id: number, version: number, reason: string) => request<DivePlan>(`/plans/${id}/archive`, {
  method: 'POST', body: JSON.stringify({ target_status: 'archived' satisfies PlanStatus, version, reason }),
})
