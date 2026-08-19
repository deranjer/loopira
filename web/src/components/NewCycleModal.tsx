import { Button, Group, Modal, Stack, TextInput } from '@mantine/core'
import { useForm } from '@mantine/form'
import { useCreateCycle } from '../lib/api/hooks'

function todayISO() {
  return new Date().toISOString().slice(0, 10)
}

function twoWeeksFromISO(startISO: string) {
  const d = new Date(startISO)
  d.setDate(d.getDate() + 14)
  return d.toISOString().slice(0, 10)
}

export function NewCycleModal({
  opened,
  onClose,
  teamId,
}: {
  opened: boolean
  onClose: () => void
  teamId: string
}) {
  const createCycle = useCreateCycle()
  const form = useForm({
    initialValues: { startDate: todayISO(), endDate: twoWeeksFromISO(todayISO()) },
  })

  function submit(values: typeof form.values) {
    createCycle.mutate(
      { teamId, startDate: values.startDate, endDate: values.endDate },
      {
        onSuccess: () => {
          form.reset()
          onClose()
        },
      },
    )
  }

  return (
    <Modal opened={opened} onClose={onClose} title="New cycle" size="md">
      <form onSubmit={form.onSubmit(submit)}>
        <Stack gap="sm">
          <Group grow>
            <TextInput label="Start date" type="date" required {...form.getInputProps('startDate')} />
            <TextInput label="End date" type="date" required {...form.getInputProps('endDate')} />
          </Group>
          <Group justify="flex-end" mt="sm">
            <Button variant="subtle" onClick={onClose} type="button">
              Cancel
            </Button>
            <Button type="submit" loading={createCycle.isPending}>
              Create cycle
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  )
}
