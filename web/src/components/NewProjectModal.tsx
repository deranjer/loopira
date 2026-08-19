import { Button, Group, Modal, Stack, Textarea, TextInput } from '@mantine/core'
import { useForm } from '@mantine/form'
import { useCreateProject } from '../lib/api/hooks'

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
  const form = useForm({ initialValues: { name: '', description: '' } })

  function submit(values: typeof form.values) {
    createProject.mutate(
      { teamId, name: values.name, description: values.description },
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
