export interface Envelope<T> {
  data?: T
  error?: { code: string; message: string }
  request_id: string
}

export interface Page<T> {
  items: T[]
  total: number
  page: number
  size: number
}

export interface AsyncState {
  loading: boolean
  error: string | null
}
