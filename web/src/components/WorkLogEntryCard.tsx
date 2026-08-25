import type { ReactNode } from 'react'
import { Badge, Group, Text } from '@mantine/core'
import Markdown, { type Components } from 'react-markdown'
import type { WorkLog } from '../lib/api/types'

function formatDate(iso: string) {
  return new Date(iso).toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  })
}

// A work log entry is a compact list item, not a document — headings from
// the agent/user's markdown shouldn't blow the card up to article size, so
// every heading level renders at the same small, bold weight as body text.
const markdownComponents: Components = {
  h1: ({ children }) => <Text fw={600} size="sm" c="dark.0" component="div" mt={6} mb={2}>{children}</Text>,
  h2: ({ children }) => <Text fw={600} size="sm" c="dark.0" component="div" mt={6} mb={2}>{children}</Text>,
  h3: ({ children }) => <Text fw={600} size="sm" c="dark.0" component="div" mt={6} mb={2}>{children}</Text>,
  h4: ({ children }) => <Text fw={600} size="sm" c="dark.0" component="div" mt={6} mb={2}>{children}</Text>,
  h5: ({ children }) => <Text fw={600} size="sm" c="dark.0" component="div" mt={6} mb={2}>{children}</Text>,
  h6: ({ children }) => <Text fw={600} size="sm" c="dark.0" component="div" mt={6} mb={2}>{children}</Text>,
  p: ({ children }) => <Text size="sm" c="dark.1" component="div" mb={4}>{children}</Text>,
  li: ({ children }: { children?: ReactNode }) => (
    <Text component="li" size="sm" c="dark.1">
      {children}
    </Text>
  ),
}

export function WorkLogEntryCard({ entry, showProject = false }: { entry: WorkLog; showProject?: boolean }) {
  return (
    <div style={{ padding: '14px 0', borderBottom: '1px solid #1d1e21' }}>
      <Group gap={8} mb={4}>
        <Text fw={600} size="sm" c="dark.0">
          {entry.title}
        </Text>
        {entry.source === 'agent' && (
          <Badge size="xs" variant="light" color="accent">
            agent
          </Badge>
        )}
        {showProject && (
          <Badge size="xs" variant="outline" color="gray">
            {entry.projectName}
          </Badge>
        )}
      </Group>
      <div style={{ fontSize: 13.5, lineHeight: 1.5 }}>
        <Markdown components={markdownComponents}>{entry.body}</Markdown>
      </div>
      <Text size="xs" c="dark.4" mt={6}>
        {entry.authorName} · {formatDate(entry.createdAt)}
      </Text>
    </div>
  )
}
