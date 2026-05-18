import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { api } from './api'

const PAGE_SIZE = 5

type Job = {
  id: number
  title: string
  description: string
  status: string
  created_at?: string
}

type JobTab = 'active' | 'archived'

function formatDate(value?: string) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function JobsPage() {
  const [jobs, setJobs] = useState<Job[]>([])
  const [tab, setTab] = useState<JobTab>('active')
  const [page, setPage] = useState(1)
  const [title, setTitle] = useState('')
  const [desc, setDesc] = useState('')
  const [err, setErr] = useState('')
  const [editingId, setEditingId] = useState<number | null>(null)
  const [editTitle, setEditTitle] = useState('')
  const [editDesc, setEditDesc] = useState('')

  async function load() {
    const r = await api<{ jobs: Job[] }>('/api/hr/jobs?page=1&page_size=50')
    setJobs(r.jobs || [])
  }

  useEffect(() => {
    api<{ jobs: Job[] }>('/api/hr/jobs?page=1&page_size=50')
      .then((r) => setJobs(r.jobs || []))
      .catch((e) => setErr(String(e)))
  }, [])

  async function create(e: FormEvent) {
    e.preventDefault()
    setErr('')
    try {
      await api('/api/hr/jobs', {
        method: 'POST',
        body: JSON.stringify({ title, description: desc }),
      })
      setTitle('')
      setDesc('')
      await load()
    } catch (ex) {
      setErr(ex instanceof Error ? ex.message : '失败')
    }
  }

  async function archive(id: number) {
    await api(`/api/hr/jobs/${id}/archive`, { method: 'POST' })
    await load()
  }

  function startEdit(j: Job) {
    setEditingId(j.id)
    setEditTitle(j.title)
    setEditDesc(j.description)
  }

  function cancelEdit() {
    setEditingId(null)
    setEditTitle('')
    setEditDesc('')
  }

  async function saveEdit(e: FormEvent) {
    e.preventDefault()
    if (editingId == null) return
    setErr('')
    try {
      await api(`/api/hr/jobs/${editingId}`, {
        method: 'PATCH',
        body: JSON.stringify({ title: editTitle, description: editDesc }),
      })
      cancelEdit()
      await load()
    } catch (ex) {
      setErr(ex instanceof Error ? ex.message : '保存失败')
    }
  }

  const visibleJobs = jobs.filter((j) =>
    tab === 'active' ? j.status === 'active' : j.status !== 'active'
  )
  const pageCount = Math.max(1, Math.ceil(visibleJobs.length / PAGE_SIZE))
  const pageJobs = visibleJobs.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE)

  function switchTab(nextTab: JobTab) {
    setTab(nextTab)
    setPage(1)
  }

  return (
    <div className="page-shell-inner">
      <form onSubmit={create} className="job-publish-form">
        <input
          placeholder="岗位标题"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          aria-label="岗位标题"
        />
        <input
          placeholder="岗位描述"
          value={desc}
          onChange={(e) => setDesc(e.target.value)}
          aria-label="岗位描述"
        />
        <button type="submit">发布</button>
      </form>

      {err && <p className="err">{err}</p>}

      <section id="job-list" className="job-list-section">
        <div className="tabs-row tabs-row--plain">
          <button
            type="button"
            className={`tab-button ${tab === 'active' ? 'tab-button--active' : ''}`}
            onClick={() => switchTab('active')}
          >
            已发布岗位 ({jobs.filter((j) => j.status === 'active').length})
          </button>
          <button
            type="button"
            className={`tab-button ${tab === 'archived' ? 'tab-button--active' : ''}`}
            onClick={() => switchTab('archived')}
          >
            已下架岗位 ({jobs.filter((j) => j.status !== 'active').length})
          </button>
        </div>
        <ul className="list job-list">
          {pageJobs.map((j) => (
            <li key={j.id} className="job-card">
              {editingId === j.id ? (
                <form onSubmit={saveEdit} className="stack-gap">
                  <label>
                    标题
                    <input
                      value={editTitle}
                      onChange={(e) => setEditTitle(e.target.value)}
                    />
                  </label>
                  <label>
                    描述
                    <textarea
                      rows={4}
                      value={editDesc}
                      onChange={(e) => setEditDesc(e.target.value)}
                    />
                  </label>
                  <div className="row" style={{ justifyContent: 'flex-start' }}>
                    <button type="submit">保存</button>
                    <button
                      type="button"
                      className="btn-secondary"
                      onClick={cancelEdit}
                    >
                      取消
                    </button>
                  </div>
                </form>
              ) : (
                <>
                  <div className="job-card-main">
                    <div className="job-card-copy">
                      <h3>{j.title}</h3>
                      <p>{j.description || '-'}</p>
                      <div className="job-meta">发布时间：{formatDate(j.created_at)}</div>
                    </div>
                    <div className="job-card-actions">
                      <button type="button" onClick={() => startEdit(j)}>
                        编辑
                      </button>
                      <button
                        type="button"
                        className="btn-secondary"
                        onClick={() => archive(j.id)}
                      >
                        {tab === 'active' ? '下架' : '上架'}
                      </button>
                    </div>
                  </div>
                </>
              )}
            </li>
          ))}
          {pageJobs.length === 0 && <li className="empty-list">暂无岗位</li>}
        </ul>
        <div className="pagination-row">
          <span>共 {visibleJobs.length} 条</span>
          <div className="pager-controls">
            <button
              type="button"
              className="btn-secondary"
              disabled={page <= 1}
              onClick={() => setPage((p) => Math.max(1, p - 1))}
            >
              上一页
            </button>
            <span>{page} / {pageCount}</span>
            <button
              type="button"
              className="btn-secondary"
              disabled={page >= pageCount}
              onClick={() => setPage((p) => Math.min(pageCount, p + 1))}
            >
              下一页
            </button>
          </div>
        </div>
      </section>
    </div>
  )
}
