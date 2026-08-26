import { AppShell as MantineAppShell, Avatar, Menu, Text, Tooltip, UnstyledButton } from '@mantine/core'
import { spotlight } from '@mantine/spotlight'
import {
  IconArrowUpCircle,
  IconFolder,
  IconLayoutKanban,
  IconListDetails,
  IconLogout,
  IconNotes,
  IconRepeat,
  IconSearch,
  IconSettings,
  IconStack2,
  IconUser,
} from '@tabler/icons-react'
import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { useCurrentTeam, useLogout, useMe, useVersion } from '../lib/api/hooks'
import { useTeamLiveUpdates } from '../lib/ws'
import { avatarColor } from '../theme'
import { CommandPalette } from './CommandPalette'

const NAV_ITEMS = [
  { to: '/issues', label: 'Issues', icon: IconListDetails },
  { to: '/my-issues', label: 'My Issues', icon: IconUser },
  { to: '/board', label: 'Board', icon: IconLayoutKanban },
  { to: '/cycles', label: 'Cycles', icon: IconRepeat },
  { to: '/projects', label: 'Projects', icon: IconFolder },
  { to: '/worklog', label: 'Work Log', icon: IconNotes },
  { to: '/views', label: 'Views', icon: IconStack2 },
  { to: '/settings', label: 'Settings', icon: IconSettings },
]

function initials(name: string) {
  return name
    .split(' ')
    .map((p) => p[0])
    .join('')
    .toUpperCase()
}

export function AppShell() {
  const { data: me } = useMe()
  const { data: versionInfo } = useVersion()
  const { team } = useCurrentTeam()
  const logout = useLogout()
  const navigate = useNavigate()
  useTeamLiveUpdates(team?.id)

  function signOut() {
    logout.mutate(undefined, { onSuccess: () => navigate('/login') })
  }

  return (
    <MantineAppShell header={{ height: 60 }} navbar={{ width: 272, breakpoint: 0 }} padding={0}>
      <MantineAppShell.Header
        bg="dark.8"
        style={{ borderBottom: '1px solid #1d1e21' }}
        px={18}
        display="flex"
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, width: '100%' }}>
          <div style={{ width: 26, height: 26, borderRadius: 6, background: 'var(--mantine-color-accent-5)', flexShrink: 0 }} />
          <Text fw={600} size="md" c="dark.0">
            {team?.name ?? 'Workspace'}
          </Text>
          <UnstyledButton
            onClick={() => spotlight.open()}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 10,
              padding: '8px 14px',
              marginLeft: 22,
              background: '#17181b',
              borderRadius: 7,
              color: '#75787f',
              fontSize: 15,
              width: 260,
            }}
          >
            <IconSearch size={17} />
            <span>Search</span>
            <span style={{ marginLeft: 'auto', fontSize: 12.5, color: '#4d4f54' }}>Ctrl K</span>
          </UnstyledButton>
          <div style={{ marginLeft: 'auto', display: 'flex', alignItems: 'center', gap: 16 }}>
            <Avatar size={30} radius="xl" color={avatarColor(me?.name ?? '')}>
              {me ? initials(me.name) : ''}
            </Avatar>
          </div>
        </div>
      </MantineAppShell.Header>

      <MantineAppShell.Navbar bg="dark.8" style={{ borderRight: '1px solid #1d1e21' }} p={10}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 2, marginBottom: 18 }}>
          {NAV_ITEMS.map(({ to, label, icon: Icon }) => (
            <NavLink key={to} to={to} style={{ textDecoration: 'none' }}>
              {({ isActive }) => (
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 11,
                    padding: '9px 10px',
                    borderRadius: 7,
                    cursor: 'pointer',
                    color: isActive ? '#f2f3f4' : '#c7c9cc',
                    background: isActive ? 'rgba(94,106,210,0.12)' : 'transparent',
                  }}
                >
                  <Icon size={18} color={isActive ? 'var(--mantine-color-accent-5)' : '#8a8f98'} />
                  <Text size="md">{label}</Text>
                </div>
              )}
            </NavLink>
          ))}
        </div>

        <Text size="xs" fw={600} c="dark.4" style={{ letterSpacing: '.04em', padding: '4px 10px', marginTop: 4 }}>
          YOUR TEAMS
        </Text>
        {team && (
          <div style={{ display: 'flex', alignItems: 'center', gap: 11, padding: '9px 10px', borderRadius: 7, color: '#c7c9cc' }}>
            <span
              style={{
                width: 20,
                height: 20,
                borderRadius: 5,
                background: 'var(--mantine-color-accent-5)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                fontSize: 11,
                fontWeight: 700,
                color: '#0b0c0e',
                flexShrink: 0,
              }}
            >
              {team.name[0]}
            </span>
            <Text size="md">{team.name}</Text>
          </div>
        )}

        <Menu position="top-start" width={200}>
          <Menu.Target>
            <UnstyledButton
              style={{
                marginTop: 'auto',
                display: 'flex',
                alignItems: 'center',
                gap: 10,
                padding: 10,
                borderTop: '1px solid #1d1e21',
                width: '100%',
              }}
            >
              <Avatar size={30} radius="xl" color={avatarColor(me?.name ?? '')}>
                {me ? initials(me.name) : ''}
              </Avatar>
              <Text size="md" c="dark.2">
                {me?.name}
              </Text>
            </UnstyledButton>
          </Menu.Target>
          <Menu.Dropdown>
            <Menu.Item leftSection={<IconLogout size={16} />} onClick={signOut}>
              Sign out
            </Menu.Item>
          </Menu.Dropdown>
        </Menu>

        <div style={{ display: 'flex', justifyContent: 'center', padding: '6px 10px 2px' }}>
          {versionInfo?.updateAvailable ? (
            <Tooltip label={`Update available: ${versionInfo.latestVersion}`} position="top">
              <UnstyledButton
                component="a"
                href={versionInfo.releaseUrl}
                target="_blank"
                rel="noopener noreferrer"
                style={{ display: 'flex', alignItems: 'center', gap: 6, color: 'var(--mantine-color-accent-5)' }}
              >
                <Text size="xs" c="inherit">
                  {versionInfo.version}
                </Text>
                <IconArrowUpCircle size={13} />
              </UnstyledButton>
            </Tooltip>
          ) : (
            <Text size="xs" c="dark.4">
              {versionInfo?.version ?? ''}
            </Text>
          )}
        </div>
      </MantineAppShell.Navbar>

      <MantineAppShell.Main bg="dark.7">
        <Outlet context={{ team }} />
      </MantineAppShell.Main>

      <CommandPalette teamId={team?.id} />
    </MantineAppShell>
  )
}
