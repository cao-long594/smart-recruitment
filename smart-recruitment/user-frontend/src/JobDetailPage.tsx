import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { api, getToken } from './api'

type Job = {
  id: number
  title: string
  description: string
  status: string
}

export function JobDetailPage() {
  const { id } = useParams()
  const nav = useNavigate()
  const jobId = Number(id)
  const invalidJobId = !id || Number.isNaN(jobId)
  const [jobResult, setJobResult] = useState<{ id: number; job: Job | null } | null>(null)
  const [err, setErr] = useState('')
  const authed = !!getToken()
  const loading = !invalidJobId && jobResult?.id !== jobId
  const job = jobResult?.id === jobId ? jobResult.job : null

  useEffect(() => {
    if (invalidJobId) {
      return
    }
    let ignore = false
    api<{ jobs: Job[] }>('/api/jobs?page=1&page_size=100')
      .then((r) => {
        if (ignore) return
        const found = (r.jobs || []).find((j) => j.id === jobId)
        setJobResult({ id: jobId, job: found ?? null })
      })
      .catch((e) => {
        if (ignore) return
        setErr(String(e))
        setJobResult({ id: jobId, job: null })
      })
    return () => {
      ignore = true
    }
  }, [invalidJobId, jobId])

  async function apply() {
    if (!job) return
    setErr('')
    try {
      await api(`/api/jobs/${job.id}/apply`, { method: 'POST' })
      alert('投递成功')
    } catch (ex) {
      const msg = ex instanceof Error ? ex.message : '投递失败'
      setErr(
        msg ||
          '投递失败。请确认已补全姓名、电话、学历、院校、经历、技能标签，并上传 PDF、DOC 或 DOCX 简历。',
      )
    }
  }

  if (!invalidJobId && loading) {
    return (
      <div className="page-shell-inner">
        <div className="content-card">加载中…</div>
      </div>
    )
  }

  if (!job) {
    return (
      <div className="page-shell-inner">
        <div className="content-card">
          <p>未找到该岗位或岗位已下线。</p>
          <p>
            <Link to="/jobs">返回岗位列表</Link>
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="page-shell-inner">
      <div className="content-card detail-card">
        <p className="backline">
          <Link to="/jobs">返回岗位列表</Link>
        </p>
        <div className="detail-heading">
          <div>
            <p className="eyebrow">岗位详情</p>
            <h1 className="page-title">{job.title}</h1>
          </div>
          <span className="status-pill">{job.status}</span>
        </div>
        <div className="job-detail-body">
          <p>{job.description}</p>
        </div>
        <div className="apply-rule-box">
          <strong>投递规则</strong>
          <span>
            仅候选人账号登录后可投递；后端会校验结构化档案完整性和简历格式。
          </span>
        </div>
        <div className="row job-detail-actions">
          {authed ? (
            <button type="button" onClick={() => void apply()}>
              投递岗位
            </button>
          ) : (
            <button type="button" onClick={() => nav('/login', { state: { from: `/jobs/${job.id}` } })}>
              登录后投递
            </button>
          )}
        </div>
        {err && <p className="err err-block">{err}</p>}
      </div>
    </div>
  )
}
