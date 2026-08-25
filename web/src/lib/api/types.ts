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

export type ProjectStatus = 'backlog' | 'planned' | 'in_progress' | 'paused' | 'completed' | 'canceled'

export interface Project {
  id: string
  name: string
  description: string | null
  status: ProjectStatus
  priority: number
  leadId: string | null
  leadName: string | null
  targetDate: string | null
  issueCount: number
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

export interface Attachment {
  id: string
  filename: string
  contentType: string
  sizeBytes: number
  uploadedBy: string
  uploadedByName: string
  createdAt: string
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

export interface ViewFilters {
  status?: string
  assigneeId?: string // may be the literal 'me', resolved client-side at apply-time
  projectId?: string
  labelId?: string
  priority?: number
}

export type ViewGroupBy = 'status' | 'assignee' | 'project' | 'priority' | 'none'
export type ViewSortBy = 'createdAt' | 'updatedAt' | 'priority' | 'dueDate'
export type ViewSortDir = 'asc' | 'desc'

export interface ViewDefinition {
  filters: ViewFilters
  groupBy: ViewGroupBy
  sortBy: ViewSortBy
  sortDir: ViewSortDir
}

export interface View {
  id: string
  ownerId: string
  name: string
  definition: ViewDefinition
  shared: boolean
  createdAt: string
}

export type WorkLogSource = 'human' | 'agent'

export interface WorkLog {
  id: string
  projectId: string
  projectName: string
  authorId: string
  authorName: string
  source: WorkLogSource
  title: string
  body: string
  createdAt: string
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
