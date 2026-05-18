import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api, setToken, setUsername as saveUsername } from './api'

export function LoginPage() {
  const nav = useNavigate()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [err, setErr] = useState('')

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setErr('')
    const account = username.trim()
    try {
      const r = await api<{ token: string; role: string }>('/api/login', {
        method: 'POST',
        body: JSON.stringify({ email: account, password }),
      })
      if (r.role !== 'hr') {
        setErr('请使用 HR 账号登录本端')
        return
      }
      setToken(r.token)
      saveUsername(account)
      nav('/')
    } catch (ex) {
      setErr(ex instanceof Error ? ex.message : '登录失败')
    }
  }

  return (
    <div className="auth-page">
      <div className="auth-card-outer">
        <aside className="auth-aside">
          <h1>智能招聘 HR 端</h1>
          <p>
            管理在招岗位、查看投递台账并与 AI
            助手协同分析招聘数据。请使用 HR 账号登录。
          </p>
        </aside>
        <div className="auth-main">
          <h2>登录</h2>
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
                autoComplete="current-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </label>
            <button type="submit">登录</button>
          </form>
          {err && <p className="err">{err}</p>}
          <p className="auth-footer muted">
            <Link to="/register">注册 HR 账号</Link>
          </p>
        </div>
      </div>
    </div>
  )
}
