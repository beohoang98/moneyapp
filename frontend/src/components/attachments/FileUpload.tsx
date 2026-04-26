import { useState, useRef, type DragEvent, type ChangeEvent } from 'react'
import { uploadAttachment } from '../../api/attachments'
import './Attachments.css'

const MAX_FILE_SIZE = 10 * 1024 * 1024
const ALLOWED_TYPES = ['application/pdf', 'image/jpeg', 'image/png']
const ALLOWED_EXTENSIONS = ['.pdf', '.jpg', '.jpeg', '.png']

interface FileUploadProps {
  entityType: string
  entityId: number
  onUploaded: () => void
}

export function FileUpload({ entityType, entityId, onUploaded }: FileUploadProps) {
  const [dragOver, setDragOver] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [progress, setProgress] = useState(0)
  const [error, setError] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)

  const validateFile = (file: File): string | null => {
    if (file.size > MAX_FILE_SIZE) return 'File exceeds the 10 MB limit'
    const ext = '.' + file.name.split('.').pop()?.toLowerCase()
    if (!ALLOWED_TYPES.includes(file.type) && !ALLOWED_EXTENSIONS.includes(ext)) {
      return 'Only PDF, JPEG, and PNG files are allowed'
    }
    return null
  }

  const handleFile = async (file: File) => {
    const validationError = validateFile(file)
    if (validationError) {
      setError(validationError)
      return
    }

    setError('')
    setUploading(true)
    setProgress(0)

    try {
      await uploadAttachment(entityType, entityId, file, setProgress)
      onUploaded()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Upload failed')
    } finally {
      setUploading(false)
      setProgress(0)
    }
  }

  const handleDrop = (e: DragEvent) => {
    e.preventDefault()
    setDragOver(false)
    const file = e.dataTransfer.files[0]
    if (file) handleFile(file)
  }

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) handleFile(file)
    if (inputRef.current) inputRef.current.value = ''
  }

  return (
    <div className="file-upload">
      <div
        className={`file-upload__dropzone ${dragOver ? 'file-upload__dropzone--active' : ''}`}
        onDragOver={(e) => { e.preventDefault(); setDragOver(true) }}
        onDragLeave={() => setDragOver(false)}
        onDrop={handleDrop}
        onClick={() => inputRef.current?.click()}
      >
        <input
          ref={inputRef}
          type="file"
          accept=".pdf,.jpg,.jpeg,.png"
          onChange={handleChange}
          hidden
        />
        {uploading ? (
          <div className="file-upload__progress">
            <div className="file-upload__progress-bar" style={{ width: `${progress}%` }} />
            <span>{progress}%</span>
          </div>
        ) : (
          <span className="file-upload__label">
            Drop file here or click to upload
            <small>PDF, JPEG, PNG — max 10 MB</small>
          </span>
        )}
      </div>
      {error && <div className="file-upload__error">{error}</div>}
    </div>
  )
}
