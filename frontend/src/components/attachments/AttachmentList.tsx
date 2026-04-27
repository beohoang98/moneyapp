import { useEffect, useState, useRef } from 'react'
import { getAttachments, deleteAttachment, getDownloadUrl, getPreviewUrl } from '../../api/attachments'
import { TOKEN_KEY } from '../../api/client'
import { ConfirmDialog } from '../ConfirmDialog'
import { useToast } from '../../hooks/useToast'
import type { Attachment } from '../../types/attachment'
import './Attachments.css'

interface AttachmentListProps {
  entityType: string
  entityId: number
  refreshKey?: number
}

function revokeMap(map: Map<number, string>) {
  map.forEach((url) => URL.revokeObjectURL(url))
}

export function AttachmentList({ entityType, entityId, refreshKey }: AttachmentListProps) {
  const [attachments, setAttachments] = useState<Attachment[]>([])
  const [previewBlobs, setPreviewBlobs] = useState<Map<number, string>>(new Map())
  const [loading, setLoading] = useState(true)
  const [deleteTarget, setDeleteTarget] = useState<Attachment | null>(null)
  const [lightboxSrc, setLightboxSrc] = useState<string | null>(null)
  const blobsRef = useRef<Map<number, string>>(new Map())
  // Revoke-on-unmount only: revoking while the PDF tab is still open blanks it.
  const pdfBlobsRef = useRef<Set<string>>(new Set())
  const { addToast } = useToast()

  useEffect(() => {
    let cancelled = false
    const pdfUrls = pdfBlobsRef.current
    revokeMap(blobsRef.current)
    blobsRef.current = new Map()
    queueMicrotask(() => {
      if (cancelled) return
      setPreviewBlobs(new Map())
      setLoading(true)
    })

    const token = localStorage.getItem(TOKEN_KEY)

    getAttachments(entityType, entityId)
      .then(async (list) => {
        if (cancelled) return
        setAttachments(list)
        if (!token) {
          return
        }
        const map = new Map<number, string>()
        const images = list.filter((a) => a.mime_type.startsWith('image/'))
        await Promise.all(
          images.map(async (att) => {
            try {
              const r = await fetch(getPreviewUrl(att.id), {
                headers: { Authorization: `Bearer ${token}` },
              })
              if (!r.ok || cancelled) return
              const url = URL.createObjectURL(await r.blob())
              map.set(att.id, url)
            } catch {
              /* skip broken preview */
            }
          }),
        )
        if (cancelled) {
          revokeMap(map)
          return
        }
        blobsRef.current = map
        setPreviewBlobs(new Map(map))
      })
      .catch(() => {
        if (!cancelled) setAttachments([])
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
      revokeMap(blobsRef.current)
      blobsRef.current = new Map()
      pdfUrls.forEach((url) => URL.revokeObjectURL(url))
      pdfUrls.clear()
    }
  }, [entityType, entityId, refreshKey])

  const handleDelete = async () => {
    if (!deleteTarget) return
    try {
      await deleteAttachment(deleteTarget.id)
      const blobUrl = blobsRef.current.get(deleteTarget.id)
      if (blobUrl) {
        URL.revokeObjectURL(blobUrl)
        blobsRef.current.delete(deleteTarget.id)
        setPreviewBlobs(new Map(blobsRef.current))
      }
      setAttachments((prev) => prev.filter((a) => a.id !== deleteTarget.id))
      setDeleteTarget(null)
    } catch {
      addToast('Failed to delete attachment', 'error')
      setDeleteTarget(null)
    }
  }

  const formatSize = (bytes: number) => {
    if (bytes < 1024) return bytes + ' B'
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
  }

  const handleDownload = (att: Attachment) => {
    const token = localStorage.getItem(TOKEN_KEY)
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
                previewBlobs.has(att.id) ? (
                  <img
                    src={previewBlobs.get(att.id)}
                    alt={att.filename}
                    className="attachment-item__thumb"
                    onClick={() => setLightboxSrc(previewBlobs.get(att.id) ?? null)}
                  />
                ) : (
                  <div
                    className="attachment-item__thumb-placeholder"
                    title="Loading preview"
                    aria-hidden
                  />
                )
              ) : (
                <div
                  className="attachment-item__icon"
                  onClick={() => {
                    const t = localStorage.getItem(TOKEN_KEY)
                    const url = getPreviewUrl(att.id)
                    if (t) {
                      fetch(url, { headers: { Authorization: `Bearer ${t}` } })
                        .then((r) => r.blob())
                        .then((blob) => {
                          const blobUrl = URL.createObjectURL(blob)
                          pdfBlobsRef.current.add(blobUrl)
                          window.open(blobUrl)
                        })
                    } else {
                      window.open(url)
                    }
                  }}
                >
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
