import { ActionIcon, Group, Select, Stack, Text } from '@mantine/core'
import { IconX } from '@tabler/icons-react'
import {
  useChangeIssueBlocker,
  useIssueLinks,
  useIssues,
  useSetIssueParent,
} from '../lib/api/hooks'
import type { Issue, IssueRef } from '../lib/api/types'
import { StatusDot } from './StatusDot'

function RefRow({ issueRef, onRemove }: { issueRef: IssueRef; onRemove?: () => void }) {
  return (
    <Group gap="xs" wrap="nowrap">
      <StatusDot status={issueRef.status} size={14} />
      <Text size="sm" c="dark.3" style={{ flexShrink: 0 }}>{issueRef.identifier}</Text>
      <Text size="sm" c="dark.1" style={{ flex: 1, minWidth: 0 }} truncate>{issueRef.title}</Text>
      {onRemove && (
        <ActionIcon size="xs" variant="subtle" color="gray" aria-label="Remove" onClick={onRemove}>
          <IconX size={12} />
        </ActionIcon>
      )}
    </Group>
  )
}

export function IssueLinks({ issue, teamId }: { issue: Issue; teamId: string | undefined }) {
  const { data: links } = useIssueLinks(issue.id)
  const { data: allIssues } = useIssues(teamId ? { teamId } : undefined)
  const setParent = useSetIssueParent()
  const changeBlocker = useChangeIssueBlocker()

  const options = (allIssues ?? [])
    .filter((i) => i.id !== issue.id)
    .map((i) => ({ value: i.id, label: `${i.identifier} ${i.title}` }))
  const blockerOptions = options.filter((o) => !links?.blockedBy.some((b) => b.id === o.value))

  return (
    <Stack gap="md">
      <Select
        label="Parent issue"
        placeholder="None"
        searchable
        clearable
        data={options}
        value={issue.parentId}
        onChange={(v) => setParent.mutate({ id: issue.id, parentId: v })}
        error={setParent.error ? setParent.error.message : undefined}
      />

      <Stack gap={6}>
        <Text size="sm" fw={500}>Blocked by{issue.blocked ? ' — currently blocked' : ''}</Text>
        {links?.blockedBy.map((b) => (
          <RefRow
            key={b.id}
            issueRef={b}
            onRemove={() => changeBlocker.mutate({ id: issue.id, blockerId: b.id, remove: true })}
          />
        ))}
        <Select
          placeholder="Add a blocking issue…"
          searchable
          data={blockerOptions}
          value={null}
          onChange={(v) => v && changeBlocker.mutate({ id: issue.id, blockerId: v })}
          error={changeBlocker.error ? changeBlocker.error.message : undefined}
        />
      </Stack>

      {!!links?.blocks.length && (
        <Stack gap={6}>
          <Text size="sm" fw={500}>Blocking</Text>
          {links.blocks.map((b) => <RefRow key={b.id} issueRef={b} />)}
        </Stack>
      )}

      {!!links?.children.length && (
        <Stack gap={6}>
          <Text size="sm" fw={500}>Sub-issues</Text>
          {links.children.map((c) => <RefRow key={c.id} issueRef={c} />)}
        </Stack>
      )}
    </Stack>
  )
}
