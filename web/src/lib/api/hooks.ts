import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  apiKeysApi,
  authApi,
  type CreateIssueInput,
  type CreateWorkLogInput,
  cyclesApi,
  documentsApi,
  type IssueFilters,
  issuesApi,
  labelsApi,
  projectMembersApi,
  projectsApi,
  teamsApi,
  type UpdateIssueDetailsInput,
  type UpdateProjectInput,
  usersApi,
  viewsApi,
  type WorkLogFilters,
  workLogsApi,
} from './client'
import type { Issue, ViewDefinition } from './types'

export function useMe() {
  return useQuery({ queryKey: ['me'], queryFn: authApi.me, retry: false })
}

export function useLogin() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ email, password }: { email: string; password: string }) =>
      authApi.login(email, password),
    onSuccess: (user) => qc.setQueryData(['me'], user),
  })
}

export function useLogout() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: authApi.logout,
    onSuccess: () => qc.clear(),
  })
}

export function useTeams() {
  return useQuery({ queryKey: ['teams'], queryFn: teamsApi.list })
}

// v1 is single-workspace: the app always operates on the first (only)
// seeded team, so every view just needs "the current team" rather than a
// picker.
export function useCurrentTeam() {
  const { data: teams, isLoading } = useTeams()
  return { team: teams?.[0], isLoading }
}

export function useRenameTeam() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, name }: { id: string; name: string }) => teamsApi.rename(id, name),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['teams'] }),
  })
}

export function useUsers() {
  return useQuery({ queryKey: ['users'], queryFn: usersApi.list })
}

export function useLabels(teamId: string | undefined) {
  return useQuery({
    queryKey: ['labels', teamId],
    queryFn: () => labelsApi.list(teamId!),
    enabled: !!teamId,
  })
}

export function useProjects(teamId: string | undefined) {
  return useQuery({
    queryKey: ['projects', teamId],
    queryFn: () => projectsApi.list(teamId!),
    enabled: !!teamId,
  })
}

export function useCycles(teamId: string | undefined) {
  return useQuery({
    queryKey: ['cycles', teamId],
    queryFn: () => cyclesApi.list(teamId!),
    enabled: !!teamId,
  })
}

export function useProject(id: string | undefined) {
  return useQuery({
    queryKey: ['project', id],
    queryFn: () => projectsApi.get(id!),
    enabled: !!id,
  })
}

export function useCreateProject() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ teamId, name, description }: { teamId: string; name: string; description: string }) =>
      projectsApi.create(teamId, name, description),
    onSuccess: (_data, { teamId }) => qc.invalidateQueries({ queryKey: ['projects', teamId] }),
  })
}

export function useUpdateProject() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateProjectInput }) => projectsApi.update(id, input),
    onSuccess: (project) => {
      qc.invalidateQueries({ queryKey: ['projects'] })
      qc.setQueryData(['project', project.id], project)
      // A lead change may have upserted a project_members row server-side
      // (internal/api/projects.go's update-project handler) — refresh so
      // the Members list picks it up without a manual reload.
      if (project.leadId) qc.invalidateQueries({ queryKey: ['projectMembers', project.id] })
    },
  })
}

export function useCreateCycle() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ teamId, startDate, endDate }: { teamId: string; startDate: string; endDate: string }) =>
      cyclesApi.create(teamId, startDate, endDate),
    onSuccess: (_data, { teamId }) => qc.invalidateQueries({ queryKey: ['cycles', teamId] }),
  })
}

export function useIssues(filters: IssueFilters | undefined) {
  return useQuery({
    queryKey: ['issues', filters],
    queryFn: () => issuesApi.list(filters!),
    enabled: !!filters?.teamId,
  })
}

export function useIssue(id: string | undefined) {
  return useQuery({
    queryKey: ['issue', id],
    queryFn: () => issuesApi.get(id!),
    enabled: !!id,
  })
}

export function useCreateIssue() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateIssueInput) => issuesApi.create(input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['issues'] }),
  })
}

export function useUpdateIssueStatus() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) =>
      issuesApi.updateStatus(id, status),
    // Optimistic update: the board should snap a card into its new column
    // the instant it's dropped, not after a round-trip. Patch every cached
    // issues list, and roll back if the request fails.
    onMutate: async ({ id, status }) => {
      await qc.cancelQueries({ queryKey: ['issues'] })
      const previous = qc.getQueriesData<Issue[]>({ queryKey: ['issues'] })
      qc.setQueriesData<Issue[]>({ queryKey: ['issues'] }, (old) =>
        old?.map((issue) =>
          issue.id === id ? { ...issue, status: status as Issue['status'] } : issue,
        ),
      )
      return { previous }
    },
    onError: (_err, _vars, context) => {
      context?.previous.forEach(([key, data]) => qc.setQueryData(key, data))
    },
    onSettled: () => qc.invalidateQueries({ queryKey: ['issues'] }),
  })
}

export function useApiKeys() {
  return useQuery({ queryKey: ['apiKeys'], queryFn: apiKeysApi.list })
}

export function useCreateApiKey() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ name, readOnly }: { name: string; readOnly: boolean }) =>
      apiKeysApi.create(name, readOnly),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['apiKeys'] }),
  })
}

export function useDeleteApiKey() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => apiKeysApi.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['apiKeys'] }),
  })
}

export function useProjectMembers(projectId: string | undefined) {
  return useQuery({
    queryKey: ['projectMembers', projectId],
    queryFn: () => projectMembersApi.list(projectId!),
    enabled: !!projectId,
  })
}

export function useAddProjectMember() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ projectId, userId }: { projectId: string; userId: string }) =>
      projectMembersApi.add(projectId, userId),
    onSuccess: (_data, { projectId }) => qc.invalidateQueries({ queryKey: ['projectMembers', projectId] }),
  })
}

export function useRemoveProjectMember() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ projectId, userId }: { projectId: string; userId: string }) =>
      projectMembersApi.remove(projectId, userId),
    onSuccess: (_data, { projectId }) => qc.invalidateQueries({ queryKey: ['projectMembers', projectId] }),
  })
}

export function useProjectDocuments(projectId: string | undefined) {
  return useQuery({
    queryKey: ['projectDocuments', projectId],
    queryFn: () => documentsApi.list(projectId!),
    enabled: !!projectId,
  })
}

export function useUploadDocument() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ projectId, file }: { projectId: string; file: File }) => documentsApi.upload(projectId, file),
    onSuccess: (_data, { projectId }) => qc.invalidateQueries({ queryKey: ['projectDocuments', projectId] }),
  })
}

export function useDeleteDocument() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ attachmentId }: { attachmentId: string; projectId: string }) => documentsApi.delete(attachmentId),
    onSuccess: (_data, { projectId }) => qc.invalidateQueries({ queryKey: ['projectDocuments', projectId] }),
  })
}

export function useViews() {
  return useQuery({ queryKey: ['views'], queryFn: viewsApi.list })
}

export function useCreateView() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ name, definition, shared }: { name: string; definition: ViewDefinition; shared: boolean }) =>
      viewsApi.create(name, definition, shared),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['views'] }),
  })
}

export function useUpdateView() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({
      id,
      name,
      definition,
      shared,
    }: {
      id: string
      name: string
      definition: ViewDefinition
      shared: boolean
    }) => viewsApi.update(id, name, definition, shared),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['views'] }),
  })
}

export function useDeleteView() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => viewsApi.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['views'] }),
  })
}

export function useProjectWorkLogs(projectId: string | undefined) {
  return useQuery({
    queryKey: ['projectWorkLogs', projectId],
    queryFn: () => workLogsApi.listForProject(projectId!),
    enabled: !!projectId,
  })
}

export function useWorkLogs(filters: WorkLogFilters) {
  return useQuery({
    queryKey: ['workLogs', filters],
    queryFn: () => workLogsApi.list(filters),
    placeholderData: keepPreviousData,
  })
}

export function useCreateWorkLog() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ projectId, input }: { projectId: string; input: CreateWorkLogInput }) =>
      workLogsApi.create(projectId, input),
    onSuccess: (_data, { projectId }) => {
      qc.invalidateQueries({ queryKey: ['projectWorkLogs', projectId] })
      qc.invalidateQueries({ queryKey: ['workLogs'] })
    },
  })
}

export function useUpdateIssueDetails() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateIssueDetailsInput }) =>
      issuesApi.updateDetails(id, input),
    onSuccess: (issue) => {
      qc.invalidateQueries({ queryKey: ['issues'] })
      qc.setQueryData(['issue', issue.id], issue)
    },
  })
}
