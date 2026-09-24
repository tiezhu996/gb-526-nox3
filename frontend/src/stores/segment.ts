import { create } from 'zustand'
import { createSegment, listSegments, reorderSegments, updateSegment } from '@/api/segment'
import { errorMessage } from '@/api/client'
import type { CreateExposureSegment, ExposureSegment } from '@/types/segment'

interface SegmentStore {
  planId: number | null
  items: ExposureSegment[]
  loading: boolean
  error: string | null
  load: (planId: number) => Promise<void>
  create: (planId: number, input: CreateExposureSegment) => Promise<ExposureSegment>
  update: (id: number, input: Omit<CreateExposureSegment, 'sequence_no'>) => Promise<ExposureSegment>
  reorder: (planId: number, ids: number[], version: number) => Promise<void>
}

export const useSegmentStore = create<SegmentStore>((set) => ({
  planId: null, items: [], loading: false, error: null,
  load: async (planId) => {
    set({ planId, loading: true, error: null })
    try { set({ items: await listSegments(planId), loading: false }) }
    catch (error) { set({ error: errorMessage(error), loading: false, items: [] }) }
  },
  create: async (planId, input) => {
    const item = await createSegment(planId, input)
    set((state) => ({ items: [...state.items, item].sort((a, b) => a.sequence_no - b.sequence_no) }))
    return item
  },
  update: async (id, input) => {
    const item = await updateSegment(id, input)
    set((state) => ({ items: state.items.map((current) => current.id === id ? item : current) }))
    return item
  },
  reorder: async (planId, ids, version) => {
    const items = await reorderSegments(planId, ids, version)
    set({ items })
  },
}))
