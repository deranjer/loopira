import { Button, Group, Modal, Select, Stack, Textarea, TextInput } from '@mantine/core'
import { useForm } from '@mantine/form'
import { useCreateProject, useTemplates } from '../lib/api/hooks'

export function NewProjectModal({
  opened,
  onClose,
  teamId,
}: {
  opened: boolean
  onClose: () => void
  teamId: string
}) {
  const createProject = useCreateProject()
  const { data: templates } = useTemplates()
  const form = useForm({ initialValues: { name: '', description: '', templateId: '' } })

  function submit(values: typeof form.values) {
    createProject.mutate(
      { teamId, name: values.name, description: values.description, templateId: values.templateId || undefined },
      {
        onSuccess: () => {
          form.reset()
          onClose()
        },
      },
    )
  }

  return (
    <Modal opened={opened} onClose={onClose} title="New project" size="lg">
      <form onSubmit={form.onSubmit(submit)}>
        <Stack gap="sm">
          <TextInput placeholder="Project name" required autoFocus {...form.getInputProps('name')} />
          <Textarea placeholder="Add description..." minRows={3} autosize {...form.getInputProps('description')} />
          <Select
            label="Tech stack template"
            placeholder="No template"
            data={(templates ?? []).map((t) => ({ value: t.id, label: t.name }))}
            clearable
            searchable
            {...form.getInputProps('templateId')}
          />
          <Group justify="flex-end" mt="sm">
            <Button variant="subtle" onClick={onClose} type="button">
              Cancel
            </Button>
            <Button type="submit" loading={createProject.isPending} disabled={!form.values.name}>
              Create project
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  )
}
