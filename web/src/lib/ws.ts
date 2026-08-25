import { useEffect } from 'react'
import { useQueryClient } from '@tanstack/react-query'

interface WsEvent {
  type: string
  teamId: string
  payload: unknown
}

// Opens the app's websocket for a team and invalidates the issue queries
// whenever the backend broadcasts a mutation, so other open tabs/clients
// pick up changes without polling.
export function useTeamLiveUpdates(teamId: string | undefined) {
  const qc = useQueryClient()

  useEffect(() => {
    if (!teamId) return

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const socket = new WebSocket(`${protocol}//${window.location.host}/ws?team=${teamId}`)

    socket.onmessage = (event) => {
      let msg: WsEvent
      try {
        msg = JSON.parse(event.data)
      } catch {
        return
      }
      if (msg.type.startsWith('issue.')) {
        qc.invalidateQueries({ queryKey: ['issues'] })
      } else if (msg.type.startsWith('worklog.')) {
        qc.invalidateQueries({ queryKey: ['projectWorkLogs'] })
        qc.invalidateQueries({ queryKey: ['workLogs'] })
      }
    }

    return () => socket.close()
  }, [teamId, qc])
}
