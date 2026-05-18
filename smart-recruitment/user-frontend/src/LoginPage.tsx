import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { api, setToken } from './api'

export function LoginPage() {
  const nav = useNavigate()
  const loc = useLocation()
  const state = loc.state as { from?: string } | undefined
  const from =
    state?.from && state.from.startsWith('/') ? state.from : '/jobs'

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [err, setErr] = useState('')

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setErr('')
    try {
      const r = await api<{ token: string; role: string }>('/api/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      })
      if (r.role !== 'candidate') {
        setErr('请使用候选人账号登录本端')
        return
      }
      setToken(r.token)
      nav(from, { replace: true })
    } catch (ex) {
      setErr(ex instanceof Error ? ex.message : '登录失败')
    }
  }

  return (
    <div className="auth-page">
      <div className="auth-card-outer">
        <aside className="auth-aside">
          <p className="eyebrow">候选人账号</p>
          <h1>候选人端登录</h1>
          <p>登录后可维护结构化档案、上传私有 OSS 简历，并在岗位详情页发起投递。</p>
        </aside>
        <div className="auth-main">
          <h2>登录</h2>
          <form onSubmit={onSubmit} className="form">
            <label>
              邮箱
              <input
                type="email"
                autoComplete="username"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
            </label>
            <label>
              密码
              <input
                type="password"
                autoComplete="current-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </label>
            <button type="submit">登录</button>
          </form>
          {err && <p className="err">{err}</p>}
          <p className="auth-footer muted">
            <Link to="/register">注册账号</Link>
            {' · '}
            <Link to="/jobs">浏览岗位</Link>
          </p>
        </div>
      </div>
    </div>
  )
}
