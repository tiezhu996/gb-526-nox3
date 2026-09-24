import { request } from './client'
import type { Page } from '@/types/common'
import type { AssessmentComparison, DecompressionAssessment } from '@/types/assessment'
import type { PlanStatus } from '@/types/plan'

export const listAssessments = (planId?: number) => request<Page<DecompressionAssessment>>(`/assessments?size=100${planId ? `&plan_id=${planId}` : ''}`)
export const getAssessment = (id: number) => request<DecompressionAssessment>(`/assessments/${id}`)
export const runAssessment = (planId: number, planVersion: number) => request<DecompressionAssessment>(`/plans/${planId}/assessments/run`, { method: 'POST', body: JSON.stringify({ plan_version: planVersion }) })
const transition = (id: number, target_status: PlanStatus, version: number, reason: string, action: 'submit' | 'approve') => request<DecompressionAssessment>(`/assessments/${id}/${action}`, { method: 'POST', body: JSON.stringify({ target_status, version, reason }) })
export const submitAssessment = (id: number, version: number, reason: string) => transition(id, 'pending_supervisor_review', version, reason, 'submit')
export const approveAssessment = (id: number, version: number, reason: string) => transition(id, 'approved_for_training', version, reason, 'approve')
export const compareAssessments = (leftId: number, rightId: number) => request<AssessmentComparison>(`/assessments/${leftId}/compare?other_id=${rightId}`)
