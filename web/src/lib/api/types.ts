export interface User {
  id: string
  name: string
  email: string
  role: string
}

export interface Team {
  id: string
  name: string
  key: string
}

export interface Label {
  id: string
  name: string
  color: string
}

export interface Project {
  id: string
  name: string
  description: string | null
  status: string
  progress: number
}

export interface Cycle {
  id: string
  number: number
  startDate: string | null
  endDate: string | null
  active: boolean
  done: number
  total: number
  progress: number
}

export interface ApiKey {
  id: string
  name: string
  scopes: string[]
  createdAt: string
  lastUsedAt: string | null
}

export interface NewApiKey extends ApiKey {
  key: string // plaintext, present only on creation
}

export type IssueStatusValue = 'backlog' | 'todo' | 'in_progress' | 'done' | 'canceled'

export interface Issue {
  id: string
  identifier: string
  title: string
  description: string
  status: IssueStatusValue
  priority: number
  assigneeId: string | null
  assigneeName: string | null
  projectId: string | null
  projectName: string | null
  cycleId: string | null
  label: { id: string; name: string; color: string } | null
  createdAt: string
  updatedAt: string
}
