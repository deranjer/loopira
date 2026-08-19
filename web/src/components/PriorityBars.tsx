import { PRIORITY_META } from '../theme'

export function PriorityBars({ priority }: { priority: number }) {
  const filled = PRIORITY_META[priority]?.bars ?? 0
  return (
    <span style={{ display: 'flex', gap: 3, alignItems: 'flex-end', height: 16 }}>
      {[0, 1, 2, 3].map((i) => (
        <span
          key={i}
          style={{
            width: 4,
            height: 6 + i * 4,
            background: i < filled ? '#c7c9cc' : '#2a2b2f',
            borderRadius: 1,
          }}
        />
      ))}
    </span>
  )
}
