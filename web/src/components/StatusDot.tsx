import { STATUS_META, type IssueStatus } from '../theme'

export function StatusDot({ status, size = 16 }: { status: IssueStatus; size?: number }) {
  const meta = STATUS_META[status]
  return (
    <span
      style={{
        width: size,
        height: size,
        borderRadius: '50%',
        border: `1.5px ${meta.border} ${meta.color}`,
        background: meta.fill,
        flexShrink: 0,
        display: 'inline-block',
      }}
    />
  )
}
