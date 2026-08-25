import { useState } from 'react'
import { Button, Group, Modal, Stack, Text, TextInput, Textarea } from '@mantine/core'
import { IconPlus } from '@tabler/icons-react'
import { useCreateWorkLog, useProjectWorkLogs } from '../lib/api/hooks'
import { WorkLogEntryCard } from './WorkLogEntryCard'

function AddWorkLogModal({
  opened,
  onClose,
  projectId,
}: {
  opened: boolean
  onClose: () => void
  projectId: string
}) {
  const [title, setTitle] = useState('')
  const [body, setBody] = useState('')
  const createWorkLog = useCreateWorkLog()

  function submit() {
    createWorkLog.mutate(
      { projectId, input: { title, body } },
      {
        onSuccess: () => {
          setTitle('')
          setBody('')
          onClose()
        },
      },
    )
  }

  return (
    <Modal opened={opened} onClose={onClose} title="Add work log entry" size="lg">
      <Stack gap="sm">
        <TextInput
          placeholder="Title"
          value={title}
          onChange={(e) => setTitle(e.currentTarget.value)}
          autoFocus
          required
        />
        <Textarea
          placeholder="What happened? Markdown supported."
          value={body}
          onChange={(e) => setBody(e.currentTarget.value)}
          minRows={6}
          autosize
          required
        />
        <Group justify="flex-end" mt="sm">
          <Button variant="subtle" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={submit} loading={createWorkLog.isPending} disabled={!title.trim() || !body.trim()}>
            Add entry
          </Button>
        </Group>
      </Stack>
    </Modal>
  )
}

export function WorkLogSection({ projectId }: { projectId: string }) {
  const { data: entries } = useProjectWorkLogs(projectId)
  const [addOpen, setAddOpen] = useState(false)

  return (
    <div style={{ padding: '16px 24px' }}>
      <Group justify="space-between" mb="sm">
        <Text size="lg" fw={600} c="dark.0">
          Work Log
        </Text>
        <Button size="sm" variant="light" leftSection={<IconPlus size={14} />} onClick={() => setAddOpen(true)}>
          Add entry
        </Button>
      </Group>

      <Stack gap={0}>
        {(entries ?? []).map((entry) => (
          <WorkLogEntryCard key={entry.id} entry={entry} />
        ))}
        {entries?.length === 0 && (
          <Text size="sm" c="dark.4">
            No work log entries yet.
          </Text>
        )}
      </Stack>

      <AddWorkLogModal opened={addOpen} onClose={() => setAddOpen(false)} projectId={projectId} />
    </div>
  )
}
