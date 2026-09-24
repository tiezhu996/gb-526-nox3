import type { Envelope } from '@/types/common'

const API_BASE = '/api/v1'

export class ApiError extends Error {
  constructor(public status: number, public code: string, message: string, public requestId: string) {
    super(message)
  }
}

export async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const token = localStorage.getItem('dive-control-token')
  const headers = new Headers(init.headers)
  if (init.body && !headers.has('Content-Type')) headers.set('Content-Type', 'application/json')
  if (token) headers.set('Authorization', `Bearer ${token}`)
  const response = await fetch(`${API_BASE}${path}`, { ...init, headers })
  const payload = (await response.json().catch(() => ({ request_id: response.headers.get('X-Request-ID') ?? '' }))) as Envelope<T>
  if (!response.ok || payload.error) {
    throw new ApiError(response.status, payload.error?.code ?? 'HTTP_ERROR', payload.error?.message ?? `Request failed (${response.status})`, payload.request_id)
  }
  return payload.data as T
}

export const errorMessage = (error: unknown): string => error instanceof Error ? error.message : 'Unexpected request failure'
