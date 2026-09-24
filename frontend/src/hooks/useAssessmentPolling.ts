import { useEffect } from 'react'
import { useAssessmentStore } from '@/stores/assessment'

export function useAssessmentPolling(active = true, planId?: number) {
  const load = useAssessmentStore((state) => state.load)
  useEffect(() => {
    if (!active) return
    const timer = window.setInterval(() => { void load(planId) }, 10_000)
    return () => window.clearInterval(timer)
  }, [active, load, planId])
}
