import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { CheckCircle2, CircleAlert, Edit3, Eye, FolderOpen, Hash, ImagePlus, RefreshCw, Trash2, UploadCloud, X } from 'lucide-react'
import { apiRequest } from '../../lib/http'

type SMMQueuePageProps = { token: string; onUnauthorized: () => void }

type QueueListItem = {
  folder: string
  ready_marker: string
  queued_at: string
  status: string
}

type QueueMedia = {
  order: number
  file_name: string
  blob_path: string
  content_type: string
  size_bytes: number
}

type QueueManifest = {
  queue_id: string
  folder: string
  created_at: string
  caption: string
  hashtags: string[]
  media: QueueMedia[]
  ready_marker: string
}

type QueueResult = {
  folder: string
  ready_marker: string
  media_count: number
}

const MAX_IMAGES = 10
const MAX_FILE_BYTES = 10 * 1024 * 1024
const ALLOWED_IMAGE_TYPES = new Set(['image/jpeg', 'image/png', 'image/webp', 'image/gif'])

function formatQueueDate(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '—' : date.toLocaleString('en-IN', { day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' })
}

export function SMMQueuePage({ token, onUnauthorized }: SMMQueuePageProps) {
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [caption, setCaption] = useState('')
  const [hashtags, setHashtags] = useState('')
  const [files, setFiles] = useState<File[]>([])
  const [items, setItems] = useState<QueueListItem[]>([])
  const [selectedItem, setSelectedItem] = useState<QueueManifest | null>(null)
  const [editingFolder, setEditingFolder] = useState<string | null>(null)
  const [isComposerOpen, setIsComposerOpen] = useState(false)
  const [isLoadingQueue, setIsLoadingQueue] = useState(true)
  const [isLoadingDetail, setIsLoadingDetail] = useState(false)
  const [isWorking, setIsWorking] = useState(false)
  const [error, setError] = useState('')
  const [fileError, setFileError] = useState('')
  const [notice, setNotice] = useState('')

  const previews = useMemo(() => files.map((file) => ({ file, url: URL.createObjectURL(file) })), [files])

  useEffect(() => () => previews.forEach(({ url }) => URL.revokeObjectURL(url)), [previews])

  const requestJson = useCallback(async <T,>(path: string) => {
    const response = await apiRequest(token, onUnauthorized, path)
    return response.json() as Promise<T>
  }, [onUnauthorized, token])

  const loadQueue = useCallback(async () => {
    setIsLoadingQueue(true)
    setError('')
    try {
      const data = await requestJson<{ success?: boolean; items?: QueueListItem[] }>('/api/smm-queue')
      setItems(data.success && Array.isArray(data.items) ? data.items : [])
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to load the SMM queue.')
    } finally {
      setIsLoadingQueue(false)
    }
  }, [requestJson])

  useEffect(() => { void loadQueue() }, [loadQueue])

  const resetComposer = () => {
    setCaption('')
    setHashtags('')
    setFiles([])
    setEditingFolder(null)
    setFileError('')
    if (fileInputRef.current) fileInputRef.current.value = ''
  }

  const openComposer = () => {
    setError('')
    setNotice('')
    resetComposer()
    setIsComposerOpen(true)
  }

  const addFiles = (selected: FileList | null) => {
    if (!selected?.length) return
    const incoming = Array.from(selected)
    const invalid = incoming.filter((file) => !ALLOWED_IMAGE_TYPES.has(file.type) || file.size > MAX_FILE_BYTES)
    const imageFiles = incoming.filter((file) => ALLOWED_IMAGE_TYPES.has(file.type) && file.size <= MAX_FILE_BYTES)
    const available = Math.max(0, MAX_IMAGES - files.length)

    setFileError(invalid.length ? 'Use JPEG, PNG, WebP, or GIF images up to 10 MB each.' : '')
    if (available === 0) {
      setFileError(`A carousel can contain up to ${MAX_IMAGES} images.`)
      return
    }
    setFiles((current) => [...current, ...imageFiles.slice(0, available)])
    if (imageFiles.length > available) setFileError(`Only the first ${available} additional images were added.`)
  }

  const removeFile = (index: number) => setFiles((current) => current.filter((_, fileIndex) => fileIndex !== index))

  const openDetails = async (folder: string) => {
    setError('')
    setNotice('')
    setIsLoadingDetail(true)
    try {
      const data = await requestJson<{ success?: boolean; queue?: QueueManifest }>(`/api/smm-queue/${encodeURIComponent(folder)}`)
      if (!data.success || !data.queue) throw new Error('Queue item details were not returned.')
      setSelectedItem(data.queue)
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to load queue item details.')
    } finally {
      setIsLoadingDetail(false)
    }
  }

  const editSelectedItem = async () => {
    if (!selectedItem) return
    setError('')
    setIsWorking(true)
    try {
      const existingFiles = await Promise.all([...selectedItem.media].sort((a, b) => a.order - b.order).map(async (media) => {
        const response = await apiRequest(token, onUnauthorized, `/api/smm-queue/${encodeURIComponent(selectedItem.folder)}/media/${encodeURIComponent(media.file_name)}`)
        const blob = await response.blob()
        return new File([blob], media.file_name, { type: media.content_type || blob.type })
      }))
      setCaption(selectedItem.caption)
      setHashtags(selectedItem.hashtags.join(' '))
      setFiles(existingFiles)
      setEditingFolder(selectedItem.folder)
      setSelectedItem(null)
      setIsComposerOpen(true)
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to load this queue item for editing.')
    } finally {
      setIsWorking(false)
    }
  }

  const deleteSelectedItem = async () => {
    if (!selectedItem || !window.confirm(`Delete the queued folder ${selectedItem.folder}?`)) return
    setError('')
    setIsWorking(true)
    try {
      await apiRequest(token, onUnauthorized, `/api/smm-queue/${encodeURIComponent(selectedItem.folder)}`, { method: 'DELETE' })
      setSelectedItem(null)
      setNotice('Queue item deleted.')
      await loadQueue()
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to delete this queue item.')
    } finally {
      setIsWorking(false)
    }
  }

  const submitQueue = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setError('')
    setNotice('')
    if (!caption.trim()) {
      setError('Add a caption before saving this post.')
      return
    }
    if (!files.length) {
      setError('Add at least one image before saving this post.')
      return
    }

    const formData = new FormData()
    formData.append('caption', caption.trim())
    formData.append('hashtags', hashtags.trim())
    files.forEach((file) => formData.append('media', file, file.name))
    const isEditing = Boolean(editingFolder)
    const path = isEditing ? `/api/smm-queue/${encodeURIComponent(editingFolder || '')}` : '/api/smm-queue'

    setIsWorking(true)
    try {
      const response = await apiRequest(token, onUnauthorized, path, { method: isEditing ? 'PUT' : 'POST', body: formData })
      const data = await response.json() as { success?: boolean; queue?: QueueResult }
      if (!data.success || !data.queue) throw new Error('The saved queue item was not returned by the server.')
      setIsComposerOpen(false)
      resetComposer()
      setNotice(isEditing ? `Queue item replaced with ${data.queue.folder}.` : `Added ${data.queue.folder} to the SMM queue.`)
      await loadQueue()
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Unable to save this queue item.')
    } finally {
      setIsWorking(false)
    }
  }

  const closeComposer = () => {
    if (isWorking) return
    setIsComposerOpen(false)
    resetComposer()
  }

  return (
    <section className="workspace-page smm-queue-page" aria-labelledby="smm-queue-heading">
      <header className="workspace-page-header">
        <div>
          <p className="eyebrow">Growth / Content operations</p>
          <h2 id="smm-queue-heading">SMM Queue</h2>
          <p>Prepare carousel content once and leave it in the publishing queue for n8n to distribute across your social channels.</p>
        </div>
        <div className="smm-queue-header-actions">
          <span className="smm-queue-storage-badge"><UploadCloud size={15} aria-hidden="true" /> Azure queue</span>
          <button className="primary-button" type="button" onClick={openComposer}><ImagePlus size={16} aria-hidden="true" /> Add to Queue</button>
        </div>
      </header>

      {error && <div className="dashboard-error" role="alert"><CircleAlert size={18} aria-hidden="true" /><span>{error}</span><button type="button" onClick={() => void loadQueue()}>Try again</button></div>}
      {notice && <div className="smm-queue-notice" role="status"><CheckCircle2 size={17} aria-hidden="true" /><span>{notice}</span></div>}

      <section className="reports-table-card smm-queue-list-card" aria-labelledby="smm-queue-list-heading">
        <div className="orders-card-heading">
          <div><p className="eyebrow">Ready folders</p><h3 id="smm-queue-list-heading">{items.length} queued {items.length === 1 ? 'item' : 'items'}</h3></div>
          <button className="secondary-button" type="button" onClick={() => void loadQueue()} disabled={isLoadingQueue}><RefreshCw size={15} className={isLoadingQueue ? 'spin' : undefined} aria-hidden="true" /> Refresh</button>
        </div>
        <div className="orders-table-wrap">
          <table className="orders-table smm-queue-table">
            <caption className="sr-only">SMM Queue ready folders</caption>
            <thead><tr><th>Upload folder</th><th>Queued at</th><th>Status</th><th>Open</th></tr></thead>
            <tbody>
              {isLoadingQueue ? <tr><td className="table-state" colSpan={4}>Loading queue folders…</td></tr> : items.length === 0 ? <tr><td className="table-state" colSpan={4}><div className="smm-queue-empty"><FolderOpen size={20} aria-hidden="true" /><strong>No ready folders yet</strong><span>Add a carousel to create the first timestamped queue folder.</span><button className="secondary-button" type="button" onClick={openComposer}><ImagePlus size={15} aria-hidden="true" /> Add to Queue</button></div></td></tr> : items.map((item) => <tr key={item.folder}><td><div className="smm-queue-folder-cell"><FolderOpen size={16} aria-hidden="true" /><strong title={item.folder}>{item.folder}</strong><small>{item.ready_marker}</small></div></td><td>{formatQueueDate(item.queued_at)}</td><td><span className="status-pill status-pill-success">Queued</span></td><td><button className="table-link-button" type="button" onClick={() => void openDetails(item.folder)} disabled={isLoadingDetail}><Eye size={14} aria-hidden="true" /> View</button></td></tr>)}
            </tbody>
          </table>
        </div>
      </section>

      {isComposerOpen && <div className="modal-scrim" role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) closeComposer() }}><form className="modal-card smm-queue-modal" role="dialog" aria-modal="true" aria-labelledby="smm-queue-modal-heading" onSubmit={submitQueue} onMouseDown={(event) => event.stopPropagation()}>
        <div className="modal-heading"><div><p className="eyebrow">SMM Queue / {editingFolder ? 'Edit item' : 'New item'}</p><h2 id="smm-queue-modal-heading">{editingFolder ? 'Edit queued post' : 'Add to SMM Queue'}</h2></div><button className="icon-button" type="button" aria-label="Close queue composer" onClick={closeComposer}><X size={19} aria-hidden="true" /></button></div>
        <div className="smm-queue-modal-copy"><span>Shared post copy</span><small>The same caption and hashtags will be available to every connected platform.</small></div>
        <section className="smm-queue-modal-section" aria-labelledby="smm-modal-media-heading"><div className="smm-queue-card-heading"><div><p className="eyebrow">Carousel</p><h3 id="smm-modal-media-heading">Images in publishing order</h3></div><span>{files.length}/{MAX_IMAGES}</span></div><button className="smm-upload-dropzone" type="button" onClick={() => fileInputRef.current?.click()}><span className="smm-upload-icon"><ImagePlus size={23} aria-hidden="true" /></span><strong>{files.length ? 'Add more images' : 'Add product or campaign images'}</strong><small>JPEG, PNG, WebP, or GIF · up to {MAX_IMAGES} images · 10 MB each</small><span className="secondary-button"><UploadCloud size={15} aria-hidden="true" /> Choose images</span></button><input ref={fileInputRef} className="smm-visually-hidden" type="file" accept="image/jpeg,image/png,image/webp,image/gif" multiple onChange={(event) => { addFiles(event.target.files); event.currentTarget.value = '' }} />{fileError && <p className="smm-queue-field-error" role="alert">{fileError}</p>}{previews.length > 0 && <div className="smm-preview-grid">{previews.map(({ file, url }, index) => <div className="smm-preview-card" key={`${file.name}-${file.lastModified}-${index}`}><img src={url} alt={`Carousel image ${index + 1}: ${file.name}`} /><span className="smm-preview-order">{index + 1}</span><button className="smm-preview-remove" type="button" aria-label={`Remove ${file.name}`} onClick={() => removeFile(index)}><X size={15} aria-hidden="true" /></button><small>{file.name}</small></div>)}</div>}</section>
        <section className="smm-queue-modal-section" aria-labelledby="smm-modal-copy-heading"><div className="smm-queue-card-heading"><div><p className="eyebrow">Post copy</p><h3 id="smm-modal-copy-heading">Caption and hashtags</h3></div><span>Shared</span></div><label className="form-field" htmlFor="smm-caption"><span>Caption</span><textarea id="smm-caption" value={caption} onChange={(event) => setCaption(event.target.value)} placeholder="Write the caption that should accompany this carousel…" rows={5} maxLength={5000} required /></label><label className="form-field" htmlFor="smm-hashtags"><span><Hash size={15} aria-hidden="true" /> Hashtags <small>Optional</small></span><input id="smm-hashtags" value={hashtags} onChange={(event) => setHashtags(event.target.value)} placeholder="#millennialperfumer #newlaunch" /></label><p className="smm-queue-help">Separate hashtags with spaces or commas. The queue will normalize each tag with a leading #.</p></section>
        <div className="modal-actions"><button className="secondary-button" type="button" onClick={closeComposer}>Cancel</button><button className="primary-button" type="submit" disabled={isWorking}>{isWorking ? 'Saving…' : editingFolder ? 'Save changes' : 'Add to Queue'} <UploadCloud size={16} aria-hidden="true" /></button></div>
      </form></div>}

      {selectedItem && <div className="modal-scrim" role="presentation" onMouseDown={(event) => { if (event.target === event.currentTarget) setSelectedItem(null) }}><section className="modal-card smm-queue-detail-modal" role="dialog" aria-modal="true" aria-labelledby="smm-queue-detail-heading" onMouseDown={(event) => event.stopPropagation()}>
        <div className="modal-heading"><div><p className="eyebrow">Queued folder</p><h2 id="smm-queue-detail-heading">{selectedItem.folder}</h2></div><button className="icon-button" type="button" aria-label="Close queue item details" onClick={() => setSelectedItem(null)}><X size={19} aria-hidden="true" /></button></div>
        <div className="smm-queue-detail-status"><span className="status-pill status-pill-success">Queued</span><span>{selectedItem.media.length} image{selectedItem.media.length === 1 ? '' : 's'} · Added {formatQueueDate(selectedItem.created_at)}</span></div>
        <section className="smm-queue-detail-section"><span className="metric-label">Caption</span><p>{selectedItem.caption}</p><span className="metric-label">Hashtags</span><p>{selectedItem.hashtags.length ? selectedItem.hashtags.join(' ') : 'No hashtags added.'}</p></section>
        <section className="smm-queue-detail-section"><span className="metric-label">Carousel order</span><ol className="smm-queue-media-list">{[...selectedItem.media].sort((a, b) => a.order - b.order).map((media) => <li key={media.blob_path}><span>{media.order}</span><strong>{media.file_name}</strong><small>{Math.ceil(media.size_bytes / 1024)} KB</small></li>)}</ol></section>
        <p className="smm-queue-detail-path">Ready marker: <code>{selectedItem.ready_marker}</code></p>
        <div className="modal-actions"><button className="secondary-button danger-link" type="button" onClick={() => void deleteSelectedItem()} disabled={isWorking}><Trash2 size={14} aria-hidden="true" /> Delete</button><div className="smm-queue-detail-actions"><button className="secondary-button" type="button" onClick={() => void editSelectedItem()} disabled={isWorking}><Edit3 size={14} aria-hidden="true" /> {isWorking ? 'Loading…' : 'Edit'}</button><button className="primary-button" type="button" onClick={() => setSelectedItem(null)}>Done</button></div></div>
      </section></div>}
    </section>
  )
}
