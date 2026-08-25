import { useEffect, useMemo, useState } from 'react'
import { Button, Group, Modal, Select, Stack, Switch, Text, TextInput } from '@mantine/core'
import { IconPlus } from '@tabler/icons-react'
import { useOutletContext, useSearchParams } from 'react-router-dom'
import type { Team, ViewDefinition, ViewFilters, ViewGroupBy, ViewSortBy, ViewSortDir } from '../lib/api/types'
import { useCreateView, useIssues, useLabels, useMe, useProjects, useUsers, useViews } from '../lib/api/hooks'
import { IssueRow } from '../components/IssueRow'
import { StatusDot } from '../components/StatusDot'
import { IssueDetailPanel } from '../components/IssueDetailPanel'
import { NewIssueModal } from '../components/NewIssueModal'
import { PRIORITY_META, STATUS_META, STATUS_ORDER } from '../theme'
import type { Issue } from '../lib/api/types'

const EMPTY_DEFINITION: ViewDefinition = { filters: {}, groupBy: 'status', sortBy: 'createdAt', sortDir: 'desc' }

function buildGroups(issues: Issue[], groupBy: ViewGroupBy) {
  if (groupBy === 'none') {
    return issues.length > 0 ? [{ key: 'all', label: 'All issues', issues }] : []
  }
  if (groupBy === 'assignee') {
    const byName = new Map<string, Issue[]>()
    for (const issue of issues) {
      const key = issue.assigneeName ?? 'Unassigned'
      byName.set(key, [...(byName.get(key) ?? []), issue])
    }
    return [...byName.entries()].map(([label, groupIssues]) => ({ key: label, label, issues: groupIssues }))
  }
  if (groupBy === 'project') {
    const byName = new Map<string, Issue[]>()
    for (const issue of issues) {
      const key = issue.projectName ?? 'No project'
      byName.set(key, [...(byName.get(key) ?? []), issue])
    }
    return [...byName.entries()].map(([label, groupIssues]) => ({ key: label, label, issues: groupIssues }))
  }
  if (groupBy === 'priority') {
    return PRIORITY_META.map((meta, priority) => ({
      key: String(priority),
      label: meta.label,
      issues: issues.filter((i) => i.priority === priority),
    })).filter((g) => g.issues.length > 0)
  }
  // status (default)
  return STATUS_ORDER.map((status) => ({
    key: status,
    label: STATUS_META[status].label,
    issues: issues.filter((i) => i.status === status),
  })).filter((g) => g.issues.length > 0)
}

function sortIssues(issues: Issue[], sortBy: ViewSortBy, sortDir: ViewSortDir) {
  const sorted = [...issues].sort((a, b) => {
    let cmp = 0
    if (sortBy === 'priority') cmp = a.priority - b.priority
    else if (sortBy === 'updatedAt') cmp = a.updatedAt.localeCompare(b.updatedAt)
    else if (sortBy === 'createdAt') cmp = a.createdAt.localeCompare(b.createdAt)
    return sortDir === 'asc' ? cmp : -cmp
  })
  return sorted
}

function SaveViewModal({
  opened,
  onClose,
  definition,
}: {
  opened: boolean
  onClose: () => void
  definition: ViewDefinition
}) {
  const [name, setName] = useState('')
  const [shared, setShared] = useState(false)
  const createView = useCreateView()

  function submit() {
    createView.mutate(
      { name, definition, shared },
      {
        onSuccess: () => {
          setName('')
          setShared(false)
          onClose()
        },
      },
    )
  }

  return (
    <Modal opened={opened} onClose={onClose} title="Save view" size="sm">
      <Stack gap="sm">
        <TextInput
          placeholder="View name"
          value={name}
          onChange={(e) => setName(e.currentTarget.value)}
          autoFocus
          required
        />
        <Switch label="Shared with everyone" checked={shared} onChange={(e) => setShared(e.currentTarget.checked)} />
        <Group justify="flex-end" mt="sm">
          <Button variant="subtle" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={submit} loading={createView.isPending} disabled={!name.trim()}>
            Save view
          </Button>
        </Group>
      </Stack>
    </Modal>
  )
}

function IssueFilterBar({
  teamId,
  filters,
  onFiltersChange,
  groupBy,
  onGroupByChange,
  sortBy,
  sortDir,
  onSortChange,
  onSave,
}: {
  teamId: string
  filters: ViewFilters
  onFiltersChange: (f: ViewFilters) => void
  groupBy: ViewGroupBy
  onGroupByChange: (g: ViewGroupBy) => void
  sortBy: ViewSortBy
  sortDir: ViewSortDir
  onSortChange: (sortBy: ViewSortBy, sortDir: ViewSortDir) => void
  onSave: () => void
}) {
  const { data: users } = useUsers()
  const { data: projects } = useProjects(teamId)
  const { data: labels } = useLabels(teamId)

  return (
    <Group gap={8} px={20} py={10} style={{ borderBottom: '1px solid #1d1e21', flexWrap: 'wrap' }}>
      <Select
        placeholder="Status"
        size="xs"
        clearable
        data={STATUS_ORDER.map((s) => ({ value: s, label: STATUS_META[s].label }))}
        value={filters.status ?? null}
        onChange={(v) => onFiltersChange({ ...filters, status: v ?? undefined })}
        w={130}
      />
      <Select
        placeholder="Assignee"
        size="xs"
        clearable
        searchable
        data={[{ value: 'me', label: 'Me' }, ...(users ?? []).map((u) => ({ value: u.id, label: u.name }))]}
        value={filters.assigneeId ?? null}
        onChange={(v) => onFiltersChange({ ...filters, assigneeId: v ?? undefined })}
        w={150}
      />
      <Select
        placeholder="Project"
        size="xs"
        clearable
        searchable
        data={(projects ?? []).map((p) => ({ value: p.id, label: p.name }))}
        value={filters.projectId ?? null}
        onChange={(v) => onFiltersChange({ ...filters, projectId: v ?? undefined })}
        w={150}
      />
      <Select
        placeholder="Label"
        size="xs"
        clearable
        data={(labels ?? []).map((l) => ({ value: l.id, label: l.name }))}
        value={filters.labelId ?? null}
        onChange={(v) => onFiltersChange({ ...filters, labelId: v ?? undefined })}
        w={130}
      />
      <Select
        placeholder="Priority"
        size="xs"
        clearable
        data={PRIORITY_META.map((p, i) => ({ value: String(i), label: p.label }))}
        value={filters.priority !== undefined ? String(filters.priority) : null}
        onChange={(v) => onFiltersChange({ ...filters, priority: v ? Number(v) : undefined })}
        w={130}
      />
      <Select
        size="xs"
        data={[
          { value: 'status', label: 'Group: Status' },
          { value: 'assignee', label: 'Group: Assignee' },
          { value: 'project', label: 'Group: Project' },
          { value: 'priority', label: 'Group: Priority' },
          { value: 'none', label: 'No grouping' },
        ]}
        value={groupBy}
        onChange={(v) => v && onGroupByChange(v as ViewGroupBy)}
        allowDeselect={false}
        w={150}
      />
      <Select
        size="xs"
        data={[
          { value: 'createdAt:desc', label: 'Newest first' },
          { value: 'createdAt:asc', label: 'Oldest first' },
          { value: 'updatedAt:desc', label: 'Recently updated' },
          { value: 'priority:asc', label: 'Priority' },
        ]}
        value={`${sortBy}:${sortDir}`}
        onChange={(v) => {
          if (!v) return
          const [sb, sd] = v.split(':') as [ViewSortBy, ViewSortDir]
          onSortChange(sb, sd)
        }}
        allowDeselect={false}
        w={160}
      />
      <Button size="xs" variant="light" onClick={onSave} ml="auto">
        Save view
      </Button>
    </Group>
  )
}

export function IssuesView({ assigneeId, title = 'Issues' }: { assigneeId?: string; title?: string } = {}) {
  const { team } = useOutletContext<{ team: Team | undefined }>()
  const { data: me } = useMe()
  const [searchParams, setSearchParams] = useSearchParams()
  const { data: views } = useViews()
  const showFilterBar = assigneeId === undefined

  const viewId = showFilterBar ? searchParams.get('viewId') : null
  const appliedView = useMemo(() => views?.find((v) => v.id === viewId), [views, viewId])

  const [filters, setFilters] = useState<ViewFilters>({})
  const [groupBy, setGroupBy] = useState<ViewGroupBy>('status')
  const [sortBy, setSortBy] = useState<ViewSortBy>('createdAt')
  const [sortDir, setSortDir] = useState<ViewSortDir>('desc')
  const [saveViewOpen, setSaveViewOpen] = useState(false)

  useEffect(() => {
    const def = appliedView?.definition ?? EMPTY_DEFINITION
    setFilters(def.filters)
    setGroupBy(def.groupBy)
    setSortBy(def.sortBy)
    setSortDir(def.sortDir)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [appliedView?.id])

  const [selectedIssueId, setSelectedIssueId] = useState<string | null>(null)
  const [newIssueOpen, setNewIssueOpen] = useState(false)

  const resolvedAssigneeId = filters.assigneeId === 'me' ? me?.id : filters.assigneeId

  const { data: issues } = useIssues(
    team
      ? {
          teamId: team.id,
          assigneeId: assigneeId ?? resolvedAssigneeId,
          status: showFilterBar ? filters.status : undefined,
          projectId: showFilterBar ? filters.projectId : undefined,
          labelId: showFilterBar ? filters.labelId : undefined,
          priority: showFilterBar ? filters.priority : undefined,
        }
      : undefined,
  )

  const sortedIssues = sortIssues(issues ?? [], sortBy, sortDir)
  const groups = buildGroups(sortedIssues, groupBy)
  const currentDefinition: ViewDefinition = { filters, groupBy, sortBy, sortDir }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div style={{ height: 60, minHeight: 60, display: 'flex', alignItems: 'center', gap: 12, padding: '0 20px', borderBottom: '1px solid #1d1e21' }}>
        <Text fw={600} size="md" c="dark.0">
          {appliedView?.name ?? title}
        </Text>
        <Text size="sm" c="dark.4">
          {issues?.length ?? 0} issues
        </Text>
        <div style={{ marginLeft: 'auto' }}>
          <Button size="sm" leftSection={<IconPlus size={15} />} onClick={() => setNewIssueOpen(true)}>
            New issue
          </Button>
        </div>
      </div>

      {showFilterBar && team && (
        <IssueFilterBar
          teamId={team.id}
          filters={filters}
          onFiltersChange={(f) => {
            setFilters(f)
            setSearchParams((prev) => {
              prev.delete('viewId')
              return prev
            })
          }}
          groupBy={groupBy}
          onGroupByChange={setGroupBy}
          sortBy={sortBy}
          sortDir={sortDir}
          onSortChange={(sb, sd) => {
            setSortBy(sb)
            setSortDir(sd)
          }}
          onSave={() => setSaveViewOpen(true)}
        />
      )}

      <div style={{ flex: 1, overflowY: 'auto', paddingBottom: 40 }}>
        {groups.map((group) => (
          <div key={group.key}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '12px 20px 8px 20px', position: 'sticky', top: 0, background: 'var(--mantine-color-dark-7)' }}>
              {groupBy === 'status' && <StatusDot status={group.key as keyof typeof STATUS_META} />}
              <Text size="md" fw={600} c="dark.2">
                {group.label}
              </Text>
              <Text size="sm" c="dark.4">
                {group.issues.length}
              </Text>
            </div>
            {group.issues.map((issue) => (
              <IssueRow key={issue.id} issue={issue} onClick={() => setSelectedIssueId(issue.id)} />
            ))}
          </div>
        ))}
      </div>

      <IssueDetailPanel issueId={selectedIssueId} teamId={team?.id} onClose={() => setSelectedIssueId(null)} />
      {team && <NewIssueModal opened={newIssueOpen} onClose={() => setNewIssueOpen(false)} teamId={team.id} />}
      <SaveViewModal opened={saveViewOpen} onClose={() => setSaveViewOpen(false)} definition={currentDefinition} />
    </div>
  )
}
