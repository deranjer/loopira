import { Button, Group, Modal, Select, Stack, Textarea, TextInput } from '@mantine/core'
import { useForm } from '@mantine/form'
import { useCreateIssue, useProjects, useUsers } from '../lib/api/hooks'
import { PRIORITY_META } from '../theme'

const PRIORITY_OPTIONS = PRIORITY_META.map((p, i) => ({ value: String(i), label: p.label }))

export function NewIssueModal({
  opened,
  onClose,
  teamId,
}: {
  opened: boolean
  onClose: () => void
  teamId: string
}) {
  const { data: users } = useUsers()
  const { data: projects } = useProjects(teamId)
  const createIssue = useCreateIssue()

  const form = useForm({
    initialValues: { title: '', description: '', priority: '0', assigneeId: '', projectId: '' },
  })

  function submit(values: typeof form.values) {
    createIssue.mutate(
      {
        teamId,
        title: values.title,
        description: values.description,
        priority: Number(values.priority),
        assigneeId: values.assigneeId || null,
        projectId: values.projectId || null,
      },
      {
        onSuccess: () => {
          form.reset()
          onClose()
        },
      },
    )
  }

  return (
    <Modal opened={opened} onClose={onClose} title="New issue" size="xl">
      <form onSubmit={form.onSubmit(submit)}>
        <Stack gap="sm">
          <TextInput
            placeholder="Issue title"
            required
            autoFocus
            {...form.getInputProps('title')}
          />
          <Textarea placeholder="Add description..." minRows={3} autosize {...form.getInputProps('description')} />
          <Group grow>
            <Select
              label="Priority"
              data={PRIORITY_OPTIONS}
              {...form.getInputProps('priority')}
            />
            <Select
              label="Assignee"
              placeholder="Unassigned"
              data={(users ?? []).map((u) => ({ value: u.id, label: u.name }))}
              clearable
              {...form.getInputProps('assigneeId')}
            />
            <Select
              label="Project"
              placeholder="None"
              data={(projects ?? []).map((p) => ({ value: p.id, label: p.name }))}
              clearable
              {...form.getInputProps('projectId')}
            />
          </Group>
          <Group justify="flex-end" mt="sm">
            <Button variant="subtle" onClick={onClose} type="button">
              Cancel
            </Button>
            <Button type="submit" loading={createIssue.isPending} disabled={!form.values.title}>
              Create issue
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  )
}
