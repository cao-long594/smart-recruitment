import { useEffect, useState } from 'react'
import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { getToken, setToken } from './api'

const SIDEBAR_KEY = 'candidate_sidebar_collapsed'

export function CandidateAppShell() {
  const nav = useNavigate()
  const authed = !!getToken()
  const [collapsed, setCollapsed] = useState(() => {
    try {
      return localStorage.getItem(SIDEBAR_KEY) === '1'
    } catch {
      return false
    }
  })

  useEffect(() => {
    try {
      localStorage.setItem(SIDEBAR_KEY, collapsed ? '1' : '0')
    } catch {
      /* ignore */
    }
  }, [collapsed])

  function logout() {
    setToken(null)
    nav('/jobs')
    window.location.reload()
  }

  const shellCls = `app-shell${collapsed ? ' app-shell--sidebar-collapsed' : ''}`

  return (
    <div className={shellCls}>
      <header className="topbar">
        <div className="topbar-brand">
          <span className="topbar-brand-mark" aria-hidden />
          <span className="topbar-brand-text">招聘候选人端</span>
        </div>
        <div className="topbar-actions">
          {authed ? (
            <button type="button" className="topbar-linkish" onClick={logout}>
              退出
            </button>
          ) : (
            <>
              <NavLink to="/login" className="topbar-navlink">
                登录
              </NavLink>
              <NavLink to="/register" className="topbar-navlink">
                注册
              </NavLink>
            </>
          )}
        </div>
      </header>
      <div className="app-body">
        <aside id="candidate-sidebar" className="sidebar" aria-label="主导航">
          <button
            type="button"
            className="sidebar-toggle"
            onClick={() => setCollapsed((c) => !c)}
            aria-expanded={!collapsed}
            aria-controls="candidate-sidebar"
          >
            <span aria-hidden>{collapsed ? '→' : '←'}</span>
            {collapsed ? '展开导航' : '收起导航'}
          </button>
          <nav className="sidebar-nav">
            <div className="sidebar-group">
              <div className="sidebar-group-title">公开岗位</div>
              <NavLink className="sidebar-link" to="/jobs">
                浏览全部岗位
              </NavLink>
              <p className="sidebar-hint">
                游客可直接浏览所有公开在招岗位。
              </p>
              {authed && (
                <NavLink className="sidebar-link" to="/my-applications">
                  已投递与状态
                </NavLink>
              )}
            </div>
            <div className="sidebar-group">
              <div className="sidebar-group-title">投递准备</div>
              {authed ? (
                <NavLink className="sidebar-link" to="/profile">
                  个人档案与简历
                </NavLink>
              ) : (
                <>
                  <NavLink className="sidebar-link" to="/login">
                    登录候选人账号
                  </NavLink>
                  <NavLink className="sidebar-link" to="/register">
                    注册后投递
                  </NavLink>
                </>
              )}
            </div>
          </nav>
        </aside>
        <main id="candidate-main" className="main-content">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
