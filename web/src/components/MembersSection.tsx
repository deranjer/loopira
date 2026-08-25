import { useMemo, useState } from 'react'
import { ActionIcon, Avatar, Button, Group, Select, Stack, Text } from '@mantine/core'
import { IconTrash } from '@tabler/icons-react'
import { useAddProjectMember, useProjectMembers, useRemoveProjectMember, useUsers } from '../lib/api/hooks'
import { avatarColor } from '../theme'

export function MembersSection({ projectId }: { projectId: string }) {
  const { data: members } = useProjectMembers(projectId)
  const { data: allUsers } = useUsers()
  const addMember = useAddProjectMember()
  const removeMember = useRemoveProjectMember()
  const [selectedUserId, setSelectedUserId] = useState<string | null>(null)

  const memberIds = useMemo(() => new Set((members ?? []).map((m) => m.id)), [members])
  const candidates = (allUsers ?? []).filter((u) => !memberIds.has(u.id))

  function handleAdd() {
    if (!selectedUserId) return
    addMember.mutate({ projectId, userId: selectedUserId }, { onSuccess: () => setSelectedUserId(null) })
  }

  return (
    <div>
      <Text size="lg" fw={600} c="dark.0" mb="sm">
        Members
      </Text>

      <Stack gap="xs" mb="md">
        {(members ?? []).map((member) => (
          <Group key={member.id} justify="space-between">
            <Group gap={8}>
              <Avatar size={24} radius="xl" color={avatarColor(member.name)}>
                {member.name.slice(0, 1).toUpperCase()}
              </Avatar>
              <Text size="sm" c="dark.1">
                {member.name}
              </Text>
            </Group>
            <ActionIcon
              variant="subtle"
              color="red"
              size="sm"
              onClick={() => removeMember.mutate({ projectId, userId: member.id })}
              loading={removeMember.isPending && removeMember.variables?.userId === member.id}
            >
              <IconTrash size={14} />
            </ActionIcon>
          </Group>
        ))}
        {members?.length === 0 && (
          <Text size="sm" c="dark.4">
            No members yet.
          </Text>
        )}
      </Stack>

      <Group gap="xs">
        <Select
          placeholder="Add a member..."
          data={candidates.map((u) => ({ value: u.id, label: u.name }))}
          value={selectedUserId}
          onChange={setSelectedUserId}
          searchable
          flex={1}
        />
        <Button size="sm" onClick={handleAdd} disabled={!selectedUserId} loading={addMember.isPending}>
          Add
        </Button>
      </Group>
    </div>
  )
}
