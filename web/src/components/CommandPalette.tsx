import { Spotlight, type SpotlightActionData } from '@mantine/spotlight'
import {
  IconFolder,
  IconLayoutKanban,
  IconListDetails,
  IconRepeat,
  IconSearch,
  IconSettings,
  IconStack2,
  IconUser,
} from '@tabler/icons-react'
import { useNavigate } from 'react-router-dom'
import { useIssues } from '../lib/api/hooks'

export function CommandPalette({ teamId }: { teamId: string | undefined }) {
  const navigate = useNavigate()
  const { data: issues } = useIssues(teamId ? { teamId } : undefined)

  const navActions: SpotlightActionData[] = [
    { id: 'nav-issues', label: 'Go to Issues', leftSection: <IconListDetails size={18} />, onClick: () => navigate('/issues') },
    { id: 'nav-my-issues', label: 'Go to My Issues', leftSection: <IconUser size={18} />, onClick: () => navigate('/my-issues') },
    { id: 'nav-board', label: 'Go to Board', leftSection: <IconLayoutKanban size={18} />, onClick: () => navigate('/board') },
    { id: 'nav-cycles', label: 'Go to Cycles', leftSection: <IconRepeat size={18} />, onClick: () => navigate('/cycles') },
    { id: 'nav-projects', label: 'Go to Projects', leftSection: <IconFolder size={18} />, onClick: () => navigate('/projects') },
    { id: 'nav-views', label: 'Go to Views', leftSection: <IconStack2 size={18} />, onClick: () => navigate('/views') },
    { id: 'nav-settings', label: 'Go to Settings', leftSection: <IconSettings size={18} />, onClick: () => navigate('/settings') },
  ]

  const issueActions: SpotlightActionData[] = (issues ?? []).slice(0, 20).map((issue) => ({
    id: `issue-${issue.id}`,
    label: issue.title,
    description: issue.identifier,
    leftSection: <IconSearch size={16} />,
    onClick: () => navigate('/issues'),
  }))

  return (
    <Spotlight
      actions={[...navActions, ...issueActions]}
      nothingFound="Nothing found"
      searchProps={{ placeholder: 'Search issues, jump to a view...' }}
      shortcut={['mod + K']}
    />
  )
}
