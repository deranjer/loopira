import { useEffect, useState } from 'react'
import { ActionIcon, Button, Select, Tabs, Text, Textarea, TextInput } from '@mantine/core'
import { IconArrowLeft, IconPlus } from '@tabler/icons-react'
import { useNavigate, useOutletContext, useParams } from 'react-router-dom'
import type { Team } from '../lib/api/types'
import { useIssues, useProject, useUpdateProject, useUsers } from '../lib/api/hooks'
import { IssueRow } from '../components/IssueRow'
import { IssueDetailPanel } from '../components/IssueDetailPanel'
import { NewIssueModal } from '../components/NewIssueModal'
import { MembersSection } from '../components/MembersSection'
import { DocumentsSection } from '../components/DocumentsSection'
import { WorkLogSection } from '../components/WorkLogSection'
import { PRIORITY_META, PROJECT_STATUS_META, PROJECT_STATUS_ORDER } from '../theme'

export function ProjectDetailView() {
  const { id } = useParams<{ id: string }>()
  const { team } = useOutletContext<{ team: Team | undefined }>()
  const navigate = useNavigate()
  const { data: project } = useProject(id)
  const { data: users } = useUsers()
  const updateProject = useUpdateProject()
  const { data: issues } = useIssues(team && id ? { teamId: team.id, projectId: id } : undefined)
  const [selectedIssueId, setSelectedIssueId] = useState<string | null>(null)
  const [newIssueOpen, setNewIssueOpen] = useState(false)

  const [name, setName] = useState('')
  const [description, setDescription] = useState('')

  useEffect(() => {
    if (project) {
      setName(project.name)
      setDescription(project.description ?? '')
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [project?.id])

  if (!id || !project) {
    return null
  }

  function save(patch: Partial<{ name: string; description: string; status: string; priority: number; leadId: string | null; targetDate: string | null }>) {
    if (!project) return
    updateProject.mutate({
      id: project.id,
      input: {
        name: patch.name ?? project.name,
        description: patch.description ?? project.description ?? '',
        status: patch.status ?? project.status,
        priority: patch.priority ?? project.priority,
        leadId: patch.leadId !== undefined ? patch.leadId : project.leadId,
        targetDate: patch.targetDate !== undefined ? patch.targetDate : project.targetDate,
      },
    })
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div style={{ height: 60, minHeight: 60, display: 'flex', alignItems: 'center', gap: 12, padding: '0 20px', borderBottom: '1px solid #1d1e21' }}>
        <ActionIcon variant="subtle" onClick={() => navigate('/projects')}>
          <IconArrowLeft size={18} />
        </ActionIcon>
        <Text fw={600} size="md" c="dark.0">
          {project.name}
        </Text>
        <div style={{ marginLeft: 'auto' }}>
          <Button size="sm" leftSection={<IconPlus size={15} />} onClick={() => setNewIssueOpen(true)}>
            New issue
          </Button>
        </div>
      </div>

      <div style={{ flex: 1, overflowY: 'auto', display: 'flex' }}>
        <div style={{ flex: 2, minWidth: 0, borderRight: '1px solid #1d1e21', display: 'flex', flexDirection: 'column' }}>
          <div style={{ padding: '20px 24px', borderBottom: '1px solid #1d1e21', display: 'flex', flexDirection: 'column', gap: 10 }}>
            <TextInput
              value={name}
              onChange={(e) => setName(e.currentTarget.value)}
              onBlur={() => name.trim() && name !== project.name && save({ name })}
              variant="unstyled"
              styles={{ input: { fontSize: 20, fontWeight: 600 } }}
            />
            <Textarea
              placeholder="Add a description..."
              value={description}
              onChange={(e) => setDescription(e.currentTarget.value)}
              onBlur={() => description !== (project.description ?? '') && save({ description })}
              variant="unstyled"
              minRows={2}
              autosize
              styles={{ input: { fontSize: 14, color: 'var(--mantine-color-dark-2)' } }}
            />
          </div>

          <Tabs defaultValue="issues" style={{ display: 'flex', flexDirection: 'column', flex: 1, minHeight: 0 }}>
            <Tabs.List px={12}>
              <Tabs.Tab value="issues">Issues</Tabs.Tab>
              <Tabs.Tab value="worklog">Work Log</Tabs.Tab>
            </Tabs.List>

            <Tabs.Panel value="issues" style={{ flex: 1, overflowY: 'auto' }}>
              {(issues ?? []).map((issue) => (
                <IssueRow key={issue.id} issue={issue} onClick={() => setSelectedIssueId(issue.id)} />
              ))}
              {issues?.length === 0 && (
                <Text size="sm" c="dark.4" p="lg">
                  No issues in this project yet.
                </Text>
              )}
            </Tabs.Panel>

            <Tabs.Panel value="worklog" style={{ flex: 1, overflowY: 'auto' }}>
              <WorkLogSection projectId={project.id} />
            </Tabs.Panel>
          </Tabs>
        </div>

        <div style={{ flex: 1, minWidth: 280, maxWidth: 340, padding: '20px 24px', display: 'flex', flexDirection: 'column', gap: 20 }}>
          <div>
            <Text size="xs" fw={600} c="dark.4" mb={6} style={{ letterSpacing: '.04em' }}>
              STATUS
            </Text>
            <Select
              data={PROJECT_STATUS_ORDER.map((s) => ({ value: s, label: PROJECT_STATUS_META[s].label }))}
              value={project.status}
              onChange={(v) => v && save({ status: v })}
              allowDeselect={false}
            />
          </div>

          <div>
            <Text size="xs" fw={600} c="dark.4" mb={6} style={{ letterSpacing: '.04em' }}>
              PRIORITY
            </Text>
            <Select
              data={PRIORITY_META.map((p, i) => ({ value: String(i), label: p.label }))}
              value={String(project.priority)}
              onChange={(v) => v !== null && save({ priority: Number(v) })}
              allowDeselect={false}
            />
          </div>

          <div>
            <Text size="xs" fw={600} c="dark.4" mb={6} style={{ letterSpacing: '.04em' }}>
              LEAD
            </Text>
            <Select
              placeholder="No lead"
              data={(users ?? []).map((u) => ({ value: u.id, label: u.name }))}
              value={project.leadId}
              onChange={(v) => save({ leadId: v })}
              clearable
              searchable
            />
          </div>

          <div>
            <Text size="xs" fw={600} c="dark.4" mb={6} style={{ letterSpacing: '.04em' }}>
              TARGET DATE
            </Text>
            <TextInput
              type="date"
              value={project.targetDate ?? ''}
              onChange={(e) => save({ targetDate: e.currentTarget.value || null })}
            />
          </div>

          <MembersSection projectId={project.id} />
          <DocumentsSection projectId={project.id} />
        </div>
      </div>

      <IssueDetailPanel issueId={selectedIssueId} teamId={team?.id} onClose={() => setSelectedIssueId(null)} />
      {team && <NewIssueModal opened={newIssueOpen} onClose={() => setNewIssueOpen(false)} teamId={team.id} />}
    </div>
  )
}
