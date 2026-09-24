import { create } from 'zustand'
import { archivePlan, createPlan, getPlan, listPlans } from '@/api/plan'
import { errorMessage } from '@/api/client'
import type { CreateDivePlan, DivePlan } from '@/types/plan'

interface PlanStore {
  items: DivePlan[]
  selected: DivePlan | null
  loading: boolean
  error: string | null
  load: () => Promise<void>
  select: (id: number) => Promise<void>
  create: (input: CreateDivePlan) => Promise<DivePlan>
  archive: (id: number, version: number, reason: string) => Promise<DivePlan>
}

export const usePlanStore = create<PlanStore>((set, get) => ({
  items: [], selected: null, loading: false, error: null,
  load: async () => {
    set({ loading: true, error: null })
    try {
      const page = await listPlans()
      const selectedId = get().selected?.id ?? page.items[0]?.id
      set({ items: page.items, loading: false })
      if (selectedId) await get().select(selectedId)
    } catch (error) { set({ error: errorMessage(error), loading: false }) }
  },
  select: async (id) => {
    try { set({ selected: await getPlan(id), error: null }) }
    catch (error) { set({ error: errorMessage(error) }) }
  },
  create: async (input) => {
    const item = await createPlan(input)
    set((state) => ({ items: [item, ...state.items], selected: item }))
    return item
  },
  archive: async (id, version, reason) => {
    const item = await archivePlan(id, version, reason)
    set((state) => ({ items: state.items.map((current) => current.id === id ? item : current), selected: state.selected?.id === id ? item : state.selected }))
    return item
  },
}))
