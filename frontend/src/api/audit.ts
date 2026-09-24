import { request } from './client'
import type { Page } from '@/types/common'
import type { AuditEvent } from '@/types/audit'

export const listAuditEvents = () => request<Page<AuditEvent>>('/audit-events?size=100')
