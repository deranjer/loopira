import { useEffect, useState } from 'react'
import { ActionIcon, Badge, Button, Group, Modal, Select, Stack, Tabs, Text, Textarea, TextInput, Tooltip } from '@mantine/core'
import { useClipboard } from '@mantine/hooks'
import { IconCheck, IconCopy, IconPlus, IconRefresh, IconTrash } from '@tabler/icons-react'
import { projectGuideFragmentsApi } from '../lib/api/client'
import {
  useAddProjectGuideFragment,
  useDeleteProjectGuideFragment,
  useProjectGuideFragments,
  useResetProjectGuideFragment,
  useTemplateFragments,
  useUpdateProjectGuideFragment,
} from '../lib/api/hooks'
import type { ProjectGuideFragment } from '../lib/api/types'

function GuideFragmentCard({ projectId, fragment }: { projectId: string; fragment: ProjectGuideFragment }) {
  const updateFragment = useUpdateProjectGuideFragment()
  const deleteFragment = useDeleteProjectGuideFragment()
  const resetFragment = useResetProjectGuideFragment()
  const [name, setName] = useState(fragment.name)
  const [content, setContent] = useState(fragment.content)

  useEffect(() => {
    setName(fragment.name)
    setContent(fragment.content)
  }, [fragment.id, fragment.name, fragment.content])

  function save(patch: Partial<{ name: string; content: string }>) {
    updateFragment.mutate({
      projectId,
      fragmentInstanceId: fragment.id,
      name: patch.name ?? name,
      content: patch.content ?? content,
    })
  }

  return (
    <div style={{ border: '1px solid #1d1e21', borderRadius: 8, padding: '12px 14px', marginBottom: 10 }}>
      <Group justify="space-between" mb={6} wrap="nowrap">
        <Group gap={8} wrap="nowrap" style={{ flex: 1, minWidth: 0 }}>
          <TextInput
            value={name}
            onChange={(e) => setName(e.currentTarget.value)}
            onBlur={() => name !== fragment.name && save({ name })}
            variant="unstyled"
            styles={{ input: { fontSize: 15, fontWeight: 600 } }}
          />
          {fragment.baseVersion != null && (
            <Badge size="xs" variant="light" color="gray">
              based on v{fragment.baseVersion}
            </Badge>
          )}
          {fragment.locallyModified && (
            <Badge size="xs" variant="light" color="accent">
              modified locally
            </Badge>
          )}
        </Group>
        <Group gap={4}>
          {fragment.fragmentId && fragment.locallyModified && (
            <Tooltip label="Reset to base fragment's current content">
              <ActionIcon
                variant="subtle"
                size="sm"
                onClick={() => resetFragment.mutate({ projectId, fragmentInstanceId: fragment.id })}
                loading={resetFragment.isPending && resetFragment.variables?.fragmentInstanceId === fragment.id}
              >
                <IconRefresh size={14} />
              </ActionIcon>
            </Tooltip>
          )}
          <ActionIcon
            variant="subtle"
            size="sm"
            color="red"
            onClick={() => deleteFragment.mutate({ projectId, fragmentInstanceId: fragment.id })}
            loading={deleteFragment.isPending && deleteFragment.variables?.fragmentInstanceId === fragment.id}
          >
            <IconTrash size={14} />
          </ActionIcon>
        </Group>
      </Group>
      <Textarea
        value={content}
        onChange={(e) => setContent(e.currentTarget.value)}
        onBlur={() => content !== fragment.content && save({ content })}
        minRows={3}
        autosize
        styles={{ input: { fontFamily: 'var(--mantine-font-family-monospace)', fontSize: 13 } }}
      />
    </div>
  )
}

function AddGuideFragmentModal({ opened, onClose, projectId }: { opened: boolean; onClose: () => void; projectId: string }) {
  const { data: catalog } = useTemplateFragments()
  const addFragment = useAddProjectGuideFragment()
  const [catalogId, setCatalogId] = useState<string | null>(null)
  const [customName, setCustomName] = useState('')
  const [customContent, setCustomContent] = useState('')

  function reset() {
    setCatalogId(null)
    setCustomName('')
    setCustomContent('')
  }

  return (
    <Modal opened={opened} onClose={onClose} title="Add guide fragment" size="lg">
      <Tabs defaultValue="catalog">
        <Tabs.List>
          <Tabs.Tab value="catalog">From catalog</Tabs.Tab>
          <Tabs.Tab value="custom">Custom</Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel value="catalog" pt="sm">
          <Stack gap="sm">
            <Select
              placeholder="Pick a fragment..."
              data={(catalog ?? []).map((f) => ({ value: f.id, label: f.category ? `${f.name} (${f.category})` : f.name }))}
              value={catalogId}
              onChange={setCatalogId}
              searchable
            />
            <Group justify="flex-end">
              <Button
                disabled={!catalogId}
                loading={addFragment.isPending}
                onClick={() =>
                  addFragment.mutate(
                    { projectId, fragmentId: catalogId! },
                    { onSuccess: () => { reset(); onClose() } },
                  )
                }
              >
                Add
              </Button>
            </Group>
          </Stack>
        </Tabs.Panel>
        <Tabs.Panel value="custom" pt="sm">
          <Stack gap="sm">
            <TextInput placeholder="Name" value={customName} onChange={(e) => setCustomName(e.currentTarget.value)} />
            <Textarea placeholder="Markdown content..." minRows={6} autosize value={customContent} onChange={(e) => setCustomContent(e.currentTarget.value)} />
            <Group justify="flex-end">
              <Button
                disabled={!customName}
                loading={addFragment.isPending}
                onClick={() =>
                  addFragment.mutate(
                    { projectId, name: customName, content: customContent },
                    { onSuccess: () => { reset(); onClose() } },
                  )
                }
              >
                Add
              </Button>
            </Group>
          </Stack>
        </Tabs.Panel>
      </Tabs>
    </Modal>
  )
}

export function AgentGuideSection({ projectId }: { projectId: string }) {
  const { data: fragments } = useProjectGuideFragments(projectId)
  const [addOpen, setAddOpen] = useState(false)
  const clipboard = useClipboard({ timeout: 1500 })
  const agentsMdUrl = `${window.location.origin}${projectGuideFragmentsApi.agentsMdUrl(projectId)}`

  return (
    <div style={{ padding: '16px 24px' }}>
      <Group justify="space-between" mb="xs">
        <Text size="lg" fw={600} c="dark.0">
          Agent Guide
        </Text>
        <Button size="sm" variant="light" leftSection={<IconPlus size={14} />} onClick={() => setAddOpen(true)}>
          Add fragment
        </Button>
      </Group>

      <Group gap={6} mb="md">
        <Text size="xs" c="dark.4">
          Point an AI coding agent at:
        </Text>
        <Text size="xs" c="dark.2" style={{ fontFamily: 'var(--mantine-font-family-monospace)' }}>
          {agentsMdUrl}
        </Text>
        <Tooltip label={clipboard.copied ? 'Copied' : 'Copy URL'}>
          <ActionIcon size="xs" variant="subtle" onClick={() => clipboard.copy(agentsMdUrl)}>
            {clipboard.copied ? <IconCheck size={12} /> : <IconCopy size={12} />}
          </ActionIcon>
        </Tooltip>
      </Group>

      <Stack gap={0}>
        {(fragments ?? []).map((f) => (
          <GuideFragmentCard key={f.id} projectId={projectId} fragment={f} />
        ))}
        {fragments?.length === 0 && (
          <Text size="sm" c="dark.4">
            No guide fragments yet — stamp this project from a template, or add fragments directly.
          </Text>
        )}
      </Stack>

      <AddGuideFragmentModal opened={addOpen} onClose={() => setAddOpen(false)} projectId={projectId} />
    </div>
  )
}
