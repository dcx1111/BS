import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import ImageListPage from './pages/ImageListPage'
import UploadPage from './pages/UploadPage'
import ImageDetailPage from './pages/ImageDetailPage'
import TagManagementPage from './pages/TagManagementPage'
import MCPSearchPage from './pages/MCPSearchPage'
import SlideshowPage from './pages/SlideshowPage'
import SlideshowEditPage from './pages/SlideshowEditPage'
import ProtectedRoute from './components/ProtectedRoute'
import AppLayout from './components/AppLayout'
import './App.css'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Navigate to="/images" />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />

        <Route element={<ProtectedRoute />}>
          <Route path="/slideshow" element={<SlideshowPage />} />
          <Route path="/slideshow/edit" element={<SlideshowEditPage />} />
          <Route element={<AppLayout />}>
            <Route path="/images" element={<ImageListPage />} />
            <Route path="/images/:id" element={<ImageDetailPage />} />
            <Route path="/upload" element={<UploadPage />} />
            <Route path="/tags" element={<TagManagementPage />} />
            <Route path="/mcp" element={<MCPSearchPage />} />
          </Route>
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App
