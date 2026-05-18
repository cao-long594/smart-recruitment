import { useEffect, useState } from 'react'
import type { FormEvent, KeyboardEvent } from 'react'
import { api } from './api'

type Profile = {
  user_id: number
  name: string
  phone: string
  education: string
  school: string
  experience: string
  skills: string
  profile_complete?: boolean
  has_resume?: boolean
}

function skillsToTags(s: string): string[] {
  return s
    .split(/[,，]/)
    .map((x) => x.trim())
    .filter(Boolean)
}

function isAllowedResume(file: File) {
  return /\.(pdf|doc|docx)$/i.test(file.name)
}

export function ProfilePage() {
  const [p, setP] = useState<Profile | null>(null)
  const [err, setErr] = useState('')
  const [file, setFile] = useState<File | null>(null)
  const [tags, setTags] = useState<string[]>([])
  const [tagInput, setTagInput] = useState('')
  const [parsing, setParsing] = useState(false)

  async function load() {
    const r = await api<{ profile: Profile }>('/api/me/profile')
    const prof = r.profile || null
    setP(prof)
    if (prof) setTags(skillsToTags(prof.skills))
  }

  useEffect(() => {
    let ignore = false
    api<{ profile: Profile }>('/api/me/profile')
      .then((r) => {
        if (ignore) return
        const prof = r.profile || null
        setP(prof)
        if (prof) setTags(skillsToTags(prof.skills))
      })
      .catch((e) => {
        if (!ignore) setErr(String(e))
      })
    return () => {
      ignore = true
    }
  }, [])

  function addTag(raw: string) {
    const t = raw.trim()
    if (!t) return
    if (tags.includes(t)) {
      setTagInput('')
      return
    }
    setTags((prev) => [...prev, t])
    setTagInput('')
  }

  function removeTag(t: string) {
    setTags((prev) => prev.filter((x) => x !== t))
  }

  function onTagKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter') {
      e.preventDefault()
      addTag(tagInput)
    }
  }

  async function save(e: FormEvent) {
    e.preventDefault()
    if (!p) return
    setErr('')
    try {
      await api('/api/me/profile', {
        method: 'PUT',
        body: JSON.stringify({
          name: p.name,
          phone: p.phone,
          education: p.education,
          school: p.school,
          experience: p.experience,
          skills: tags.join(','),
        }),
      })
      await load()
      alert('已保存')
    } catch (ex) {
      setErr(ex instanceof Error ? ex.message : '保存失败')
    }
  }

  async function uploadResume() {
    if (!file) return
    setErr('')
    if (!isAllowedResume(file)) {
      setErr('仅支持上传 PDF、DOC、DOCX 三种标准办公格式。')
      return
    }
    const name = file.name
    const ct =
      file.type ||
      (name.endsWith('.pdf')
        ? 'application/pdf'
        : name.endsWith('.docx')
          ? 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
          : 'application/msword')
    try {
      const cred = await api<{
        upload_url: string
        object_key: string
        headers: Record<string, string>
      }>('/api/me/resume/upload_url', {
        method: 'POST',
        body: JSON.stringify({ file_name: name, content_type: ct }),
      })
      const h = new Headers()
      Object.entries(cred.headers || {}).forEach(([k, v]) => {
        const key = k.toLowerCase()
        if (key === 'host') return
        h.set(k, v)
      })
      const buf = await file.arrayBuffer()
      const put = await fetch(cred.upload_url, {
        method: 'PUT',
        headers: h,
        body: buf,
      })
      if (!put.ok) {
        throw new Error('上传到 OSS 失败 ' + put.status)
      }
      await api('/api/me/resume/confirm', {
        method: 'POST',
        body: JSON.stringify({
          object_key: cred.object_key,
          file_name: name,
          content_type: ct,
          size_bytes: file.size,
        }),
      })
      setFile(null)
      await load()
      alert('简历已上传')
    } catch (ex) {
      const msg = ex instanceof Error ? ex.message : '上传失败'
      if (msg === 'Failed to fetch') {
        setErr(
          '无法连接上传地址。请检查：1）OSS 跨域已勾选 PUT、GET、HEAD；2）来源与地址栏一致（localhost 与 127.0.0.1 需分别添加）；3）允许 Headers 建议 *；4）保存后等待配置生效。',
        )
      } else {
        setErr(msg)
      }
    }
  }

  async function parseResume() {
    if (!p?.has_resume || parsing) return
    setErr('')
    setParsing(true)
    try {
      const r = await api<{ profile: Partial<Profile> }>('/api/me/resume/parse', {
        method: 'POST',
      })
      const parsed = r.profile || {}
      setP({
        ...p,
        name: parsed.name ?? p.name,
        phone: parsed.phone ?? p.phone,
        education: parsed.education ?? p.education,
        school: parsed.school ?? p.school,
        experience: parsed.experience ?? p.experience,
        skills: parsed.skills ?? p.skills,
      })
      setTags(skillsToTags(parsed.skills || p.skills || ''))
    } catch (ex) {
      const msg = ex instanceof Error ? ex.message : '解析失败'
      if (msg.includes('docx only')) {
        setErr('当前自动解析仅支持 DOCX 简历。请上传 DOCX 后再解析，PDF/DOC 仍可作为投递附件使用。')
      } else if (msg.includes('upload resume first')) {
        setErr('当前账号还没有已确认上传的简历，请先上传简历并等待上传成功后再解析。')
      } else if (msg.includes('cannot read resume')) {
        setErr('后端无法从 OSS 读取当前简历。请检查 OSS 配置、object_key 是否存在，或重新选择本地 DOCX 文件后点击解析。')
      } else if (msg.includes('cannot parse docx') || msg.includes('no readable text')) {
        setErr('当前 DOCX 没有可读取文本，可能是图片版简历或文件结构异常。请换成可复制文本的 DOCX 简历。')
      } else {
        setErr(msg)
      }
    } finally {
      setParsing(false)
    }
  }

  if (!p) {
    return (
      <div className="page-shell-inner">
        <div className="content-card">加载中…</div>
      </div>
    )
  }

  return (
    <div className="page-shell-inner">
      <section className="page-hero profile-hero">
        <div>
          <p className="eyebrow">投递资料</p>
          <h1 className="page-title">个人档案与简历</h1>
          <p className="page-lead">
            这些信息会实时保存到后端，用于 HR 查看投递台账和筛选候选人。
          </p>
        </div>
        <div className="profile-checks" aria-label="资料状态">
          <span className={p.profile_complete ? 'check-pill is-ok' : 'check-pill'}>
            档案{p.profile_complete ? '已完整' : '待补全'}
          </span>
          <span className={p.has_resume ? 'check-pill is-ok' : 'check-pill'}>
            简历{p.has_resume ? '已上传' : '未上传'}
          </span>
        </div>
      </section>

      <div className="content-card">
        {p.profile_complete === false && (
          <p className="err">请补全姓名、联系电话、最高学历、毕业院校、经历和技能标签后再投递岗位。</p>
        )}
        <form onSubmit={save} className="form profile-form-wrap">
          <div className="profile-two-col">
            <div className="profile-col">
              <label>
                姓名 <span className="required-mark">必填</span>
                <input
                  value={p.name}
                  onChange={(e) => setP({ ...p, name: e.target.value })}
                />
              </label>
              <label>
                联系电话 <span className="required-mark">必填</span>
                <input
                  type="tel"
                  value={p.phone}
                  onChange={(e) => setP({ ...p, phone: e.target.value })}
                />
              </label>
              <label>
                最高学历 <span className="required-mark">必填</span>
                <input
                  value={p.education}
                  onChange={(e) => setP({ ...p, education: e.target.value })}
                />
              </label>
              <label>
                毕业院校 <span className="required-mark">必填</span>
                <input
                  value={p.school}
                  onChange={(e) => setP({ ...p, school: e.target.value })}
                />
              </label>
            </div>
            <div className="profile-col">
              <label>
                工作/项目经历 <span className="required-mark">必填</span>
                <textarea
                  rows={10}
                  value={p.experience}
                  onChange={(e) => setP({ ...p, experience: e.target.value })}
                />
              </label>
              <div className="tag-field">
                <span className="labelish">核心技能标签 <span className="required-mark">必填</span></span>
                <div className="tag-list" aria-label="已添加的技能标签">
                  {tags.map((t) => (
                    <span key={t} className="tag-chip">
                      {t}
                      <button
                        type="button"
                        onClick={() => removeTag(t)}
                        aria-label={`移除标签 ${t}`}
                      >
                        ×
                      </button>
                    </span>
                  ))}
                </div>
                <div className="tag-input-row">
                  <input
                    value={tagInput}
                    onChange={(e) => setTagInput(e.target.value)}
                    onKeyDown={onTagKeyDown}
                    placeholder="输入后按回车添加"
                    aria-label="添加技能标签，回车确认"
                  />
                  <button type="button" className="btn-secondary" onClick={() => addTag(tagInput)}>
                    添加
                  </button>
                </div>
              </div>
            </div>
          </div>
          <div className="profile-actions">
            <button type="submit">保存档案</button>
          </div>
        </form>

        <div className="resume-card">
          <h2 className="resume-heading">我的简历</h2>
          <p className="muted">
            {p.has_resume
              ? '已上传合规简历。重新上传后会覆盖当前简历记录。'
              : '未上传简历，投递前必须上传 PDF、DOC 或 DOCX。'}
          </p>
          <div className="resume-upload-row row">
            <input
              id="resume-file"
              type="file"
              accept=".pdf,.doc,.docx,application/pdf,application/msword,application/vnd.openxmlformats-officedocument.wordprocessingml.document"
              onChange={(e) => setFile(e.target.files?.[0] || null)}
            />
            <span className="muted resume-file-name">
              {file ? file.name : '未选择文件'}
            </span>
            <button type="button" className="btn-secondary" onClick={uploadResume} disabled={!file}>
              重新上传
            </button>
            <button
              type="button"
              className="btn-secondary"
              onClick={() => void parseResume()}
              disabled={!p.has_resume || parsing}
              title={p.has_resume ? '解析当前已上传简历并填入上方表单' : '请先上传简历'}
            >
              {parsing ? '解析中…' : '解析并填充'}
            </button>
          </div>
        </div>

        {err && <p className="err" style={{ marginTop: '0.75rem' }}>{err}</p>}
      </div>
    </div>
  )
}
