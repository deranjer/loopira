import { useState } from 'react'
import { Avatar, Button, Table, Text } from '@mantine/core'
import { IconPlus } from '@tabler/icons-react'
import { useNavigate, useOutletContext } from 'react-router-dom'
import type { Team } from '../lib/api/types'
import { useProjects } from '../lib/api/hooks'
import { avatarColor } from '../theme'
import { NewProjectModal } from '../components/NewProjectModal'
import { PriorityBars } from '../components/PriorityBars'
import { ProjectStatusBadge } from '../components/ProjectStatusBadge'

function formatTargetDate(iso: string | null) {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
}

export function ProjectsView() {
  const { team } = useOutletContext<{ team: Team | undefined }>()
  const { data: projects } = useProjects(team?.id)
  const [newProjectOpen, setNewProjectOpen] = useState(false)
  const navigate = useNavigate()

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div style={{ height: 60, minHeight: 60, display: 'flex', alignItems: 'center', gap: 12, padding: '0 20px', borderBottom: '1px solid #1d1e21' }}>
        <Text fw={600} size="md" c="dark.0">
          Projects
        </Text>
        <div style={{ marginLeft: 'auto' }}>
          <Button size="sm" leftSection={<IconPlus size={15} />} onClick={() => setNewProjectOpen(true)}>
            New project
          </Button>
        </div>
      </div>
      <div style={{ flex: 1, overflowY: 'auto', padding: '0 8px' }}>
        {projects && projects.length > 0 ? (
          <Table verticalSpacing="sm" highlightOnHover>
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Name</Table.Th>
                <Table.Th>Issues</Table.Th>
                <Table.Th>Priority</Table.Th>
                <Table.Th>Lead</Table.Th>
                <Table.Th>Target date</Table.Th>
                <Table.Th>Status</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {projects.map((project) => (
                <Table.Tr
                  key={project.id}
                  style={{ cursor: 'pointer' }}
                  onClick={() => navigate(`/projects/${project.id}`)}
                >
                  <Table.Td>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                      <span style={{ width: 20, height: 20, borderRadius: 6, background: avatarColor(project.name), flexShrink: 0 }} />
                      <Text size="sm" fw={500} c="dark.0">
                        {project.name}
                      </Text>
                    </div>
                  </Table.Td>
                  <Table.Td>
                    <Text size="sm" c="dark.3">
                      {project.issueCount}
                    </Text>
                  </Table.Td>
                  <Table.Td>
                    <PriorityBars priority={project.priority} />
                  </Table.Td>
                  <Table.Td>
                    {project.leadName ? (
                      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                        <Avatar size={20} radius="xl" color={avatarColor(project.leadName)}>
                          {project.leadName.slice(0, 1).toUpperCase()}
                        </Avatar>
                        <Text size="sm" c="dark.2">
                          {project.leadName}
                        </Text>
                      </div>
                    ) : (
                      <Text size="sm" c="dark.4">
                        —
                      </Text>
                    )}
                  </Table.Td>
                  <Table.Td>
                    <Text size="sm" c="dark.3">
                      {formatTargetDate(project.targetDate)}
                    </Text>
                  </Table.Td>
                  <Table.Td>
                    <ProjectStatusBadge status={project.status} />
                  </Table.Td>
                </Table.Tr>
              ))}
            </Table.Tbody>
          </Table>
        ) : (
          <Text size="md" c="dark.4" p="lg">
            No projects yet.
          </Text>
        )}
      </div>

      {team && <NewProjectModal opened={newProjectOpen} onClose={() => setNewProjectOpen(false)} teamId={team.id} />}
    </div>
  )
}
