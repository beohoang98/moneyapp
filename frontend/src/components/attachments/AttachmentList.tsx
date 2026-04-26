import { useEffect, useState } from 'react'
import { getAttachments, deleteAttachment, getDownloadUrl, getPreviewUrl } from '../../api/attachments'
import { ConfirmDialog } from '../ConfirmDialog'
import type { Attachment } from '../../types/attachment'
import './Attachments.css'

interface AttachmentListProps {
  entityType: string
  entityId: number
  refreshKey?: number
}

export function AttachmentList({ entityType, entityId, refreshKey }: AttachmentListProps) {
  const [attachments, setAttachments] = useState<Attachment[]>([])
  const [loading, setLoading] = useState(true)
  const [deleteTarget, setDeleteTarget] = useState<Attachment | null>(null)
  const [lightboxSrc, setLightboxSrc] = useState<string | null>(null)

  useEffect(() => {
    setLoading(true)
    getAttachments(entityType, entityId)
      .then(setAttachments)
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [entityType, entityId, refreshKey])

  const handleDelete = async () => {
    if (!deleteTarget) return
    await deleteAttachment(deleteTarget.id)
    setAttachments((prev) => prev.filter((a) => a.id !== deleteTarget.id))
    setDeleteTarget(null)
  }

  const formatSize = (bytes: number) => {
    if (bytes < 1024) return bytes + ' B'
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
  }

  const handleDownload = (att: Attachment) => {
    const token = localStorage.getItem('moneyapp_token')
    const url = getDownloadUrl(att.id)

    const a = document.createElement('a')
    if (token) {
      fetch(url, { headers: { Authorization: `Bearer ${token}` } })
        .then((r) => r.blob())
        .then((blob) => {
          const blobUrl = URL.createObjectURL(blob)
          a.href = blobUrl
          a.download = att.filename
          a.click()
          URL.revokeObjectURL(blobUrl)
        })
    } else {
      a.href = url
      a.download = att.filename
      a.click()
    }
  }

  if (loading) return <div className="attachments-loading">Loading attachments...</div>
  if (attachments.length === 0) return null

  return (
    <div className="attachment-list">
      <h4 className="attachment-list__title">Attachments ({attachments.length})</h4>
      <div className="attachment-list__items">
        {attachments.map((att) => (
          <div key={att.id} className="attachment-item">
            <div className="attachment-item__preview">
              {att.mime_type.startsWith('image/') ? (
                <img
                  src={getPreviewUrl(att.id)}
                  alt={att.filename}
                  className="attachment-item__thumb"
                  onClick={() => setLightboxSrc(getPreviewUrl(att.id))}
                />
              ) : (
                <div className="attachment-item__icon" onClick={() => {
                  const token = localStorage.getItem('moneyapp_token')
                  const url = getPreviewUrl(att.id)
                  if (token) {
                    fetch(url, { headers: { Authorization: `Bearer ${token}` } })
                      .then((r) => r.blob())
                      .then((blob) => {
                        const blobUrl = URL.createObjectURL(blob)
                        window.open(blobUrl)
                      })
                  } else {
                    window.open(url)
                  }
                }}>
                  PDF
                </div>
              )}
            </div>
            <div className="attachment-item__info">
              <span className="attachment-item__name">{att.filename}</span>
              <span className="attachment-item__size">{formatSize(att.size_bytes)}</span>
            </div>
            <div className="attachment-item__actions">
              <button className="btn btn-sm" onClick={() => handleDownload(att)} title="Download">
                Download
              </button>
              <button className="btn btn-sm btn-danger" onClick={() => setDeleteTarget(att)} title="Delete">
                Delete
              </button>
            </div>
          </div>
        ))}
      </div>

      {lightboxSrc && (
        <div className="lightbox" onClick={() => setLightboxSrc(null)}>
          <img src={lightboxSrc} alt="Preview" className="lightbox__image" />
        </div>
      )}

      <ConfirmDialog
        open={!!deleteTarget}
        title="Delete Attachment"
        message={`Delete "${deleteTarget?.filename}"?`}
        confirmLabel="Delete"
        onConfirm={handleDelete}
        onCancel={() => setDeleteTarget(null)}
        danger
      />
    </div>
  )
}
