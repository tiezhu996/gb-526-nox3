import { create } from 'zustand'
import { approveAssessment, compareAssessments, getAssessment, listAssessments, runAssessment, submitAssessment } from '@/api/assessment'
import { errorMessage } from '@/api/client'
import type { AssessmentComparison, DecompressionAssessment } from '@/types/assessment'

interface AssessmentStore {
  items: DecompressionAssessment[]
  selected: DecompressionAssessment | null
  comparison: AssessmentComparison | null
  loading: boolean
  error: string | null
  load: (planId?: number) => Promise<void>
  select: (id: number) => Promise<void>
  run: (planId: number, version: number) => Promise<DecompressionAssessment>
  submit: (id: number, version: number, reason: string) => Promise<void>
  approve: (id: number, version: number, reason: string) => Promise<void>
  compare: (leftId: number, rightId: number) => Promise<void>
}

export const useAssessmentStore = create<AssessmentStore>((set, get) => ({
  items: [], selected: null, comparison: null, loading: false, error: null,
  load: async (planId) => {
    set({ loading: true, error: null })
    try {
      const page = await listAssessments(planId)
      const selectedId = get().selected?.id ?? page.items[0]?.id
      set({ items: page.items, loading: false })
      if (selectedId && page.items.some((item) => item.id === selectedId)) await get().select(selectedId)
      else set({ selected: page.items[0] ?? null })
    } catch (error) { set({ error: errorMessage(error), loading: false }) }
  },
  select: async (id) => {
    try { set({ selected: await getAssessment(id), error: null, comparison: null }) }
    catch (error) { set({ error: errorMessage(error) }) }
  },
  run: async (planId, version) => {
    const item = await runAssessment(planId, version)
    set((state) => ({ items: [item, ...state.items], selected: item }))
    return item
  },
  submit: async (id, version, reason) => {
    const item = await submitAssessment(id, version, reason)
    set((state) => ({ items: state.items.map((current) => current.id === id ? item : current), selected: item }))
  },
  approve: async (id, version, reason) => {
    const item = await approveAssessment(id, version, reason)
    set((state) => ({ items: state.items.map((current) => current.id === id ? item : current), selected: item }))
  },
  compare: async (leftId, rightId) => {
    try { set({ comparison: await compareAssessments(leftId, rightId), error: null }) }
    catch (error) { set({ error: errorMessage(error) }) }
  },
}))
