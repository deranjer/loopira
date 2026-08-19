import { useEffect } from 'react'
import { Avatar, ColorSwatch, Drawer, Group, Select, Stack, Text, Textarea, TextInput } from '@mantine/core'
import { useForm } from '@mantine/form'
import {
  useIssue,
  useLabels,
  useProjects,
  useUpdateIssueDetails,
  useUpdateIssueStatus,
  useUsers,
} from '../lib/api/hooks'
import { PRIORITY_META, STATUS_META, STATUS_ORDER, avatarColor } from '../theme'

const STATUS_OPTIONS = STATUS_ORDER.map((s) => ({ value: s, label: STATUS_META[s].label }))
const PRIORITY_OPTIONS = PRIORITY_META.map((p, i) => ({ value: String(i), label: p.label }))

function labelSelectProps(labels: { id: string; name: string; color: string }[] | undefined) {
  const colorById = new Map((labels ?? []).map((l) => [l.id, l.color]))
  return {
    data: (labels ?? []).map((l) => ({ value: l.id, label: l.name })),
    renderOption: ({ option }: { option: { value: string; label: string } }) => (
      <Group gap="xs" wrap="nowrap">
        <ColorSwatch color={colorById.get(option.value) ?? '#8a8f98'} size={12} />
        {option.label}
      </Group>
    ),
  }
}

export function IssueDetailPanel({
  issueId,
  teamId,
  onClose,
}: {
  issueId: string | null
  teamId: string | undefined
  onClose: () => void
}) {
  const { data: issue } = useIssue(issueId ?? undefined)
  const { data: users } = useUsers()
  const { data: projects } = useProjects(teamId)
  const { data: labels } = useLabels(teamId)
  const updateStatus = useUpdateIssueStatus()
  const updateDetails = useUpdateIssueDetails()

  const form = useForm({
    initialValues: {
      title: '',
      description: '',
      priority: '0',
      assigneeId: '',
      projectId: '',
      labelId: '',
    },
  })

  useEffect(() => {
    if (issue) {
      form.setValues({
        title: issue.title,
        description: issue.description,
        priority: String(issue.priority),
        assigneeId: issue.assigneeId ?? '',
        projectId: issue.projectId ?? '',
        labelId: issue.label?.id ?? '',
      })
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [issue?.id])

  if (!issue) {
    return <Drawer opened={!!issueId} onClose={onClose} position="right" size={560} withCloseButton={false} />
  }

  const userOptions = (users ?? []).map((u) => ({ value: u.id, label: u.name }))
  const projectOptions = (projects ?? []).map((p) => ({ value: p.id, label: p.name }))

  function save() {
    updateDetails.mutate({
      id: issue!.id,
      input: {
        title: form.values.title,
        description: form.values.description,
        priority: Number(form.values.priority),
        assigneeId: form.values.assigneeId || null,
        projectId: form.values.projectId || null,
        cycleId: issue!.cycleId,
        labelId: form.values.labelId || null,
      },
    })
  }

  return (
    <Drawer
      opened={!!issueId}
      onClose={onClose}
      position="right"
      size={560}
      title={<Text size="sm" c="dark.3">{issue.identifier}</Text>}
      styles={{ content: { background: '#0e0f11' }, header: { background: '#0e0f11' } }}
    >
      <Stack gap="lg">
        <TextInput
          value={form.values.title}
          onChange={(e) => form.setFieldValue('title', e.currentTarget.value)}
          onBlur={save}
          variant="unstyled"
          styles={{ input: { fontSize: 22, fontWeight: 600, color: '#f2f3f4' } }}
        />
        <Textarea
          value={form.values.description}
          onChange={(e) => form.setFieldValue('description', e.currentTarget.value)}
          onBlur={save}
          autosize
          minRows={3}
          variant="unstyled"
          styles={{ input: { fontSize: 15, color: '#a9adb3', lineHeight: 1.6 } }}
        />

        <Stack gap="md" style={{ borderTop: '1px solid #1d1e21', paddingTop: 20 }}>
          <Select
            label="Status"
            data={STATUS_OPTIONS}
            value={issue.status}
            onChange={(v) => v && updateStatus.mutate({ id: issue.id, status: v })}
          />
          <Select
            label="Priority"
            data={PRIORITY_OPTIONS}
            value={String(issue.priority)}
            onChange={(v) => {
              if (!v) return
              form.setFieldValue('priority', v)
              updateDetails.mutate({
                id: issue.id,
                input: {
                  title: issue.title,
                  description: issue.description,
                  priority: Number(v),
                  assigneeId: issue.assigneeId,
                  projectId: issue.projectId,
                  cycleId: issue.cycleId,
                  labelId: form.values.labelId || null,
                },
              })
            }}
          />
          <Select
            label="Assignee"
            data={userOptions}
            value={form.values.assigneeId || null}
            clearable
            leftSection={
              issue.assigneeName ? (
                <Avatar size={19} radius="xl" color={avatarColor(issue.assigneeName)} />
              ) : undefined
            }
            onChange={(v) => {
              form.setFieldValue('assigneeId', v ?? '')
              updateDetails.mutate({
                id: issue.id,
                input: {
                  title: issue.title,
                  description: issue.description,
                  priority: issue.priority,
                  assigneeId: v,
                  projectId: issue.projectId,
                  cycleId: issue.cycleId,
                  labelId: form.values.labelId || null,
                },
              })
            }}
          />
          <Select
            label="Project"
            data={projectOptions}
            value={form.values.projectId || null}
            clearable
            onChange={(v) => {
              form.setFieldValue('projectId', v ?? '')
              updateDetails.mutate({
                id: issue.id,
                input: {
                  title: issue.title,
                  description: issue.description,
                  priority: issue.priority,
                  assigneeId: issue.assigneeId,
                  projectId: v,
                  cycleId: issue.cycleId,
                  labelId: form.values.labelId || null,
                },
              })
            }}
          />
          <Select
            label="Label"
            value={form.values.labelId || null}
            clearable
            {...labelSelectProps(labels)}
            onChange={(v) => {
              form.setFieldValue('labelId', v ?? '')
              updateDetails.mutate({
                id: issue.id,
                input: {
                  title: issue.title,
                  description: issue.description,
                  priority: issue.priority,
                  assigneeId: issue.assigneeId,
                  projectId: issue.projectId,
                  cycleId: issue.cycleId,
                  labelId: v,
                },
              })
            }}
          />
        </Stack>
      </Stack>
    </Drawer>
  )
}
