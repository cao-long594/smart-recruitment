import { useEffect, useState } from 'react'
import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { getUsername, setToken } from './api'

const SIDEBAR_KEY = 'hr_sidebar_collapsed'

export function HrAppShell() {
  const nav = useNavigate()
  const [userMenuOpen, setUserMenuOpen] = useState(false)
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
    nav('/login')
  }

  const shellCls = `app-shell${collapsed ? ' app-shell--sidebar-collapsed' : ''}`

  return (
    <div className={shellCls}>
      <header className="topbar">
        <div className="topbar-brand">
          <span className="topbar-brand-mark" aria-hidden>
            |
          </span>
          <span className="topbar-brand-text">HR 管理端</span>
        </div>
        <div className="topbar-actions">
          <button
            type="button"
            className="topbar-user-button"
            onClick={() => setUserMenuOpen((open) => !open)}
            aria-haspopup="menu"
            aria-expanded={userMenuOpen}
          >
            {getUsername()}
          </button>
          {userMenuOpen && (
            <div className="topbar-user-menu" role="menu">
              <button type="button" role="menuitem" onClick={logout}>
                退出
              </button>
            </div>
          )}
        </div>
      </header>
      <div className="app-body">
        {collapsed && (
          <button
            type="button"
            className="sidebar-restore"
            onClick={() => setCollapsed(false)}
            aria-label="展开导航"
            title="展开导航"
          >
            <span className="hamburger-icon" aria-hidden>
              <span />
              <span />
              <span />
            </span>
          </button>
        )}
        <aside
          id="hr-sidebar"
          className="sidebar"
          aria-label="主导航"
        >
          <button
            type="button"
            className="sidebar-toggle"
            onClick={() => setCollapsed((c) => !c)}
            aria-expanded={!collapsed}
            aria-controls="hr-sidebar"
          >
            <span className="collapse-mark" aria-hidden>
              <span />
              <span />
            </span>
            收起导航
          </button>
          <nav className="sidebar-nav">
            <NavLink className="sidebar-link" end to="/">
              发布与管理岗位
            </NavLink>
            <NavLink className="sidebar-link" to="/applications">
              按岗位查看投递
            </NavLink>
            <NavLink className="sidebar-link" to="/ai">
              打开 AI 助手
            </NavLink>
          </nav>
        </aside>
        <main id="hr-main" className="main-content">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
