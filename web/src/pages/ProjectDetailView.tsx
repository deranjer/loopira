import { useEffect, useMemo, useState } from 'react'
import { ActionIcon, Button, Group, Select, Tabs, Text, Textarea, TextInput } from '@mantine/core'
import { IconArrowLeft, IconPlus, IconSearch } from '@tabler/icons-react'
import { useNavigate, useOutletContext, useParams } from 'react-router-dom'
import type { Team } from '../lib/api/types'
import { useIssues, useLabels, useProject, useUpdateProject, useUsers } from '../lib/api/hooks'
import { IssueRow } from '../components/IssueRow'
import { IssueDetailPanel } from '../components/IssueDetailPanel'
import { NewIssueModal } from '../components/NewIssueModal'
import { MembersSection } from '../components/MembersSection'
import { DocumentsSection } from '../components/DocumentsSection'
import { WorkLogSection } from '../components/WorkLogSection'
import { AgentGuideSection } from '../components/AgentGuideSection'
import { PRIORITY_META, PROJECT_STATUS_META, PROJECT_STATUS_ORDER, STATUS_META, STATUS_ORDER } from '../theme'

export function ProjectDetailView() {
  const { id } = useParams<{ id: string }>()
  const { team } = useOutletContext<{ team: Team | undefined }>()
  const navigate = useNavigate()
  const { data: project } = useProject(id)
  const { data: users } = useUsers()
  const { data: labels } = useLabels(team?.id)
  const updateProject = useUpdateProject()
  const { data: issues } = useIssues(team && id ? { teamId: team.id, projectId: id } : undefined)
  const [selectedIssueId, setSelectedIssueId] = useState<string | null>(null)
  const [newIssueOpen, setNewIssueOpen] = useState(false)
  const [issueSearch, setIssueSearch] = useState('')
  const [issueStatus, setIssueStatus] = useState<string | null>(null)
  const [issueAssignee, setIssueAssignee] = useState<string | null>(null)
  const [issueLabel, setIssueLabel] = useState<string | null>(null)
  const [issuePriority, setIssuePriority] = useState<string | null>(null)
  const [issueSort, setIssueSort] = useState('updatedAt:desc')

  const [name, setName] = useState('')
  const [description, setDescription] = useState('')

  useEffect(() => {
    if (project) {
      setName(project.name)
      setDescription(project.description ?? '')
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [project?.id])

  const visibleIssues = useMemo(() => {
    const query = issueSearch.trim().toLocaleLowerCase()
    const [sortBy, sortDirection] = issueSort.split(':')
    return (issues ?? [])
      .filter((issue) => {
        if (
          query &&
          !`${issue.identifier} ${issue.title} ${issue.description}`.toLocaleLowerCase().includes(query)
        ) return false
        if (issueStatus && issue.status !== issueStatus) return false
        if (
          issueAssignee &&
          (issueAssignee === 'unassigned' ? issue.assigneeId !== null : issue.assigneeId !== issueAssignee)
        ) return false
        if (issueLabel && issue.label?.id !== issueLabel) return false
        if (issuePriority !== null && issue.priority !== Number(issuePriority)) return false
        return true
      })
      .sort((a, b) => {
        let comparison = 0
        if (sortBy === 'title') comparison = a.title.localeCompare(b.title)
        else if (sortBy === 'priority') comparison = a.priority - b.priority
        else {
          const dateField = sortBy === 'createdAt' ? 'createdAt' : 'updatedAt'
          comparison = a[dateField].localeCompare(b[dateField])
        }
        return sortDirection === 'asc' ? comparison : -comparison
      })
  }, [issues, issueSearch, issueStatus, issueAssignee, issueLabel, issuePriority, issueSort])

  if (!id || !project) {
    return null
  }

  function save(patch: Partial<{ name: string; description: string; status: string; priority: number; leadId: string | null; targetDate: string | null }>) {
    if (!project) return
    updateProject.mutate({
      id: project.id,
      input: {
        name: patch.name ?? project.name,
        description: patch.description ?? project.description ?? '',
        status: patch.status ?? project.status,
        priority: patch.priority ?? project.priority,
        leadId: patch.leadId !== undefined ? patch.leadId : project.leadId,
        targetDate: patch.targetDate !== undefined ? patch.targetDate : project.targetDate,
      },
    })
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div style={{ height: 60, minHeight: 60, display: 'flex', alignItems: 'center', gap: 12, padding: '0 20px', borderBottom: '1px solid #1d1e21' }}>
        <ActionIcon variant="subtle" onClick={() => navigate('/projects')}>
          <IconArrowLeft size={18} />
        </ActionIcon>
        <Text fw={600} size="md" c="dark.0">
          {project.name}
        </Text>
        <div style={{ marginLeft: 'auto' }}>
          <Button size="sm" leftSection={<IconPlus size={15} />} onClick={() => setNewIssueOpen(true)}>
            New issue
          </Button>
        </div>
      </div>

      <div style={{ flex: 1, overflowY: 'auto', display: 'flex' }}>
        <div style={{ flex: 2, minWidth: 0, borderRight: '1px solid #1d1e21', display: 'flex', flexDirection: 'column' }}>
          <div style={{ padding: '20px 24px', borderBottom: '1px solid #1d1e21', display: 'flex', flexDirection: 'column', gap: 10 }}>
            <TextInput
              value={name}
              onChange={(e) => setName(e.currentTarget.value)}
              onBlur={() => name.trim() && name !== project.name && save({ name })}
              variant="unstyled"
              styles={{ input: { fontSize: 20, fontWeight: 600 } }}
            />
            <Textarea
              placeholder="Add a description..."
              value={description}
              onChange={(e) => setDescription(e.currentTarget.value)}
              onBlur={() => description !== (project.description ?? '') && save({ description })}
              variant="unstyled"
              minRows={2}
              autosize
              styles={{ input: { fontSize: 14, color: 'var(--mantine-color-dark-2)' } }}
            />
          </div>

          <Tabs defaultValue="issues" style={{ display: 'flex', flexDirection: 'column', flex: 1, minHeight: 0 }}>
            <Tabs.List px={12}>
              <Tabs.Tab value="issues">Issues</Tabs.Tab>
              <Tabs.Tab value="worklog">Work Log</Tabs.Tab>
              <Tabs.Tab value="agent-guide">Agent Guide</Tabs.Tab>
            </Tabs.List>

            <Tabs.Panel value="issues" style={{ flex: 1, overflowY: 'auto' }}>
              <Group
                gap={8}
                px={16}
                py={10}
                style={{ borderBottom: '1px solid #1d1e21', position: 'sticky', top: 0, background: '#0e0f11', zIndex: 1 }}
              >
                <TextInput
                  placeholder="Search issues"
                  aria-label="Search project issues"
                  size="xs"
                  leftSection={<IconSearch size={14} />}
                  value={issueSearch}
                  onChange={(event) => setIssueSearch(event.currentTarget.value)}
                  style={{ flex: '1 1 180px' }}
                />
                <Select
                  placeholder="Status"
                  aria-label="Filter by status"
                  size="xs"
                  clearable
                  data={STATUS_ORDER.map((status) => ({ value: status, label: STATUS_META[status].label }))}
                  value={issueStatus}
                  onChange={setIssueStatus}
                  w={125}
                />
                <Select
                  placeholder="Assignee"
                  aria-label="Filter by assignee"
                  size="xs"
                  clearable
                  searchable
                  data={[
                    { value: 'unassigned', label: 'Unassigned' },
                    ...(users ?? []).map((user) => ({ value: user.id, label: user.name })),
                  ]}
                  value={issueAssignee}
                  onChange={setIssueAssignee}
                  w={140}
                />
                <Select
                  placeholder="Label"
                  aria-label="Filter by label"
                  size="xs"
                  clearable
                  searchable
                  data={(labels ?? []).map((label) => ({ value: label.id, label: label.name }))}
                  value={issueLabel}
                  onChange={setIssueLabel}
                  w={125}
                />
                <Select
                  placeholder="Priority"
                  aria-label="Filter by priority"
                  size="xs"
                  clearable
                  data={PRIORITY_META.map((priority, index) => ({
                    value: String(index),
                    label: priority.label,
                  }))}
                  value={issuePriority}
                  onChange={setIssuePriority}
                  w={120}
                />
                <Select
                  aria-label="Sort issues"
                  size="xs"
                  data={[
                    { value: 'updatedAt:desc', label: 'Recently updated' },
                    { value: 'createdAt:desc', label: 'Newest first' },
                    { value: 'createdAt:asc', label: 'Oldest first' },
                    { value: 'priority:asc', label: 'Priority' },
                    { value: 'title:asc', label: 'Title A–Z' },
                  ]}
                  value={issueSort}
                  onChange={(value) => value && setIssueSort(value)}
                  allowDeselect={false}
                  w={150}
                />
                <Text size="xs" c="dark.4" ml="auto">
                  {visibleIssues.length} of {issues?.length ?? 0}
                </Text>
              </Group>
              {visibleIssues.map((issue) => (
                <IssueRow key={issue.id} issue={issue} onClick={() => setSelectedIssueId(issue.id)} />
              ))}
              {visibleIssues.length === 0 && (
                <Text size="sm" c="dark.4" p="lg">
                  {issues?.length === 0 ? 'No issues in this project yet.' : 'No issues match these filters.'}
                </Text>
              )}
            </Tabs.Panel>

            <Tabs.Panel value="worklog" style={{ flex: 1, overflowY: 'auto' }}>
              <WorkLogSection projectId={project.id} />
            </Tabs.Panel>

            <Tabs.Panel value="agent-guide" style={{ flex: 1, overflowY: 'auto' }}>
              <AgentGuideSection projectId={project.id} />
            </Tabs.Panel>
          </Tabs>
        </div>

        <div style={{ flex: 1, minWidth: 280, maxWidth: 340, padding: '20px 24px', display: 'flex', flexDirection: 'column', gap: 20 }}>
          <div>
            <Text size="xs" fw={600} c="dark.4" mb={6} style={{ letterSpacing: '.04em' }}>
              STATUS
            </Text>
            <Select
              data={PROJECT_STATUS_ORDER.map((s) => ({ value: s, label: PROJECT_STATUS_META[s].label }))}
              value={project.status}
              onChange={(v) => v && save({ status: v })}
              allowDeselect={false}
            />
          </div>

          <div>
            <Text size="xs" fw={600} c="dark.4" mb={6} style={{ letterSpacing: '.04em' }}>
              PRIORITY
            </Text>
            <Select
              data={PRIORITY_META.map((p, i) => ({ value: String(i), label: p.label }))}
              value={String(project.priority)}
              onChange={(v) => v !== null && save({ priority: Number(v) })}
              allowDeselect={false}
            />
          </div>

          <div>
            <Text size="xs" fw={600} c="dark.4" mb={6} style={{ letterSpacing: '.04em' }}>
              LEAD
            </Text>
            <Select
              placeholder="No lead"
              data={(users ?? []).map((u) => ({ value: u.id, label: u.name }))}
              value={project.leadId}
              onChange={(v) => save({ leadId: v })}
              clearable
              searchable
            />
          </div>

          <div>
            <Text size="xs" fw={600} c="dark.4" mb={6} style={{ letterSpacing: '.04em' }}>
              TARGET DATE
            </Text>
            <TextInput
              type="date"
              value={project.targetDate ?? ''}
              onChange={(e) => save({ targetDate: e.currentTarget.value || null })}
            />
          </div>

          {project.templateName && (
            <Text size="xs" c="dark.4">
              Stamped from template: {project.templateName}
            </Text>
          )}

          <MembersSection projectId={project.id} />
          <DocumentsSection projectId={project.id} />
        </div>
      </div>

      <IssueDetailPanel issueId={selectedIssueId} teamId={team?.id} onClose={() => setSelectedIssueId(null)} />
      {team && <NewIssueModal opened={newIssueOpen} onClose={() => setNewIssueOpen(false)} teamId={team.id} />}
    </div>
  )
}
