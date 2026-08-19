import { Avatar, Badge, Text } from '@mantine/core'
import type { Issue } from '../lib/api/types'
import { avatarColor } from '../theme'
import { PriorityBars } from './PriorityBars'

export function IssueRow({ issue, onClick }: { issue: Issue; onClick: () => void }) {
  return (
    <div
      onClick={onClick}
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        padding: '11px 20px',
        borderBottom: '1px solid #17181b',
        cursor: 'pointer',
      }}
      onMouseEnter={(e) => (e.currentTarget.style.background = '#131417')}
      onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
    >
      <Text size="sm" c="dark.3" style={{ width: 68, flexShrink: 0 }}>
        {issue.identifier}
      </Text>
      <PriorityBars priority={issue.priority} />
      <Text
        size="md"
        c="dark.1"
        style={{ flex: 1, minWidth: 0, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}
      >
        {issue.title}
      </Text>
      {issue.label && (
        <Badge
          size="md"
          variant="light"
          radius="sm"
          styles={{ root: { backgroundColor: `${issue.label.color}26`, color: issue.label.color, textTransform: 'none' } }}
        >
          {issue.label.name}
        </Badge>
      )}
      {issue.projectName && (
        <Text size="sm" c="dark.3" style={{ flexShrink: 0 }}>
          {issue.projectName}
        </Text>
      )}
      {issue.assigneeName ? (
        <Avatar size={28} radius="xl" color={avatarColor(issue.assigneeName)} style={{ flexShrink: 0 }}>
          {issue.assigneeName
            .split(' ')
            .map((p) => p[0])
            .join('')}
        </Avatar>
      ) : (
        <Avatar size={28} radius="xl" style={{ flexShrink: 0 }} />
      )}
    </div>
  )
}
