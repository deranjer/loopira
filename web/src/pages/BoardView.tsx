import { useState, type CSSProperties } from 'react'
import { Avatar, Button, Text } from '@mantine/core'
import { IconPlus } from '@tabler/icons-react'
import {
  DndContext,
  DragOverlay,
  type DragEndEvent,
  type DragStartEvent,
  PointerSensor,
  useDraggable,
  useDroppable,
  useSensor,
  useSensors,
} from '@dnd-kit/core'
import { useOutletContext } from 'react-router-dom'
import type { Issue, Team } from '../lib/api/types'
import { useIssues, useUpdateIssueStatus } from '../lib/api/hooks'
import { StatusDot } from '../components/StatusDot'
import { PriorityBars } from '../components/PriorityBars'
import { IssueDetailPanel } from '../components/IssueDetailPanel'
import { NewIssueModal } from '../components/NewIssueModal'
import { STATUS_META, STATUS_ORDER, avatarColor } from '../theme'

const CARD_STYLE: CSSProperties = {
  background: '#141518',
  border: '1px solid #1d1e21',
  borderRadius: 10,
  padding: 14,
  display: 'flex',
  flexDirection: 'column',
  gap: 10,
}

function BoardCardContent({ issue }: { issue: Issue }) {
  return (
    <>
      <Text size="sm" c="dark.4">
        {issue.identifier}
      </Text>
      <Text size="md" c="dark.1" style={{ lineHeight: 1.4 }}>
        {issue.title}
      </Text>
      <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
        <PriorityBars priority={issue.priority} />
        <Text size="sm" c="dark.4" style={{ marginLeft: 'auto' }}>
          {issue.projectName}
        </Text>
        {issue.assigneeName && (
          <Avatar size={24} radius="xl" color={avatarColor(issue.assigneeName)}>
            {issue.assigneeName.split(' ').map((p) => p[0]).join('')}
          </Avatar>
        )}
      </div>
    </>
  )
}

function BoardCard({ issue, onClick }: { issue: Issue; onClick: () => void }) {
  // Deliberately no transform/translate here — the element stays put in
  // the (overflow: auto) column while dragging and just dims. The actual
  // floating "you're dragging this" visual is DragOverlay, portaled to the
  // document root so it isn't clipped by the column's scroll container
  // (the classic dnd-kit bug: transform an in-flow element inside an
  // overflow:auto box and it gets sliced by the scroll clip instead of
  // floating above it).
  const { attributes, listeners, setNodeRef, isDragging } = useDraggable({
    id: issue.id,
    data: { status: issue.status },
  })
  return (
    <div
      ref={setNodeRef}
      {...listeners}
      {...attributes}
      onClick={onClick}
      style={{
        ...CARD_STYLE,
        cursor: 'grab',
        opacity: isDragging ? 0.35 : 1,
        touchAction: 'none',
      }}
    >
      <BoardCardContent issue={issue} />
    </div>
  )
}

function BoardColumn({
  status,
  issues,
  onCardClick,
}: {
  status: (typeof STATUS_ORDER)[number]
  issues: Issue[]
  onCardClick: (id: string) => void
}) {
  const { setNodeRef, isOver } = useDroppable({ id: status, data: { status } })
  const meta = STATUS_META[status]
  return (
    <div style={{ width: 300, minWidth: 300, display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '6px 8px 12px 8px' }}>
        <StatusDot status={status} size={14} />
        <Text size="md" fw={600} c="dark.2">
          {meta.label}
        </Text>
        <Text size="sm" c="dark.4">
          {issues.length}
        </Text>
      </div>
      <div
        ref={setNodeRef}
        style={{
          flex: 1,
          overflowY: 'auto',
          display: 'flex',
          flexDirection: 'column',
          gap: 8,
          background: isOver ? 'rgba(94,106,210,0.06)' : undefined,
          borderRadius: 10,
          padding: 4,
          transition: 'background 0.1s ease',
        }}
      >
        {issues.map((issue) => (
          <BoardCard key={issue.id} issue={issue} onClick={() => onCardClick(issue.id)} />
        ))}
      </div>
    </div>
  )
}

export function BoardView() {
  const { team } = useOutletContext<{ team: Team | undefined }>()
  const { data: issues } = useIssues(team ? { teamId: team.id } : undefined)
  const updateStatus = useUpdateIssueStatus()
  const [selectedIssueId, setSelectedIssueId] = useState<string | null>(null)
  const [newIssueOpen, setNewIssueOpen] = useState(false)
  const [activeIssue, setActiveIssue] = useState<Issue | null>(null)
  // A distance threshold keeps a plain click from being swallowed as a
  // drag — without it, PointerSensor activates (and preventDefaults) on
  // any pointerdown, which suppresses the card's own onClick.
  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 8 } }))

  function handleDragStart(event: DragStartEvent) {
    setActiveIssue(issues?.find((i) => i.id === event.active.id) ?? null)
  }

  function handleDragEnd(event: DragEndEvent) {
    setActiveIssue(null)
    const { active, over } = event
    if (!over) return
    const newStatus = over.id as string
    const issue = issues?.find((i) => i.id === active.id)
    if (issue && issue.status !== newStatus) {
      updateStatus.mutate({ id: issue.id, status: newStatus })
    }
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div style={{ height: 60, minHeight: 60, display: 'flex', alignItems: 'center', gap: 12, padding: '0 20px', borderBottom: '1px solid #1d1e21' }}>
        <Text fw={600} size="md" c="dark.0">
          Board
        </Text>
        <div style={{ marginLeft: 'auto' }}>
          <Button size="sm" leftSection={<IconPlus size={15} />} onClick={() => setNewIssueOpen(true)}>
            New issue
          </Button>
        </div>
      </div>

      <DndContext sensors={sensors} onDragStart={handleDragStart} onDragEnd={handleDragEnd}>
        <div style={{ flex: 1, overflowX: 'auto', overflowY: 'hidden', display: 'flex', padding: '18px 16px', gap: 16 }}>
          {STATUS_ORDER.map((status) => (
            <BoardColumn
              key={status}
              status={status}
              issues={(issues ?? []).filter((i) => i.status === status)}
              onCardClick={setSelectedIssueId}
            />
          ))}
        </div>

        <DragOverlay dropAnimation={null}>
          {activeIssue && (
            <div style={{ ...CARD_STYLE, width: 292, cursor: 'grabbing', boxShadow: '0 12px 32px rgba(0,0,0,0.5)', rotate: '2deg' }}>
              <BoardCardContent issue={activeIssue} />
            </div>
          )}
        </DragOverlay>
      </DndContext>

      <IssueDetailPanel issueId={selectedIssueId} teamId={team?.id} onClose={() => setSelectedIssueId(null)} />
      {team && <NewIssueModal opened={newIssueOpen} onClose={() => setNewIssueOpen(false)} teamId={team.id} />}
    </div>
  )
}
