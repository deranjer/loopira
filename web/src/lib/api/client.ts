import type { ApiKey, Cycle, Issue, Label, NewApiKey, Project, Team, User } from './types'

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

export const projectsApi = {
  list: (teamId: string) => request<Project[]>(`/projects?teamId=${teamId}`),
}

export const cyclesApi = {
  list: (teamId: string) => request<Cycle[]>(`/cycles?teamId=${teamId}`),
}

export interface IssueFilters {
  teamId: string
  status?: string
  projectId?: string
  cycleId?: string
}

export interface CreateIssueInput {
  teamId: string
  title: string
  description: string
  priority: number
  assigneeId?: string | null
  projectId?: string | null
}

export interface UpdateIssueDetailsInput {
  title: string
  description: string
  priority: number
  assigneeId?: string | null
  projectId?: string | null
  cycleId?: string | null
}

export const issuesApi = {
  list: (filters: IssueFilters) => {
    const params = new URLSearchParams({ teamId: filters.teamId })
    if (filters.status) params.set('status', filters.status)
    if (filters.projectId) params.set('projectId', filters.projectId)
    if (filters.cycleId) params.set('cycleId', filters.cycleId)
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
