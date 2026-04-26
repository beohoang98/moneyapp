import { apiClient } from './client'
import type { Attachment } from '../types/attachment'

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api'
const TOKEN_KEY = 'moneyapp_token'

interface AttachmentListResponse {
  data: Attachment[]
}

export async function uploadAttachment(
  entityType: string,
  entityId: number,
  file: File,
  onProgress?: (percent: number) => void,
): Promise<Attachment> {
  const formData = new FormData()
  formData.append('entity_type', entityType)
  formData.append('entity_id', String(entityId))
  formData.append('file', file)

  if (onProgress) {
    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest()
      xhr.open('POST', `${BASE_URL}/attachments`)
      const token = localStorage.getItem(TOKEN_KEY)
      if (token) xhr.setRequestHeader('Authorization', `Bearer ${token}`)

      xhr.upload.addEventListener('progress', (e) => {
        if (e.lengthComputable) onProgress(Math.round((e.loaded / e.total) * 100))
      })

      xhr.addEventListener('load', () => {
        if (xhr.status === 201) {
          const body = JSON.parse(xhr.responseText)
          resolve(body.data)
        } else {
          try {
            const body = JSON.parse(xhr.responseText)
            reject(new Error(body.error || 'Upload failed'))
          } catch {
            reject(new Error('Upload failed'))
          }
        }
      })

      xhr.addEventListener('error', () => reject(new Error('Upload failed')))
      xhr.send(formData)
    })
  }

  const res = await apiClient.upload<{ data: Attachment }>('/attachments', formData)
  return res.data
}

export async function getAttachments(entityType: string, entityId: number): Promise<Attachment[]> {
  const res = await apiClient.get<AttachmentListResponse>(
    `/attachments?entity_type=${entityType}&entity_id=${entityId}`,
  )
  return res.data
}

export async function deleteAttachment(id: number): Promise<void> {
  await apiClient.delete(`/attachments/${id}`)
}

export function getDownloadUrl(id: number): string {
  return `${BASE_URL}/attachments/${id}/download`
}

export function getPreviewUrl(id: number): string {
  return `${BASE_URL}/attachments/${id}/preview`
}
