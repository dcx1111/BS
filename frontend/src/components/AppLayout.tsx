import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import './AppLayout.css'

const AppLayout = () => {
  const navigate = useNavigate()
  const { logout, user } = useAuthStore()

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  return (
    <div className="app-shell">
      <header className="app-header">
        <div className="logo logo-desktop">Image Manager</div>
        <div className="header-top-mobile">
          <div className="logo logo-mobile">Image Manager</div>
          <div className="user-section-mobile">
            <div className="user-info">
              <span className="username">{user?.username}</span>
            </div>
          </div>
        </div>
        <nav className="nav-links nav-links-desktop">
          <NavLink to="/images">图片库</NavLink>
          <NavLink to="/upload">上传</NavLink>
          <NavLink to="/tags">标签</NavLink>
          <NavLink to="/mcp">AI搜索</NavLink>
          <NavLink to="/slideshow">轮播</NavLink>
        </nav>
        <div className="user-section user-section-desktop">
          <div className="user-info">
            <span className="username">{user?.username}</span>
          </div>
          <button onClick={handleLogout} className="logout-btn-desktop">退出</button>
        </div>
        <div className="header-bottom">
          <nav className="nav-links nav-links-mobile">
            <NavLink to="/images">图片库</NavLink>
            <NavLink to="/upload">上传</NavLink>
            <NavLink to="/tags">标签</NavLink>
            <NavLink to="/mcp">AI搜索</NavLink>
            <NavLink to="/slideshow">轮播</NavLink>
          </nav>
          <button onClick={handleLogout} className="logout-btn logout-btn-mobile">退出</button>
        </div>
      </header>
      <main className="app-main">
        <Outlet />
      </main>
    </div>
  )
}

export default AppLayout

