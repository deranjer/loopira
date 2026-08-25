import { useMe } from '../lib/api/hooks'
import { IssuesView } from './IssuesView'

export function MyIssuesView() {
  const { data: me } = useMe()
  return <IssuesView assigneeId={me?.id} title="My Issues" />
}
