import { createTheme, type MantineColorsTuple } from '@mantine/core'

// Dark ramp pulled directly from the pulled design (see MEMORY.md / plan):
// light end = text tiers, dark end = surface/body backgrounds. Mantine
// treats index 6 as the default elevated-surface bg (Paper/Card/Modal) and
// 7 as the body bg, which lines up with the mockup's own elevation order
// (page darkest, cards/inputs lightest surface).
const dark: MantineColorsTuple = [
  '#f2f3f4', // heading text
  '#e9eaec', // primary text
  '#c7c9cc', // secondary text
  '#8a8f98', // muted text
  '#5a5d63', // faint text
  '#2c2d31', // hover border
  '#141518', // surface bg (cards, inputs, modals)
  '#0b0c0e', // body bg (page)
  '#08090a',
  '#050506',
]

const accent: MantineColorsTuple = [
  '#eef1fc',
  '#d6dbf6',
  '#b3bcee',
  '#8f97e4',
  '#7480da',
  '#5e6ad2',
  '#4c58bb',
  '#3d47a0',
  '#2f3780',
  '#232963',
]

export const theme = createTheme({
  fontFamily:
    "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif",
  defaultRadius: 'md',
  primaryColor: 'accent',
  primaryShade: 5,
  colors: { dark, accent },
  fontSizes: {
    xs: '12px',
    sm: '14px',
    md: '15px',
    lg: '17px',
    xl: '22px',
  },
  spacing: {
    xs: '8px',
    sm: '10px',
    md: '14px',
    lg: '18px',
    xl: '24px',
  },
})

// Status/priority/label colors don't map cleanly onto Mantine's theme-color
// model (a conic-gradient ring, arbitrary per-label bg/fg pairs) — kept as
// plain constants mirroring the mockup's own STATUS_META/PRIORITY_META.
export const STATUS_META = {
  backlog: { label: 'Backlog', color: '#8a8f98', fill: 'transparent', border: 'dashed' },
  todo: { label: 'Todo', color: '#8a8f98', fill: 'transparent', border: 'solid' },
  in_progress: {
    label: 'In Progress',
    color: '#f2c94c',
    fill: 'conic-gradient(#f2c94c 50%, transparent 0)',
    border: 'solid',
  },
  done: { label: 'Done', color: '#5e6ad2', fill: '#5e6ad2', border: 'solid' },
  canceled: { label: 'Canceled', color: '#5a5d63', fill: 'transparent', border: 'solid' },
} as const

export type IssueStatus = keyof typeof STATUS_META
export const STATUS_ORDER: IssueStatus[] = ['backlog', 'todo', 'in_progress', 'done', 'canceled']

// Indexed by the DB's `priority` smallint (0=none,1=urgent,2=high,3=medium,
// 4=low, per internal/db/migrations/00001_init.sql). `bars` is how many of
// the 4 priority bars render filled — urgent is most-filled, matching the
// mockup even though it's the lowest numeric DB value.
export const PRIORITY_META = [
  { label: 'No priority', bars: 0 },
  { label: 'Urgent', bars: 4 },
  { label: 'High', bars: 3 },
  { label: 'Medium', bars: 2 },
  { label: 'Low', bars: 1 },
] as const

export const AVATAR_COLORS = ['#f2c94c', '#5e6ad2', '#4fd1c5', '#eb5757', '#f2994a', '#9b59b6']

export function avatarColor(name: string): string {
  let hash = 0
  for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash)
  return AVATAR_COLORS[Math.abs(hash) % AVATAR_COLORS.length]
}
