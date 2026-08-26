import { useEffect, useState } from 'react'
import {
  ActionIcon,
  Badge,
  Button,
  Checkbox,
  Group,
  Modal,
  Select,
  Stack,
  Table,
  Tabs,
  Text,
  Textarea,
  TextInput,
} from '@mantine/core'
import { IconArrowDown, IconArrowUp, IconPlus, IconTrash, IconX } from '@tabler/icons-react'
import Markdown from 'react-markdown'
import {
  useAddTemplateFragment,
  useCreateTemplate,
  useCreateTemplateFragment,
  useDeleteTemplate,
  useDeleteTemplateFragment,
  useFragmentUsage,
  usePushFragmentUpdate,
  useRemoveTemplateFragment,
  useReorderTemplateFragments,
  useTemplate,
  useTemplateFragments,
  useTemplates,
  useUpdateTemplate,
  useUpdateTemplateFragment,
} from '../lib/api/hooks'
import type { TemplateFragment } from '../lib/api/types'

function TemplateComposer({ templateId }: { templateId: string }) {
  const { data: template } = useTemplate(templateId)
  const { data: allFragments } = useTemplateFragments()
  const addFragment = useAddTemplateFragment()
  const removeFragment = useRemoveTemplateFragment()
  const reorderFragments = useReorderTemplateFragments()
  const [pickerValue, setPickerValue] = useState<string | null>(null)

  const fragments = template?.fragments ?? []
  const availableFragments = (allFragments ?? []).filter((f) => !fragments.some((tf) => tf.id === f.id))

  function move(index: number, dir: -1 | 1) {
    const next = [...fragments]
    const target = index + dir
    if (target < 0 || target >= next.length) return
    ;[next[index], next[target]] = [next[target], next[index]]
    reorderFragments.mutate({ id: templateId, fragmentIds: next.map((f) => f.id) })
  }

  return (
    <Stack gap="sm">
      <Text size="sm" fw={600} c="dark.0">
        Composition
      </Text>
      <Stack gap={4}>
        {fragments.map((f, i) => (
          <Group key={f.id} justify="space-between" wrap="nowrap" style={{ padding: '6px 10px', background: '#17181b', borderRadius: 6 }}>
            <Group gap={8} wrap="nowrap">
              <Text size="sm" c="dark.1">
                {f.name}
              </Text>
              {f.category && (
                <Badge size="xs" variant="light" color="gray">
                  {f.category}
                </Badge>
              )}
            </Group>
            <Group gap={2}>
              <ActionIcon variant="subtle" size="sm" disabled={i === 0} onClick={() => move(i, -1)}>
                <IconArrowUp size={14} />
              </ActionIcon>
              <ActionIcon variant="subtle" size="sm" disabled={i === fragments.length - 1} onClick={() => move(i, 1)}>
                <IconArrowDown size={14} />
              </ActionIcon>
              <ActionIcon variant="subtle" size="sm" color="red" onClick={() => removeFragment.mutate({ id: templateId, fragmentId: f.id })}>
                <IconX size={14} />
              </ActionIcon>
            </Group>
          </Group>
        ))}
        {fragments.length === 0 && (
          <Text size="sm" c="dark.4">
            No fragments added yet.
          </Text>
        )}
      </Stack>
      <Group gap={8}>
        <Select
          placeholder="Add a fragment..."
          data={availableFragments.map((f) => ({ value: f.id, label: f.category ? `${f.name} (${f.category})` : f.name }))}
          value={pickerValue}
          onChange={setPickerValue}
          searchable
          style={{ flex: 1 }}
        />
        <Button
          size="sm"
          disabled={!pickerValue}
          loading={addFragment.isPending}
          onClick={() => {
            if (!pickerValue) return
            addFragment.mutate({ id: templateId, fragmentId: pickerValue })
            setPickerValue(null)
          }}
        >
          Add
        </Button>
      </Group>
    </Stack>
  )
}

function TemplateEditorModal({
  opened,
  onClose,
  templateId,
  onCreated,
}: {
  opened: boolean
  onClose: () => void
  templateId: string | null
  onCreated: (id: string) => void
}) {
  const createTemplate = useCreateTemplate()
  const updateTemplate = useUpdateTemplate()
  const { data: template } = useTemplate(templateId ?? undefined)
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')

  function startEditing(id: string, initialName: string, initialDescription: string) {
    setName(initialName)
    setDescription(initialDescription)
    onCreated(id)
  }

  if (!templateId) {
    // Creation step: a template needs to exist before fragments can be
    // attached to it, so this modal first collects name/description, then
    // switches into the composer once the template row exists.
    return (
      <Modal opened={opened} onClose={onClose} title="New template" size="lg">
        <Stack gap="sm">
          <TextInput placeholder="Template name" required autoFocus value={name} onChange={(e) => setName(e.currentTarget.value)} />
          <Textarea placeholder="Add description..." minRows={2} autosize value={description} onChange={(e) => setDescription(e.currentTarget.value)} />
          <Group justify="flex-end" mt="sm">
            <Button variant="subtle" onClick={onClose} type="button">
              Cancel
            </Button>
            <Button
              loading={createTemplate.isPending}
              disabled={!name}
              onClick={() =>
                createTemplate.mutate(
                  { name, description },
                  { onSuccess: (created) => startEditing(created.id, created.name, created.description ?? '') },
                )
              }
            >
              Create &amp; add fragments
            </Button>
          </Group>
        </Stack>
      </Modal>
    )
  }

  return (
    <Modal opened={opened} onClose={onClose} title="Edit template" size="lg">
      <Stack gap="sm">
        <TextInput
          label="Name"
          required
          value={template?.name ?? name}
          onChange={(e) => setName(e.currentTarget.value)}
          onBlur={() => template && name && updateTemplate.mutate({ id: templateId, name, description: template.description ?? description })}
        />
        <Textarea
          label="Description"
          minRows={2}
          autosize
          value={template?.description ?? description}
          onChange={(e) => setDescription(e.currentTarget.value)}
          onBlur={() => template && updateTemplate.mutate({ id: templateId, name: template.name ?? name, description })}
        />
        <TemplateComposer templateId={templateId} />
        <Group justify="flex-end" mt="sm">
          <Button onClick={onClose}>Done</Button>
        </Group>
      </Stack>
    </Modal>
  )
}

function TemplatesTab() {
  const { data: templates } = useTemplates()
  const deleteTemplate = useDeleteTemplate()
  const [editingId, setEditingId] = useState<string | null>(null)
  const [modalOpen, setModalOpen] = useState(false)

  function openCreate() {
    setEditingId(null)
    setModalOpen(true)
  }

  function openEdit(id: string) {
    setEditingId(id)
    setModalOpen(true)
  }

  return (
    <div style={{ padding: '16px 24px' }}>
      <Group justify="space-between" mb="sm">
        <Text size="lg" fw={600} c="dark.0">
          Templates
        </Text>
        <Button size="sm" variant="light" leftSection={<IconPlus size={14} />} onClick={openCreate}>
          New template
        </Button>
      </Group>

      {templates && templates.length > 0 ? (
        <Table verticalSpacing="sm" highlightOnHover>
          <Table.Thead>
            <Table.Tr>
              <Table.Th>Name</Table.Th>
              <Table.Th>Description</Table.Th>
              <Table.Th>Author</Table.Th>
              <Table.Th />
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {templates.map((t) => (
              <Table.Tr key={t.id} style={{ cursor: 'pointer' }} onClick={() => openEdit(t.id)}>
                <Table.Td>
                  <Text size="sm" fw={500} c="dark.0">
                    {t.name}
                  </Text>
                </Table.Td>
                <Table.Td>
                  <Text size="sm" c="dark.3">
                    {t.description || '—'}
                  </Text>
                </Table.Td>
                <Table.Td>
                  <Text size="sm" c="dark.3">{t.authorName ?? '—'}</Text>
                </Table.Td>
                <Table.Td onClick={(e) => e.stopPropagation()}>
                  <ActionIcon
                    variant="subtle"
                    color="red"
                    onClick={() => deleteTemplate.mutate(t.id)}
                    loading={deleteTemplate.isPending && deleteTemplate.variables === t.id}
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
          No templates yet — combine fragments into a template to stamp out new projects with it.
        </Text>
      )}

      <TemplateEditorModal
        opened={modalOpen}
        onClose={() => setModalOpen(false)}
        templateId={editingId}
        onCreated={(id) => setEditingId(id)}
      />
    </div>
  )
}

function FragmentUsagePanel({ fragmentId }: { fragmentId: string }) {
  const { data: usage } = useFragmentUsage(fragmentId)
  const pushUpdate = usePushFragmentUpdate()
  const [selected, setSelected] = useState<string[]>([])

  const unchanged = (usage ?? []).filter((u) => !u.locallyModified)
  const diverged = (usage ?? []).filter((u) => u.locallyModified)

  function toggle(id: string) {
    setSelected((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]))
  }

  if (!usage || usage.length === 0) {
    return (
      <Text size="sm" c="dark.4">
        Not used in any project yet.
      </Text>
    )
  }

  return (
    <Stack gap="sm">
      <Text size="sm" c="dark.3">
        Used in {usage.length} project{usage.length === 1 ? '' : 's'} — {unchanged.length} unchanged, {diverged.length} diverged.
      </Text>
      {unchanged.length > 0 && (
        <Stack gap={4}>
          {unchanged.map((u) => (
            <Checkbox
              key={u.projectGuideFragmentId}
              label={`${u.projectName} (v${u.baseVersion ?? '?'})`}
              checked={selected.includes(u.projectGuideFragmentId)}
              onChange={() => toggle(u.projectGuideFragmentId)}
            />
          ))}
          <Button
            size="xs"
            mt={4}
            disabled={selected.length === 0}
            loading={pushUpdate.isPending}
            onClick={() => pushUpdate.mutate({ id: fragmentId, projectGuideFragmentIds: selected }, { onSuccess: () => setSelected([]) })}
          >
            Push update to {selected.length || ''} selected
          </Button>
        </Stack>
      )}
      {diverged.length > 0 && (
        <Stack gap={4}>
          <Text size="xs" fw={600} c="dark.4">
            DIVERGED (not eligible for push)
          </Text>
          {diverged.map((u) => (
            <Text key={u.projectGuideFragmentId} size="sm" c="dark.3">
              {u.projectName}
            </Text>
          ))}
        </Stack>
      )}
    </Stack>
  )
}

function FragmentEditorModal({
  opened,
  onClose,
  fragment,
}: {
  opened: boolean
  onClose: () => void
  fragment: TemplateFragment | null
}) {
  const createFragment = useCreateTemplateFragment()
  const updateFragment = useUpdateTemplateFragment()
  const [name, setName] = useState('')
  const [category, setCategory] = useState('')
  const [content, setContent] = useState('')

  useEffect(() => {
    if (opened) {
      setName(fragment?.name ?? '')
      setCategory(fragment?.category ?? '')
      setContent(fragment?.content ?? '')
    }
  }, [opened, fragment])

  function save() {
    if (fragment) {
      updateFragment.mutate({ id: fragment.id, name, category, content }, { onSuccess: onClose })
    } else {
      createFragment.mutate({ name, category, content }, { onSuccess: onClose })
    }
  }

  const saving = createFragment.isPending || updateFragment.isPending

  return (
    <Modal opened={opened} onClose={onClose} title={fragment ? 'Edit fragment' : 'New fragment'} size="xl">
      <Stack gap="sm">
        <TextInput label="Name" placeholder="e.g. Go / huma backend conventions" required value={name} onChange={(e) => setName(e.currentTarget.value)} />
        <TextInput
          label="Category"
          placeholder="e.g. engineering, stack"
          description="Freeform — group related fragments however you like"
          value={category}
          onChange={(e) => setCategory(e.currentTarget.value)}
        />
        <Tabs defaultValue="edit">
          <Tabs.List>
            <Tabs.Tab value="edit">Edit</Tabs.Tab>
            <Tabs.Tab value="preview">Preview</Tabs.Tab>
          </Tabs.List>
          <Tabs.Panel value="edit" pt="xs">
            <Textarea
              placeholder="Markdown content..."
              minRows={12}
              autosize
              value={content}
              onChange={(e) => setContent(e.currentTarget.value)}
              styles={{ input: { fontFamily: 'var(--mantine-font-family-monospace)', fontSize: 13 } }}
            />
          </Tabs.Panel>
          <Tabs.Panel value="preview" pt="xs">
            <div style={{ minHeight: 200, fontSize: 13.5, lineHeight: 1.5 }}>
              <Markdown>{content || '*Nothing to preview yet.*'}</Markdown>
            </div>
          </Tabs.Panel>
        </Tabs>

        {fragment && (
          <>
            <Text size="sm" fw={600} c="dark.0" mt="sm">
              Used by
            </Text>
            <FragmentUsagePanel fragmentId={fragment.id} />
          </>
        )}

        <Group justify="flex-end" mt="sm">
          <Button variant="subtle" onClick={onClose} type="button">
            Cancel
          </Button>
          <Button onClick={save} loading={saving} disabled={!name}>
            {fragment ? 'Save' : 'Create fragment'}
          </Button>
        </Group>
      </Stack>
    </Modal>
  )
}

function FragmentsTab() {
  const { data: fragments } = useTemplateFragments()
  const deleteFragment = useDeleteTemplateFragment()
  const [editing, setEditing] = useState<TemplateFragment | null>(null)
  const [creating, setCreating] = useState(false)

  return (
    <div style={{ padding: '16px 24px' }}>
      <Group justify="space-between" mb="sm">
        <Text size="lg" fw={600} c="dark.0">
          Fragments
        </Text>
        <Button size="sm" variant="light" leftSection={<IconPlus size={14} />} onClick={() => setCreating(true)}>
          New fragment
        </Button>
      </Group>

      {fragments && fragments.length > 0 ? (
        <Table verticalSpacing="sm" highlightOnHover>
          <Table.Thead>
            <Table.Tr>
              <Table.Th>Name</Table.Th>
              <Table.Th>Category</Table.Th>
              <Table.Th>Version</Table.Th>
              <Table.Th>Author</Table.Th>
              <Table.Th />
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {fragments.map((f) => (
              <Table.Tr key={f.id} style={{ cursor: 'pointer' }} onClick={() => setEditing(f)}>
                <Table.Td>
                  <Text size="sm" fw={500} c="dark.0">
                    {f.name}
                  </Text>
                </Table.Td>
                <Table.Td>
                  {f.category ? (
                    <Badge size="sm" variant="light" color="gray">
                      {f.category}
                    </Badge>
                  ) : (
                    <Text size="sm" c="dark.4">—</Text>
                  )}
                </Table.Td>
                <Table.Td>
                  <Text size="sm" c="dark.3">v{f.version}</Text>
                </Table.Td>
                <Table.Td>
                  <Text size="sm" c="dark.3">{f.authorName ?? '—'}</Text>
                </Table.Td>
                <Table.Td onClick={(e) => e.stopPropagation()}>
                  <ActionIcon
                    variant="subtle"
                    color="red"
                    onClick={() => deleteFragment.mutate(f.id)}
                    loading={deleteFragment.isPending && deleteFragment.variables === f.id}
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
          No fragments yet — create reusable markdown building blocks to compose into templates.
        </Text>
      )}

      <FragmentEditorModal opened={creating} onClose={() => setCreating(false)} fragment={null} />
      <FragmentEditorModal opened={!!editing} onClose={() => setEditing(null)} fragment={editing} />
    </div>
  )
}

export function TemplatesView() {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div style={{ height: 60, minHeight: 60, display: 'flex', alignItems: 'center', gap: 12, padding: '0 20px', borderBottom: '1px solid #1d1e21' }}>
        <Text fw={600} size="md" c="dark.0">
          Templates
        </Text>
      </div>

      <Tabs defaultValue="templates" style={{ display: 'flex', flexDirection: 'column', flex: 1, minHeight: 0 }}>
        <Tabs.List px={12}>
          <Tabs.Tab value="templates">Templates</Tabs.Tab>
          <Tabs.Tab value="fragments">Fragments</Tabs.Tab>
        </Tabs.List>

        <Tabs.Panel value="templates" style={{ flex: 1, overflowY: 'auto' }}>
          <TemplatesTab />
        </Tabs.Panel>

        <Tabs.Panel value="fragments" style={{ flex: 1, overflowY: 'auto' }}>
          <FragmentsTab />
        </Tabs.Panel>
      </Tabs>
    </div>
  )
}
