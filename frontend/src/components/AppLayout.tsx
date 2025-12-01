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
        <div className="logo">Image Manager</div>
        <nav className="nav-links">
          <NavLink to="/images">图片库</NavLink>
          <NavLink to="/upload">上传</NavLink>
          <NavLink to="/tags">标签</NavLink>
        </nav>
        <div className="user-info">
          <span>{user?.username}</span>
          <button onClick={handleLogout}>退出</button>
        </div>
      </header>
      <main className="app-main">
        <Outlet />
      </main>
    </div>
  )
}

export default AppLayout

