import { useState } from 'react'
import {
  ActionIcon,
  Badge,
  Button,
  Checkbox,
  CopyButton,
  Group,
  Modal,
  Stack,
  Table,
  Text,
  TextInput,
} from '@mantine/core'
import { IconCopy, IconCopyCheck, IconTrash } from '@tabler/icons-react'
import { useApiKeys, useCreateApiKey, useDeleteApiKey } from '../lib/api/hooks'

function formatDate(iso: string | null) {
  if (!iso) return 'Never'
  return new Date(iso).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
}

function NewApiKeyModal({ opened, onClose }: { opened: boolean; onClose: () => void }) {
  const [name, setName] = useState('')
  const [readOnly, setReadOnly] = useState(false)
  const [revealedKey, setRevealedKey] = useState<string | null>(null)
  const createKey = useCreateApiKey()

  function reset() {
    setName('')
    setReadOnly(false)
    setRevealedKey(null)
  }

  function handleClose() {
    reset()
    onClose()
  }

  function submit() {
    createKey.mutate(
      { name, readOnly },
      { onSuccess: (created) => setRevealedKey(created.key) },
    )
  }

  return (
    <Modal opened={opened} onClose={handleClose} title={revealedKey ? 'API key created' : 'New API key'} size="md">
      {revealedKey ? (
        <Stack gap="sm">
          <Text size="sm" c="dark.3">
            Copy this key now — it won't be shown again.
          </Text>
          <Group gap="xs" wrap="nowrap">
            <TextInput value={revealedKey} readOnly flex={1} styles={{ input: { fontFamily: 'monospace', fontSize: 13 } }} />
            <CopyButton value={revealedKey}>
              {({ copied, copy }) => (
                <ActionIcon size="lg" variant="light" color={copied ? 'teal' : 'accent'} onClick={copy}>
                  {copied ? <IconCopyCheck size={16} /> : <IconCopy size={16} />}
                </ActionIcon>
              )}
            </CopyButton>
          </Group>
          <Button onClick={handleClose} mt="sm">
            Done
          </Button>
        </Stack>
      ) : (
        <Stack gap="sm">
          <TextInput
            label="Name"
            placeholder="e.g. Claude MCP"
            value={name}
            onChange={(e) => setName(e.currentTarget.value)}
            autoFocus
            required
          />
          <Checkbox
            label="Read-only"
            description="Can list and view issues, but not create or change anything."
            checked={readOnly}
            onChange={(e) => setReadOnly(e.currentTarget.checked)}
          />
          <Group justify="flex-end" mt="sm">
            <Button variant="subtle" onClick={handleClose}>
              Cancel
            </Button>
            <Button onClick={submit} loading={createKey.isPending} disabled={!name.trim()}>
              Create key
            </Button>
          </Group>
        </Stack>
      )}
    </Modal>
  )
}

export function ApiKeysSection() {
  const { data: keys } = useApiKeys()
  const deleteKey = useDeleteApiKey()
  const [modalOpen, setModalOpen] = useState(false)

  return (
    <div>
      <Group justify="space-between" mb="md">
        <div>
          <Text size="lg" fw={600} c="dark.0">
            API Keys
          </Text>
          <Text size="sm" c="dark.3">
            For MCP clients (Claude, Cursor, etc.) and scripts — see the README for how to connect.
          </Text>
        </div>
        <Button size="sm" onClick={() => setModalOpen(true)}>
          New API key
        </Button>
      </Group>

      {keys && keys.length > 0 ? (
        <Table verticalSpacing="sm">
          <Table.Thead>
            <Table.Tr>
              <Table.Th>Name</Table.Th>
              <Table.Th>Scope</Table.Th>
              <Table.Th>Created</Table.Th>
              <Table.Th>Last used</Table.Th>
              <Table.Th />
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {keys.map((key) => (
              <Table.Tr key={key.id}>
                <Table.Td>{key.name}</Table.Td>
                <Table.Td>
                  <Badge size="sm" variant="light" color={key.scopes.includes('write') ? 'accent' : 'gray'}>
                    {key.scopes.includes('write') ? 'Read-write' : 'Read-only'}
                  </Badge>
                </Table.Td>
                <Table.Td>{formatDate(key.createdAt)}</Table.Td>
                <Table.Td>{formatDate(key.lastUsedAt)}</Table.Td>
                <Table.Td>
                  <ActionIcon
                    variant="subtle"
                    color="red"
                    onClick={() => deleteKey.mutate(key.id)}
                    loading={deleteKey.isPending && deleteKey.variables === key.id}
                  >
                    <IconTrash size={16} />
                  </ActionIcon>
                </Table.Td>
              </Table.Tr>
            ))}
          </Table.Tbody>
        </Table>
      ) : (
        <Text size="sm" c="dark.4">
          No API keys yet.
        </Text>
      )}

      <NewApiKeyModal opened={modalOpen} onClose={() => setModalOpen(false)} />
    </div>
  )
}
