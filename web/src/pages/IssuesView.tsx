import { useState } from 'react'
import { Button, Text } from '@mantine/core'
import { IconPlus } from '@tabler/icons-react'
import { useOutletContext } from 'react-router-dom'
import type { Team } from '../lib/api/types'
import { useIssues } from '../lib/api/hooks'
import { IssueRow } from '../components/IssueRow'
import { StatusDot } from '../components/StatusDot'
import { IssueDetailPanel } from '../components/IssueDetailPanel'
import { NewIssueModal } from '../components/NewIssueModal'
import { STATUS_META, STATUS_ORDER } from '../theme'

export function IssuesView() {
  const { team } = useOutletContext<{ team: Team | undefined }>()
  const { data: issues } = useIssues(team ? { teamId: team.id } : undefined)
  const [selectedIssueId, setSelectedIssueId] = useState<string | null>(null)
  const [newIssueOpen, setNewIssueOpen] = useState(false)

  const groups = STATUS_ORDER.map((status) => ({
    status,
    meta: STATUS_META[status],
    issues: (issues ?? []).filter((i) => i.status === status),
  })).filter((g) => g.issues.length > 0)

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div style={{ height: 60, minHeight: 60, display: 'flex', alignItems: 'center', gap: 12, padding: '0 20px', borderBottom: '1px solid #1d1e21' }}>
        <Text fw={600} size="md" c="dark.0">
          Issues
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

      <div style={{ flex: 1, overflowY: 'auto', paddingBottom: 40 }}>
        {groups.map((group) => (
          <div key={group.status}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '12px 20px 8px 20px', position: 'sticky', top: 0, background: 'var(--mantine-color-dark-7)' }}>
              <StatusDot status={group.status} />
              <Text size="md" fw={600} c="dark.2">
                {group.meta.label}
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
    </div>
  )
}
