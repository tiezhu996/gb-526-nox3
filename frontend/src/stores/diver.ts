import { create } from 'zustand'
import { createDiver, listDivers, listDiverPlans, updateDiver } from '@/api/diver'
import { errorMessage } from '@/api/client'
import type { CreateDiverProfile, DiverProfile } from '@/types/diver'
import type { DivePlan } from '@/types/plan'

interface DiverStore {
  items: DiverProfile[]
  relatedPlans: DivePlan[]
  selectedId: number | null
  loading: boolean
  error: string | null
  load: () => Promise<void>
  select: (id: number) => Promise<void>
  create: (input: CreateDiverProfile) => Promise<DiverProfile>
  update: (id: number, input: Omit<CreateDiverProfile, 'profile_code'> & { version: number }) => Promise<DiverProfile>
}

export const useDiverStore = create<DiverStore>((set, get) => ({
  items: [], relatedPlans: [], selectedId: null, loading: false, error: null,
  load: async () => {
    set({ loading: true, error: null })
    try {
      const page = await listDivers()
      const selectedId = get().selectedId ?? page.items[0]?.id ?? null
      set({ items: page.items, selectedId, loading: false })
      if (selectedId) await get().select(selectedId)
    } catch (error) { set({ error: errorMessage(error), loading: false }) }
  },
  select: async (id) => {
    set({ selectedId: id, error: null })
    try { set({ relatedPlans: (await listDiverPlans(id)).items }) }
    catch (error) { set({ error: errorMessage(error), relatedPlans: [] }) }
  },
  create: async (input) => {
    const item = await createDiver(input)
    set((state) => ({ items: [...state.items, item], selectedId: item.id }))
    await get().select(item.id)
    return item
  },
  update: async (id, input) => {
    const item = await updateDiver(id, input)
    set((state) => ({ items: state.items.map((current) => current.id === id ? item : current) }))
    return item
  },
}))
