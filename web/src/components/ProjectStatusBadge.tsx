import { Badge } from '@mantine/core'
import { PROJECT_STATUS_META, type ProjectStatusKey } from '../theme'

export function ProjectStatusBadge({ status }: { status: ProjectStatusKey }) {
  const meta = PROJECT_STATUS_META[status]
  return (
    <Badge
      size="sm"
      variant="light"
      color="gray"
      leftSection={
        <span style={{ width: 7, height: 7, borderRadius: '50%', background: meta.color, display: 'inline-block' }} />
      }
      styles={{ label: { color: meta.color } }}
    >
      {meta.label}
    </Badge>
  )
}
