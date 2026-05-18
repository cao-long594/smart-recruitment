import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api, setToken, setUsername as saveUsername } from './api'

export function RegisterPage() {
  const nav = useNavigate()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [err, setErr] = useState('')

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setErr('')
    const account = username.trim()
    if (!account) {
      setErr('用户名不能为空')
      return
    }
    try {
      await api('/api/register', {
        method: 'POST',
        body: JSON.stringify({ email: account, password, role: 'hr' }),
      })
      const r = await api<{ token: string }>('/api/login', {
        method: 'POST',
        body: JSON.stringify({ email: account, password }),
      })
      setToken(r.token)
      saveUsername(account)
      nav('/')
    } catch (ex) {
      setErr(ex instanceof Error ? ex.message : '注册失败')
    }
  }

  return (
    <div className="auth-page">
      <div className="auth-card-outer">
        <aside className="auth-aside">
          <h1>注册 HR 管理端</h1>
          <p>创建账号后即可发布岗位、处理投递并使用 AI 助手。</p>
        </aside>
        <div className="auth-main">
          <h2>注册</h2>
          <form onSubmit={onSubmit} className="form">
            <label>
              用户名
              <input
                type="text"
                autoComplete="username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                required
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
