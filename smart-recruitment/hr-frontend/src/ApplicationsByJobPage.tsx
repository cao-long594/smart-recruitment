import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from './api'

const PAGE_SIZE = 10

type Job = {
  id: number
  title: string
  description: string
  status: string
  created_at?: string
}

type JobWithCount = Job & { applications_count: number }

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

export function ApplicationsByJobPage() {
  const [jobs, setJobs] = useState<JobWithCount[]>([])
  const [page, setPage] = useState(1)
  const [err, setErr] = useState('')

  useEffect(() => {
    api<{ jobs: Job[] }>('/api/hr/jobs?page=1&page_size=50')
      .then(async (r) => {
        const nextJobs = r.jobs || []
        const counts = await Promise.all(
          nextJobs.map((j) =>
            api<{ total?: number; rows?: unknown[] }>(
              `/api/hr/jobs/${j.id}/applications?page=1&page_size=1`
            )
              .then((apps) => apps.total ?? apps.rows?.length ?? 0)
              .catch(() => 0)
          )
        )
        setJobs(nextJobs.map((j, i) => ({ ...j, applications_count: counts[i] })))
      })
      .catch((e) => setErr(String(e)))
  }, [])

  const pageCount = Math.max(1, Math.ceil(jobs.length / PAGE_SIZE))
  const pageJobs = jobs.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE)

  return (
    <div className="page-shell-inner">
      <div className="page-toolbar">
        <h1 className="page-title">按岗位查看投递</h1>
      </div>

      {err && <p className="err">{err}</p>}

      <div className="ledger-table-wrap">
        <table className="ledger-table">
          <thead>
            <tr>
              <th>岗位</th>
              <th>发布时间</th>
              <th>投递人数</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {pageJobs.map((j) => (
              <tr key={j.id}>
                <td>
                  <strong>{j.title}</strong>
                </td>
                <td>{formatDate(j.created_at)}</td>
                <td>{j.applications_count}</td>
                <td>
                  <span className={`status-pill status-pill--${j.status}`}>
                    {j.status === 'active' ? '招聘中' : '已下架'}
                  </span>
                </td>
                <td>
                  <Link className="table-action" to={`/jobs/${j.id}`}>
                    查看投递
                  </Link>
                </td>
              </tr>
            ))}
            {pageJobs.length === 0 && (
              <tr>
                <td colSpan={5} className="empty-cell">
                  暂无岗位
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
      <div className="pagination-row">
        <span>共 {jobs.length} 条</span>
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
    </div>
  )
}
