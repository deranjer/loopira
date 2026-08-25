import { ActionIcon, Badge, Table, Text } from '@mantine/core'
import { IconTrash } from '@tabler/icons-react'
import { useNavigate } from 'react-router-dom'
import { useDeleteView, useMe, useViews } from '../lib/api/hooks'
import type { ViewFilters } from '../lib/api/types'

function summarizeFilters(filters: ViewFilters) {
  const parts = Object.entries(filters)
    .filter(([, v]) => v !== undefined && v !== '')
    .map(([k]) => k)
  return parts.length > 0 ? parts.join(', ') : 'No filters'
}

export function ViewsView() {
  const { data: views } = useViews()
  const { data: me } = useMe()
  const deleteView = useDeleteView()
  const navigate = useNavigate()

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div style={{ height: 60, minHeight: 60, display: 'flex', alignItems: 'center', gap: 12, padding: '0 20px', borderBottom: '1px solid #1d1e21' }}>
        <Text fw={600} size="md" c="dark.0">
          Views
        </Text>
      </div>
      <div style={{ flex: 1, overflowY: 'auto', padding: '0 8px' }}>
        {views && views.length > 0 ? (
          <Table verticalSpacing="sm" highlightOnHover>
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Name</Table.Th>
                <Table.Th>Filters</Table.Th>
                <Table.Th>Visibility</Table.Th>
                <Table.Th />
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {views.map((view) => (
                <Table.Tr key={view.id} style={{ cursor: 'pointer' }} onClick={() => navigate(`/issues?viewId=${view.id}`)}>
                  <Table.Td>
                    <Text size="sm" fw={500} c="dark.0">
                      {view.name}
                    </Text>
                  </Table.Td>
                  <Table.Td>
                    <Text size="sm" c="dark.3">
                      {summarizeFilters(view.definition.filters)}
                    </Text>
                  </Table.Td>
                  <Table.Td>
                    <Badge size="sm" variant="light" color={view.shared ? 'accent' : 'gray'}>
                      {view.shared ? 'Shared' : 'Only me'}
                    </Badge>
                  </Table.Td>
                  <Table.Td onClick={(e) => e.stopPropagation()}>
                    {view.ownerId === me?.id && (
                      <ActionIcon
                        variant="subtle"
                        color="red"
                        onClick={() => deleteView.mutate(view.id)}
                        loading={deleteView.isPending && deleteView.variables === view.id}
                      >
                        <IconTrash size={16} />
                      </ActionIcon>
                    )}
                  </Table.Td>
                </Table.Tr>
              ))}
            </Table.Tbody>
          </Table>
        ) : (
          <Text size="md" c="dark.4" p="lg">
            No saved views yet — build a filter on the Issues page and save it.
          </Text>
        )}
      </div>
    </div>
  )
}
