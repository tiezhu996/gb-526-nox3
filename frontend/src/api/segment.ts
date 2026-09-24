import { request } from './client'
import type { CreateExposureSegment, ExposureSegment } from '@/types/segment'

export const listSegments = (planId: number) => request<ExposureSegment[]>(`/plans/${planId}/segments`)
export const createSegment = (planId: number, input: CreateExposureSegment) => request<ExposureSegment>(`/plans/${planId}/segments`, { method: 'POST', body: JSON.stringify(input) })
export const updateSegment = (id: number, input: Omit<CreateExposureSegment, 'sequence_no'>) => request<ExposureSegment>(`/segments/${id}`, { method: 'PUT', body: JSON.stringify(input) })
export const reorderSegments = (planId: number, orderedIds: number[], version: number) => request<ExposureSegment[]>(`/plans/${planId}/segments/order`, { method: 'PUT', body: JSON.stringify({ ordered_ids: orderedIds, version }) })
