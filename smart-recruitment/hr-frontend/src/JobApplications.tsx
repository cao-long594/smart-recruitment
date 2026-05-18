import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api } from './api'

const PAGE_SIZE = 10

type Candidate = {
  name: string
  phone: string
  education: string
  school?: string
  experience?: string
  skills?: string
  has_resume?: boolean
}

type Job = {
  id: number
  title: string
}

type Row = {
  id: number
  candidate_user_id: number
  applied_at?: string
  candidate?: Candidate
}

export function JobApplications() {
  const { id } = useParams()
  const [rows, setRows] = useState<Row[]>([])
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [jobTitle, setJobTitle] = useState('')
  const [page, setPage] = useState(1)
  const [err, setErr] = useState('')

  useEffect(() => {
    if (!id) return
    api<{ rows: Row[] }>(`/api/hr/jobs/${id}/applications?page=1&page_size=50`)
      .then((r) => {
        const nextRows = r.rows || []
        setRows(nextRows)
        setSelectedId(nextRows[0]?.id ?? null)
        setPage(1)
      })
      .catch((e) => setErr(String(e)))
    api<{ jobs: Job[] }>('/api/hr/jobs?page=1&page_size=50')
      .then((r) => {
        const job = (r.jobs || []).find((item) => String(item.id) === id)
        setJobTitle(job?.title || '岗位')
      })
      .catch(() => setJobTitle('岗位'))
  }, [id])

  async function downloadResume(candidateUserId: number) {
    const r = await api<{ download_url: string }>(
      `/api/hr/resume_download?candidate_user_id=${candidateUserId}`
    )
    window.open(r.download_url, '_blank')
  }

  const selected = rows.find((r) => r.id === selectedId) || rows[0]
  const pageCount = Math.max(1, Math.ceil(rows.length / PAGE_SIZE))
  const pageRows = rows.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE)

  function formatDay(value?: string) {
    if (!value) return '-'
    const date = new Date(value)
    if (Number.isNaN(date.getTime())) return value.slice(0, 10)
    return date.toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
    })
  }

  return (
    <div className="page-shell-inner">
      <div className="back-row">
        <Link to="/applications">← 返回岗位列表</Link>
      </div>
      <h1 className="page-title">投递台账 - {jobTitle}</h1>

      {err && <p className="err">{err}</p>}

      <div className="applications-layout">
        <section className="applications-list-panel">
          <div className="tabs-row">
            <button type="button" className="tab-button tab-button--active">
              全部 ({rows.length})
            </button>
          </div>
          <div className="applications-table-wrap">
            <table className="ledger-table applications-table">
              <thead>
                <tr>
                  <th className="candidate-name-col">候选人</th>
                  <th className="candidate-phone-col">联系方式</th>
                  <th className="candidate-edu-col">最高学历</th>
                  <th className="candidate-date-col">投递时间</th>
                  <th className="candidate-action-col">操作</th>
                </tr>
              </thead>
              <tbody>
                {pageRows.map((r) => (
                  <tr key={r.id} className={selected?.id === r.id ? 'selected-row' : ''}>
                    <td className="candidate-name-col">{r.candidate?.name || `候选人 ${r.candidate_user_id}`}</td>
                    <td className="candidate-phone-col">{r.candidate?.phone || '-'}</td>
                    <td className="candidate-edu-col">{r.candidate?.education || '-'}</td>
                    <td className="candidate-date-col">{formatDay(r.applied_at)}</td>
                    <td className="candidate-action-col">
                      <button
                        type="button"
                        className="table-button"
                        onClick={() => setSelectedId(r.id)}
                      >
                        查看信息
                      </button>
                    </td>
                  </tr>
                ))}
                {pageRows.length === 0 && (
                  <tr>
                    <td colSpan={5} className="empty-cell">
                      暂无投递记录
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
          <div className="pagination-row pagination-row--inside">
            <span>共 {rows.length} 条</span>
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

        <aside className="candidate-detail-panel">
          {selected ? (
            <>
              <div className="candidate-summary">
                <div className="avatar" aria-hidden>
                  {selected.candidate?.name?.slice(0, 1) || '候'}
                </div>
                <div>
                  <div className="candidate-name-row">
                    <h2>{selected.candidate?.name || `候选人 ${selected.candidate_user_id}`}</h2>
                  </div>
                  <p>{selected.candidate?.phone || '-'}</p>
                </div>
              </div>

              <div className="detail-block">
                <h3 className="structured-title">结构化信息</h3>
                <dl className="structured-dl">
                  <div>
                    <dt>最高学历</dt>
                    <dd>{selected.candidate?.education || '-'}</dd>
                  </div>
                  <div>
                    <dt>毕业院校</dt>
                    <dd>{selected.candidate?.school || '-'}</dd>
                  </div>
                  <div className="structured-span">
                    <dt>工作/项目经历</dt>
                    <dd>{selected.candidate?.experience || '-'}</dd>
                  </div>
                  <div className="structured-span">
                    <dt>核心技能标签</dt>
                    <dd>{selected.candidate?.skills || '-'}</dd>
                  </div>
                </dl>
              </div>

              <div className="detail-block">
                <h3 className="structured-title">简历</h3>
                <div className="resume-row">
                  <span>{selected.candidate?.has_resume ? '候选人简历' : '暂无简历'}</span>
                  <button
                    type="button"
                    disabled={!selected.candidate?.has_resume}
                    onClick={() => downloadResume(selected.candidate_user_id)}
                  >
                    下载简历
                  </button>
                </div>
              </div>
            </>
          ) : (
            <div className="empty-detail">请选择候选人</div>
          )}
        </aside>
      </div>
    </div>
  )
}
