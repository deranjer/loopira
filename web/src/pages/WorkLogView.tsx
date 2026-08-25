import { useState } from 'react'
import { Group, Pagination, Select, Text, TextInput } from '@mantine/core'
import { useDebouncedValue } from '@mantine/hooks'
import { IconSearch } from '@tabler/icons-react'
import { useOutletContext } from 'react-router-dom'
import type { Team, WorkLogSource } from '../lib/api/types'
import { useProjects, useUsers, useWorkLogs } from '../lib/api/hooks'
import { WorkLogEntryCard } from '../components/WorkLogEntryCard'

const PAGE_SIZE = 25

export function WorkLogView() {
  const { team } = useOutletContext<{ team: Team | undefined }>()
  const { data: projects } = useProjects(team?.id)
  const { data: users } = useUsers()

  const [projectId, setProjectId] = useState<string | null>(null)
  const [authorId, setAuthorId] = useState<string | null>(null)
  const [source, setSource] = useState<string | null>(null)
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [searchInput, setSearchInput] = useState('')
  const [debouncedSearch] = useDebouncedValue(searchInput, 300)
  const [page, setPage] = useState(1)

  const { data } = useWorkLogs({
    projectId: projectId ?? undefined,
    authorId: authorId ?? undefined,
    source: (source as WorkLogSource) ?? undefined,
    from: from || undefined,
    to: to || undefined,
    search: debouncedSearch || undefined,
    limit: PAGE_SIZE,
    offset: (page - 1) * PAGE_SIZE,
  })

  const total = data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  function updateFilter(setter: (v: string | null) => void, value: string | null) {
    setter(value)
    setPage(1)
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div style={{ height: 60, minHeight: 60, display: 'flex', alignItems: 'center', gap: 12, padding: '0 20px', borderBottom: '1px solid #1d1e21' }}>
        <Text fw={600} size="md" c="dark.0">
          Work Log
        </Text>
        <Text size="sm" c="dark.4">
          {total} {total === 1 ? 'entry' : 'entries'}
        </Text>
      </div>

      <Group gap={8} px={20} py={10} style={{ borderBottom: '1px solid #1d1e21', flexWrap: 'wrap' }}>
        <Select
          placeholder="Project"
          size="xs"
          clearable
          searchable
          data={(projects ?? []).map((p) => ({ value: p.id, label: p.name }))}
          value={projectId}
          onChange={(v) => updateFilter(setProjectId, v)}
          w={160}
        />
        <Select
          placeholder="Author"
          size="xs"
          clearable
          searchable
          data={(users ?? []).map((u) => ({ value: u.id, label: u.name }))}
          value={authorId}
          onChange={(v) => updateFilter(setAuthorId, v)}
          w={150}
        />
        <Select
          placeholder="Source"
          size="xs"
          clearable
          data={[
            { value: 'human', label: 'Human' },
            { value: 'agent', label: 'Agent' },
          ]}
          value={source}
          onChange={(v) => updateFilter(setSource, v)}
          w={120}
        />
        <TextInput
          type="date"
          size="xs"
          value={from}
          onChange={(e) => {
            setFrom(e.currentTarget.value)
            setPage(1)
          }}
          w={140}
        />
        <TextInput
          type="date"
          size="xs"
          value={to}
          onChange={(e) => {
            setTo(e.currentTarget.value)
            setPage(1)
          }}
          w={140}
        />
        <TextInput
          placeholder="Search title and body..."
          size="xs"
          leftSection={<IconSearch size={13} />}
          value={searchInput}
          onChange={(e) => {
            setSearchInput(e.currentTarget.value)
            setPage(1)
          }}
          w={220}
        />
      </Group>

      <div style={{ flex: 1, overflowY: 'auto', padding: '0 20px' }}>
        {(data?.items ?? []).map((entry) => (
          <WorkLogEntryCard key={entry.id} entry={entry} showProject />
        ))}
        {data?.items.length === 0 && (
          <Text size="sm" c="dark.4" p="lg">
            No work log entries match these filters.
          </Text>
        )}
      </div>

      {totalPages > 1 && (
        <Group justify="center" py={16}>
          <Pagination total={totalPages} value={page} onChange={setPage} size="sm" />
        </Group>
      )}
    </div>
  )
}
