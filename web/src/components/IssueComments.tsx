import { useState } from 'react'
import { Avatar, Button, Group, Loader, Stack, Text, Textarea } from '@mantine/core'
import { useAddIssueComment, useIssueComments } from '../lib/api/hooks'
import { avatarColor } from '../theme'

export function IssueComments({ issueId }: { issueId: string }) {
  const { data: comments, isLoading } = useIssueComments(issueId)
  const addComment = useAddIssueComment()
  const [draft, setDraft] = useState('')

  function submit() {
    const body = draft.trim()
    if (!body) return
    addComment.mutate({ id: issueId, body }, { onSuccess: () => setDraft('') })
  }

  return (
    <Stack gap="md">
      {isLoading && <Loader size="sm" />}
      {!isLoading && !comments?.length && <Text size="sm" c="dark.4">No comments yet.</Text>}
      <Stack gap={0}>
        {comments?.map((comment) => (
          <div key={comment.id} style={{ display: 'flex', gap: 12, padding: '14px 0', borderBottom: '1px solid #1d1e21' }}>
            <Avatar size={28} radius="xl" color={avatarColor(comment.authorName)}>
              {comment.authorName.slice(0, 1).toUpperCase()}
            </Avatar>
            <div style={{ minWidth: 0, flex: 1 }}>
              <Group gap={6} justify="space-between" align="flex-start">
                <Text size="sm" fw={600} c="dark.1">{comment.authorName}</Text>
                <Text size="xs" c="dark.4" style={{ whiteSpace: 'nowrap' }}>
                  {new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(comment.createdAt))}
                </Text>
              </Group>
              <Text size="sm" c="dark.1" mt={4} style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
                {comment.body}
              </Text>
            </div>
          </div>
        ))}
      </Stack>
      <Textarea
        placeholder="Leave a comment…"
        value={draft}
        onChange={(e) => setDraft(e.currentTarget.value)}
        onKeyDown={(e) => {
          if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) submit()
        }}
        autosize
        minRows={3}
      />
      <Group justify="flex-end">
        <Button size="xs" onClick={submit} disabled={!draft.trim()} loading={addComment.isPending}>
          Comment
        </Button>
      </Group>
    </Stack>
  )
}
