export interface AuditEvent {
  id: number
  request_id: string
  actor_id: number
  actor_username: string
  action: string
  entity_type: string
  entity_id: number
  before_summary: string
  after_summary: string
  created_at: string
}
