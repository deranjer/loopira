import { Avatar, Group, Loader, Stack, Text } from '@mantine/core'
import { IconArrowRight } from '@tabler/icons-react'
import type { IssueHistoryChange, Label, Project, User } from '../lib/api/types'
import { useIssueHistory } from '../lib/api/hooks'
import { PRIORITY_META, STATUS_META, avatarColor } from '../theme'

const FIELD_LABELS: Record<string, string> = {
  title: 'Title',
  description: 'Description',
  status: 'Status',
  priority: 'Priority',
  assignee: 'Assignee',
  project: 'Project',
  cycle: 'Cycle',
  label: 'Label',
}

function formatValue(
  field: string,
  value: string | number | null,
  users: User[],
  projects: Project[],
  labels: Label[],
) {
  if (value === null || value === '') return 'None'
  if (field === 'status' && typeof value === 'string') {
    return STATUS_META[value as keyof typeof STATUS_META]?.label ?? value
  }
  if (field === 'priority' && typeof value === 'number') return PRIORITY_META[value]?.label ?? String(value)
  if (field === 'assignee') return users.find((user) => user.id === value)?.name ?? String(value)
  if (field === 'project') return projects.find((project) => project.id === value)?.name ?? String(value)
  if (field === 'label') return labels.find((label) => label.id === value)?.name ?? String(value)
  const text = String(value)
  return text.length > 160 ? `${text.slice(0, 157)}…` : text
}

function ChangeRow({
  field,
  change,
  users,
  projects,
  labels,
}: {
  field: string
  change: IssueHistoryChange
  users: User[]
  projects: Project[]
  labels: Label[]
}) {
  return (
    <div>
      <Text size="xs" fw={600} c="dark.3" mb={3}>
        {FIELD_LABELS[field] ?? field}
      </Text>
      <Group gap={7} wrap="nowrap" align="flex-start">
        <Text size="sm" c="dark.4" style={{ wordBreak: 'break-word' }}>
          {formatValue(field, change.from, users, projects, labels)}
        </Text>
        <IconArrowRight size={13} style={{ flex: '0 0 auto', marginTop: 4 }} />
        <Text size="sm" c="dark.1" style={{ wordBreak: 'break-word' }}>
          {formatValue(field, change.to, users, projects, labels)}
        </Text>
      </Group>
    </div>
  )
}

export function IssueHistory({
  issueId,
  users = [],
  projects = [],
  labels = [],
}: {
  issueId: string
  users?: User[]
  projects?: Project[]
  labels?: Label[]
}) {
  const { data: history, isLoading } = useIssueHistory(issueId)

  if (isLoading) return <Loader size="sm" />
  if (!history?.length) return <Text size="sm" c="dark.4">No history recorded yet.</Text>

  return (
    <Stack gap={0}>
      {history.map((entry) => {
        const changes = Object.entries(entry.changes ?? {})
        return (
          <div key={entry.id} style={{ display: 'flex', gap: 12, padding: '14px 0', borderBottom: '1px solid #1d1e21' }}>
            <Avatar size={28} radius="xl" color={avatarColor(entry.actorName)}>
              {entry.actorName.slice(0, 1).toUpperCase()}
            </Avatar>
            <div style={{ minWidth: 0, flex: 1 }}>
              <Group gap={6} justify="space-between" align="flex-start">
                <Text size="sm" c="dark.1">
                  <Text span fw={600}>{entry.actorName}</Text>{' '}
                  {entry.action === 'created' ? 'created this issue' : 'updated this issue'}
                </Text>
                <Text size="xs" c="dark.4" style={{ whiteSpace: 'nowrap' }}>
                  {new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(entry.createdAt))}
                </Text>
              </Group>
              {changes.length > 0 && (
                <Stack gap="sm" mt="sm">
                  {changes.map(([field, change]) => (
                    <ChangeRow key={field} field={field} change={change} users={users} projects={projects} labels={labels} />
                  ))}
                </Stack>
              )}
            </div>
          </div>
        )
      })}
    </Stack>
  )
}
