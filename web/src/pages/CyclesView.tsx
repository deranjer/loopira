import { useState } from 'react'
import { Badge, Button, Progress, Text } from '@mantine/core'
import { IconPlus } from '@tabler/icons-react'
import { useOutletContext } from 'react-router-dom'
import type { Team } from '../lib/api/types'
import { useCycles } from '../lib/api/hooks'
import { NewCycleModal } from '../components/NewCycleModal'

export function CyclesView() {
  const { team } = useOutletContext<{ team: Team | undefined }>()
  const { data: cycles } = useCycles(team?.id)
  const [newCycleOpen, setNewCycleOpen] = useState(false)

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div style={{ height: 60, minHeight: 60, display: 'flex', alignItems: 'center', gap: 12, padding: '0 20px', borderBottom: '1px solid #1d1e21' }}>
        <Text fw={600} size="md" c="dark.0">
          Cycles
        </Text>
        <div style={{ marginLeft: 'auto' }}>
          <Button size="sm" leftSection={<IconPlus size={15} />} onClick={() => setNewCycleOpen(true)}>
            New cycle
          </Button>
        </div>
      </div>
      <div style={{ flex: 1, overflowY: 'auto', padding: '24px 28px', display: 'flex', flexDirection: 'column', gap: 16, maxWidth: 860 }}>
        {(cycles ?? []).map((cycle) => (
          <div key={cycle.id} style={{ border: '1px solid #1d1e21', borderRadius: 12, padding: 20, background: '#111214' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 12 }}>
              <Text fw={600} size="md" c="dark.0">
                Cycle {cycle.number}
              </Text>
              {cycle.active && (
                <Badge size="md" variant="light" color="accent">
                  Active
                </Badge>
              )}
              <Text size="sm" c="dark.4" style={{ marginLeft: 'auto' }}>
                {cycle.startDate} – {cycle.endDate}
              </Text>
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
              <Progress value={cycle.progress} color="accent" flex={1} size="md" radius="xl" />
              <Text size="sm" c="dark.3" style={{ flexShrink: 0 }}>
                {cycle.done}/{cycle.total} done
              </Text>
            </div>
          </div>
        ))}
        {cycles?.length === 0 && (
          <Text size="md" c="dark.4">
            No cycles yet.
          </Text>
        )}
      </div>

      {team && <NewCycleModal opened={newCycleOpen} onClose={() => setNewCycleOpen(false)} teamId={team.id} />}
    </div>
  )
}
