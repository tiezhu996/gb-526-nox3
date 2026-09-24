import type { ReactNode } from 'react'

export function PageHeader({ eyebrow, title, detail, actions }: { eyebrow: string; title: string; detail: string; actions?: ReactNode }) {
  return (
    <header className="page-header">
      <div><span className="eyebrow">{eyebrow}</span><h1>{title}</h1><p>{detail}</p></div>
      {actions && <div className="page-actions">{actions}</div>}
    </header>
  )
}
