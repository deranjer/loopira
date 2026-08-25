import type { ApiKey, Attachment, Cycle, Issue, Label, NewApiKey, Project, Team, User, View, ViewDefinition, WorkLog, WorkLogSource } from './types'

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`/api/v1${path}`, {
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    ...init,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new ApiError(res.status, body.detail || body.title || `request failed: ${res.status}`)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

export const authApi = {
  me: () => request<User>('/auth/me'),
  login: (email: string, password: string) =>
    request<User>('/auth/login', { method: 'POST', body: JSON.stringify({ email, password }) }),
  logout: () => request<{ status: string }>('/auth/logout', { method: 'POST' }),
}

export const teamsApi = {
  list: () => request<Team[]>('/teams'),
  rename: (id: string, name: string) =>
    request<Team>(`/teams/${id}`, { method: 'PATCH', body: JSON.stringify({ name }) }),
}

export const usersApi = {
  list: () => request<User[]>('/users'),
}

export const apiKeysApi = {
  list: () => request<ApiKey[]>('/api-keys'),
  create: (name: string, readOnly: boolean) =>
    request<NewApiKey>('/api-keys', { method: 'POST', body: JSON.stringify({ name, readOnly }) }),
  delete: (id: string) => request<{ status: string }>(`/api-keys/${id}`, { method: 'DELETE' }),
}

export const labelsApi = {
  list: (teamId: string) => request<Label[]>(`/labels?teamId=${teamId}`),
}

export interface UpdateProjectInput {
  name: string
  description: string
  status: string
  priority: number
  leadId?: string | null
  targetDate?: string | null
}

export const projectsApi = {
  list: (teamId: string) => request<Project[]>(`/projects?teamId=${teamId}`),
  get: (id: string) => request<Project>(`/projects/${id}`),
  create: (teamId: string, name: string, description: string) =>
    request<Project>('/projects', { method: 'POST', body: JSON.stringify({ teamId, name, description }) }),
  update: (id: string, input: UpdateProjectInput) =>
    request<Project>(`/projects/${id}`, { method: 'PATCH', body: JSON.stringify(input) }),
}

export const projectMembersApi = {
  list: (projectId: string) => request<User[]>(`/projects/${projectId}/members`),
  add: (projectId: string, userId: string) =>
    request<User[]>(`/projects/${projectId}/members`, { method: 'POST', body: JSON.stringify({ userId }) }),
  remove: (projectId: string, userId: string) =>
    request<{ status: string }>(`/projects/${projectId}/members/${userId}`, { method: 'DELETE' }),
}

export const documentsApi = {
  list: (projectId: string) => request<Attachment[]>(`/projects/${projectId}/documents`),
  upload: async (projectId: string, file: File) => {
    const form = new FormData()
    form.append('file', file)
    // Raw fetch, not the JSON-only request() helper — the backend's
    // upload route is a multipart chi handler, not a huma JSON operation.
    const res = await fetch(`/api/v1/projects/${projectId}/documents`, {
      method: 'POST',
      credentials: 'include',
      body: form,
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({}))
      throw new ApiError(res.status, body.detail || `upload failed: ${res.status}`)
    }
    return res.json() as Promise<Attachment>
  },
  delete: (attachmentId: string) => request<{ status: string }>(`/attachments/${attachmentId}`, { method: 'DELETE' }),
  downloadUrl: (attachmentId: string) => `/api/v1/attachments/${attachmentId}/download`,
}

export const cyclesApi = {
  list: (teamId: string) => request<Cycle[]>(`/cycles?teamId=${teamId}`),
  create: (teamId: string, startDate: string, endDate: string) =>
    request<Cycle>('/cycles', { method: 'POST', body: JSON.stringify({ teamId, startDate, endDate }) }),
}

export interface IssueFilters {
  teamId: string
  status?: string
  projectId?: string
  cycleId?: string
  assigneeId?: string
  priority?: number
  labelId?: string
}

export interface CreateIssueInput {
  teamId: string
  title: string
  description: string
  priority: number
  assigneeId?: string | null
  projectId?: string | null
  labelId?: string | null
}

export interface UpdateIssueDetailsInput {
  title: string
  description: string
  priority: number
  assigneeId?: string | null
  projectId?: string | null
  cycleId?: string | null
  labelId?: string | null
}

export const viewsApi = {
  list: () => request<View[]>('/views'),
  create: (name: string, definition: ViewDefinition, shared: boolean) =>
    request<View>('/views', { method: 'POST', body: JSON.stringify({ name, definition, shared }) }),
  update: (id: string, name: string, definition: ViewDefinition, shared: boolean) =>
    request<View>(`/views/${id}`, { method: 'PATCH', body: JSON.stringify({ name, definition, shared }) }),
  delete: (id: string) => request<{ status: string }>(`/views/${id}`, { method: 'DELETE' }),
}

export interface WorkLogFilters {
  projectId?: string
  authorId?: string
  source?: WorkLogSource
  search?: string
  from?: string
  to?: string
  limit?: number
  offset?: number
}

export interface WorkLogListResult {
  items: WorkLog[]
  total: number
}

export interface CreateWorkLogInput {
  title: string
  body: string
}

export const workLogsApi = {
  listForProject: (projectId: string) => request<WorkLog[]>(`/projects/${projectId}/worklogs`),
  create: (projectId: string, input: CreateWorkLogInput) =>
    request<WorkLog>(`/projects/${projectId}/worklogs`, { method: 'POST', body: JSON.stringify(input) }),
  list: (filters: WorkLogFilters) => {
    const params = new URLSearchParams()
    if (filters.projectId) params.set('projectId', filters.projectId)
    if (filters.authorId) params.set('authorId', filters.authorId)
    if (filters.source) params.set('source', filters.source)
    if (filters.search) params.set('search', filters.search)
    if (filters.from) params.set('from', filters.from)
    if (filters.to) params.set('to', filters.to)
    params.set('limit', String(filters.limit ?? 25))
    params.set('offset', String(filters.offset ?? 0))
    return request<WorkLogListResult>(`/worklogs?${params.toString()}`)
  },
}

export const issuesApi = {
  list: (filters: IssueFilters) => {
    const params = new URLSearchParams({ teamId: filters.teamId })
    if (filters.status) params.set('status', filters.status)
    if (filters.projectId) params.set('projectId', filters.projectId)
    if (filters.cycleId) params.set('cycleId', filters.cycleId)
    if (filters.assigneeId) params.set('assigneeId', filters.assigneeId)
    if (filters.priority !== undefined) params.set('priority', String(filters.priority))
    if (filters.labelId) params.set('labelId', filters.labelId)
    return request<Issue[]>(`/issues?${params.toString()}`)
  },
  get: (id: string) => request<Issue>(`/issues/${id}`),
  create: (input: CreateIssueInput) =>
    request<Issue>('/issues', { method: 'POST', body: JSON.stringify(input) }),
  updateStatus: (id: string, status: string) =>
    request<Issue>(`/issues/${id}/status`, { method: 'PATCH', body: JSON.stringify({ status }) }),
  updateDetails: (id: string, input: UpdateIssueDetailsInput) =>
    request<Issue>(`/issues/${id}`, { method: 'PATCH', body: JSON.stringify(input) }),
}
