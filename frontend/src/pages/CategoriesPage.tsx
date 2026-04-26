import { useEffect, useState } from 'react'
import { getAllCategories, createCategory, updateCategory, deleteCategory } from '../api/categories'
import { ConfirmDialog } from '../components/ConfirmDialog'
import { useToast } from '../hooks/useToast'
import type { Category } from '../types/category'
import './Pages.css'

export function CategoriesPage() {
  const [categories, setCategories] = useState<Category[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState<'expense' | 'income' | null>(null)
  const [newName, setNewName] = useState('')
  const [editTarget, setEditTarget] = useState<Category | null>(null)
  const [editName, setEditName] = useState('')
  const [deleteTarget, setDeleteTarget] = useState<Category | null>(null)
  const [formError, setFormError] = useState('')
  const { addToast } = useToast()

  const loadCategories = () => {
    setLoading(true)
    getAllCategories()
      .then(setCategories)
      .catch(() => addToast('Failed to load categories', 'error'))
      .finally(() => setLoading(false))
  }

  useEffect(() => { loadCategories() }, []) // eslint-disable-line react-hooks/exhaustive-deps

  const expenseCategories = categories.filter((c) => c.type === 'expense')
  const incomeCategories = categories.filter((c) => c.type === 'income')

  const handleCreate = async () => {
    if (!showForm || !newName.trim()) return
    setFormError('')
    try {
      await createCategory(newName.trim(), showForm)
      setNewName('')
      setShowForm(null)
      addToast('Category created', 'success')
      loadCategories()
    } catch (err) {
      setFormError(err instanceof Error ? err.message : 'Failed to create')
    }
  }

  const handleUpdate = async () => {
    if (!editTarget || !editName.trim()) return
    setFormError('')
    try {
      await updateCategory(editTarget.id, editName.trim())
      setEditTarget(null)
      setEditName('')
      addToast('Category renamed', 'success')
      loadCategories()
    } catch (err) {
      setFormError(err instanceof Error ? err.message : 'Failed to rename')
    }
  }

  const handleDelete = async () => {
    if (!deleteTarget) return
    try {
      await deleteCategory(deleteTarget.id)
      setDeleteTarget(null)
      addToast('Category deleted', 'success')
      loadCategories()
    } catch (err) {
      addToast(err instanceof Error ? err.message : 'Failed to delete', 'error')
    }
  }

  const renderSection = (title: string, cats: Category[], type: 'expense' | 'income') => (
    <div style={{ marginBottom: 32 }}>
      <div className="page-header">
        <h2 style={{ margin: 0 }}>{title}</h2>
        <button className="btn btn-primary btn-sm" onClick={() => { setShowForm(type); setNewName(''); setFormError('') }}>
          + Add Category
        </button>
      </div>

      {showForm === type && (
        <div style={{ display: 'flex', gap: 8, alignItems: 'center', marginBottom: 12 }}>
          <input
            type="text"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            placeholder="Category name"
            style={{ flex: 1, padding: '6px 10px', borderRadius: 6, border: '1px solid var(--border)' }}
            autoFocus
            onKeyDown={(e) => e.key === 'Enter' && handleCreate()}
          />
          <button className="btn btn-sm btn-primary" onClick={handleCreate}>Save</button>
          <button className="btn btn-sm" onClick={() => setShowForm(null)}>Cancel</button>
          {formError && <span style={{ color: '#ef4444', fontSize: 13 }}>{formError}</span>}
        </div>
      )}

      <table className="data-table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Type</th>
            <th style={{ textAlign: 'right' }}>Actions</th>
          </tr>
        </thead>
        <tbody>
          {cats.map((cat) => (
            <tr key={cat.id}>
              <td>
                {editTarget?.id === cat.id ? (
                  <div style={{ display: 'flex', gap: 4, alignItems: 'center' }}>
                    <input
                      type="text"
                      value={editName}
                      onChange={(e) => setEditName(e.target.value)}
                      style={{ padding: '4px 8px', borderRadius: 4, border: '1px solid var(--border)', width: 160 }}
                      autoFocus
                      onKeyDown={(e) => e.key === 'Enter' && handleUpdate()}
                    />
                    <button className="btn btn-sm btn-primary" onClick={handleUpdate}>Save</button>
                    <button className="btn btn-sm" onClick={() => setEditTarget(null)}>Cancel</button>
                    {formError && <span style={{ color: '#ef4444', fontSize: 12 }}>{formError}</span>}
                  </div>
                ) : (
                  <span>
                    {cat.name}
                    {cat.is_default && <span style={{ marginLeft: 8, fontSize: 11, color: 'var(--text)', opacity: 0.6 }}>default</span>}
                  </span>
                )}
              </td>
              <td>{cat.type}</td>
              <td className="actions">
                {!cat.is_default && (
                  <>
                    <button className="btn btn-sm" onClick={() => { setEditTarget(cat); setEditName(cat.name); setFormError('') }}>Rename</button>
                    <button className="btn btn-sm btn-danger" onClick={() => setDeleteTarget(cat)}>Delete</button>
                  </>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )

  if (loading) return <div className="loading">Loading...</div>

  return (
    <div>
      <h1>Categories</h1>
      {renderSection('Expense Categories', expenseCategories, 'expense')}
      {renderSection('Income Categories', incomeCategories, 'income')}

      <ConfirmDialog
        open={!!deleteTarget}
        title="Delete Category"
        message={`Delete "${deleteTarget?.name}"? All transactions in this category will be moved to "Uncategorized".`}
        confirmLabel="Delete"
        onConfirm={handleDelete}
        onCancel={() => setDeleteTarget(null)}
        danger
      />
    </div>
  )
}
