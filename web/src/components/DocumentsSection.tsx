import { useRef } from 'react'
import { ActionIcon, Anchor, Button, Group, Stack, Text } from '@mantine/core'
import { IconDownload, IconTrash, IconUpload } from '@tabler/icons-react'
import { documentsApi } from '../lib/api/client'
import { useDeleteDocument, useProjectDocuments, useUploadDocument } from '../lib/api/hooks'

function formatSize(bytes: number) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
}

export function DocumentsSection({ projectId }: { projectId: string }) {
  const { data: documents } = useProjectDocuments(projectId)
  const uploadDocument = useUploadDocument()
  const deleteDocument = useDeleteDocument()
  const fileInputRef = useRef<HTMLInputElement>(null)

  function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    uploadDocument.mutate({ projectId, file })
    e.target.value = ''
  }

  return (
    <div>
      <Group justify="space-between" mb="sm">
        <Text size="lg" fw={600} c="dark.0">
          Documents
        </Text>
        <Button
          size="sm"
          variant="light"
          leftSection={<IconUpload size={14} />}
          onClick={() => fileInputRef.current?.click()}
          loading={uploadDocument.isPending}
        >
          Upload
        </Button>
        <input ref={fileInputRef} type="file" style={{ display: 'none' }} onChange={handleFileChange} />
      </Group>

      <Stack gap="xs">
        {(documents ?? []).map((doc) => (
          <Group key={doc.id} justify="space-between">
            <div style={{ minWidth: 0 }}>
              <Anchor href={documentsApi.downloadUrl(doc.id)} size="sm" c="dark.1" style={{ display: 'block' }}>
                {doc.filename}
              </Anchor>
              <Text size="xs" c="dark.4">
                {formatSize(doc.sizeBytes)} · {doc.uploadedByName} · {formatDate(doc.createdAt)}
              </Text>
            </div>
            <Group gap={4}>
              <ActionIcon
                component="a"
                href={documentsApi.downloadUrl(doc.id)}
                variant="subtle"
                size="sm"
              >
                <IconDownload size={14} />
              </ActionIcon>
              <ActionIcon
                variant="subtle"
                color="red"
                size="sm"
                onClick={() => deleteDocument.mutate({ attachmentId: doc.id, projectId })}
                loading={deleteDocument.isPending && deleteDocument.variables?.attachmentId === doc.id}
              >
                <IconTrash size={14} />
              </ActionIcon>
            </Group>
          </Group>
        ))}
        {documents?.length === 0 && (
          <Text size="sm" c="dark.4">
            No documents yet.
          </Text>
        )}
      </Stack>
    </div>
  )
}
