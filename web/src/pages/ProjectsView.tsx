import { useState } from 'react'
import { Button, Progress, Text } from '@mantine/core'
import { IconPlus } from '@tabler/icons-react'
import { useOutletContext } from 'react-router-dom'
import type { Team } from '../lib/api/types'
import { useProjects } from '../lib/api/hooks'
import { avatarColor } from '../theme'
import { NewProjectModal } from '../components/NewProjectModal'

export function ProjectsView() {
  const { team } = useOutletContext<{ team: Team | undefined }>()
  const { data: projects } = useProjects(team?.id)
  const [newProjectOpen, setNewProjectOpen] = useState(false)

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
      <div
        style={{
          flex: 1,
          overflowY: 'auto',
          padding: '24px 28px',
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
          gap: 16,
          alignContent: 'start',
        }}
      >
        {(projects ?? []).map((project) => (
          <div key={project.id} style={{ border: '1px solid #1d1e21', borderRadius: 12, padding: 20, background: '#111214', display: 'flex', flexDirection: 'column', gap: 12 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <span style={{ width: 24, height: 24, borderRadius: 6, background: avatarColor(project.name), flexShrink: 0 }} />
              <Text size="md" fw={600} c="dark.0">
                {project.name}
              </Text>
            </div>
            <Text size="sm" c="dark.3" style={{ lineHeight: 1.5 }}>
              {project.description}
            </Text>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <Progress value={project.progress} color="accent" flex={1} size="md" radius="xl" />
              <Text size="sm" c="dark.4" style={{ flexShrink: 0 }}>
                {project.progress}%
              </Text>
            </div>
          </div>
        ))}
        {projects?.length === 0 && (
          <Text size="md" c="dark.4">
            No projects yet.
          </Text>
        )}
      </div>

      {team && <NewProjectModal opened={newProjectOpen} onClose={() => setNewProjectOpen(false)} teamId={team.id} />}
    </div>
  )
}
