import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, getToken } from './api'

type Job = {
  id: number
  title: string
  description: string
  status: string
}

function excerpt(text: string, max = 120) {
  const t = text.replace(/\s+/g, ' ').trim()
  if (t.length <= max) return t
  return `${t.slice(0, max)}…`
}

export function JobsPage() {
  const [jobs, setJobs] = useState<Job[]>([])
  const [total, setTotal] = useState(0)
  const [err, setErr] = useState('')
  const authed = !!getToken()

  useEffect(() => {
    api<{ jobs: Job[]; total?: number }>('/api/jobs?page=1&page_size=50')
      .then((r) => {
        setJobs(r.jobs || [])
        setTotal(r.total ?? (r.jobs || []).length)
      })
      .catch((e) => setErr(String(e)))
  }, [])

  return (
    <div className="page-shell-inner">
      <section className="page-hero">
        <div>
          <p className="eyebrow">公开岗位大厅</p>
          <h1 className="page-title">浏览在招岗位</h1>
          <p className="page-lead">
            {authed
              ? '你已登录候选人账号，进入岗位详情后可直接发起投递。'
              : '无需登录即可查看全部公开岗位，注册或登录后再投递。'}
          </p>
        </div>
        <div className="hero-stat" aria-label="在招岗位数量">
          <strong>{total}</strong>
          <span>在招岗位</span>
        </div>
      </section>

      <div className="notice-strip">
        <span>投递前会由后端校验结构化档案和合规简历。</span>
        {!authed && <Link to="/register">注册候选人账号</Link>}
      </div>

      <div className="content-card job-list-card">
        {err && <p className="err">{err}</p>}
        {!err && jobs.length === 0 && (
          <p className="muted empty-state">暂无公开在招岗位。</p>
        )}
        <ul className="list job-list">
          {jobs.map((j) => (
            <li key={j.id} className="card job-card">
              <div className="job-card-top">
                <div>
                  <h2>{j.title}</h2>
                  <p>{excerpt(j.description)}</p>
                </div>
                <span className="status-pill">{j.status}</span>
              </div>
              <div className="job-card-actions">
                <span className="muted">
                  {authed ? '可进入详情页投递' : '登录后开放投递按钮'}
                </span>
                <Link className="action-link" to={`/jobs/${j.id}`}>
                  查看详情
                </Link>
              </div>
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}
