import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api, setToken } from './api'

export function RegisterPage() {
  const nav = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [err, setErr] = useState('')

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setErr('')
    try {
      await api('/api/register', {
        method: 'POST',
        body: JSON.stringify({ email, password, role: 'candidate' }),
      })
      const r = await api<{ token: string }>('/api/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      })
      setToken(r.token)
      nav('/jobs', { replace: true })
    } catch (ex) {
      setErr(ex instanceof Error ? ex.message : '注册失败')
    }
  }

  return (
    <div className="auth-page">
      <div className="auth-card-outer">
        <aside className="auth-aside">
          <p className="eyebrow">候选人账号</p>
          <h1>注册候选人账号</h1>
          <p>创建候选人账号后再投递岗位；HR 账号请使用独立 HR 管理端。</p>
        </aside>
        <div className="auth-main">
          <h2>注册</h2>
          <form onSubmit={onSubmit} className="form">
            <label>
              邮箱
              <input
                type="email"
                autoComplete="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
            </label>
            <label>
              密码
              <input
                type="password"
                autoComplete="new-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </label>
            <button type="submit">注册并登录</button>
          </form>
          {err && <p className="err">{err}</p>}
          <p className="auth-footer muted">
            <Link to="/login">已有账号？返回登录</Link>
          </p>
        </div>
      </div>
    </div>
  )
}
