import { useEffect } from 'react'
import { Button, Divider, Stack, Text, TextInput } from '@mantine/core'
import { useForm } from '@mantine/form'
import { useOutletContext } from 'react-router-dom'
import type { Team } from '../lib/api/types'
import { useRenameTeam } from '../lib/api/hooks'
import { ApiKeysSection } from '../components/ApiKeysSection'

export function SettingsView() {
  const { team } = useOutletContext<{ team: Team | undefined }>()
  const rename = useRenameTeam()
  const form = useForm({ initialValues: { name: '' } })

  useEffect(() => {
    if (team) form.setFieldValue('name', team.name)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [team?.id])

  return (
    <div style={{ flex: 1, display: 'flex', overflow: 'hidden' }}>
      <div style={{ flex: 1, overflowY: 'auto', padding: '28px 32px', maxWidth: 720 }}>
        <Text size="xl" fw={600} c="dark.0" mb={20}>
          General
        </Text>
        <form onSubmit={form.onSubmit((values) => team && rename.mutate({ id: team.id, name: values.name }))}>
          <Stack gap="md">
            <TextInput label="Workspace name" {...form.getInputProps('name')} />
            <TextInput label="URL" value={`${team?.key.toLowerCase() ?? ''}.loopira.app`} disabled />
            <Button type="submit" style={{ alignSelf: 'flex-start' }} loading={rename.isPending}>
              Save changes
            </Button>
          </Stack>
        </form>

        <Divider my={32} />

        <ApiKeysSection />
      </div>
    </div>
  )
}
